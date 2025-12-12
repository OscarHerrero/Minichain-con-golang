package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"minichain/core/rawdb"
	"minichain/core/state"
	"minichain/database"
	"minichain/database/leveldb"
	"minichain/evm"
	"time"
)

// Blockchain es la cadena completa de bloques
type Blockchain struct {
	Blocks       []*Block                 // Array de bloques (en memoria, para compatibilidad)
	Difficulty   int                      // Dificultad del minado (ej: 3 = "000...")
	AccountState *AccountState            // Estado de todas las cuentas (legacy)
	PendingTxs   []*Transaction           // Transacciones pendientes (mempool)
	Contracts    map[string]*evm.Contract // Contratos desplegados (legacy, ahora en StateDB)

	// Persistencia estilo Ethereum
	db      database.Database // Base de datos LevelDB
	stateDB *state.StateDB    // Estado mundial (cuentas + contratos)
}

// NewBlockchain crea una nueva blockchain con el bloque génesis (sin persistencia)
func NewBlockchain(difficulty int) *Blockchain {
	// Crear el bloque génesis (bloque #0)
	genesisBlock := NewGenesisBlock()

	// Minar el bloque génesis
	genesisBlock.MineBlock(difficulty)

	// Crear la blockchain
	bc := &Blockchain{
		Blocks:       []*Block{genesisBlock},
		Difficulty:   difficulty,
		AccountState: NewAccountState(),
		PendingTxs:   []*Transaction{},
		Contracts:    make(map[string]*evm.Contract),
	}

	return bc
}

// NewBlockchainWithDB crea una blockchain con persistencia estilo Ethereum
func NewBlockchainWithDB(difficulty int, dbPath string) (*Blockchain, error) {
	// Abrir base de datos LevelDB
	db, err := leveldb.New(dbPath, 16, 16, "", false)
	if err != nil {
		return nil, fmt.Errorf("error abriendo base de datos: %v", err)
	}

	// Intentar cargar el último bloque de la DB
	headHash, err := rawdb.ReadHeadBlockHash(db)
	var genesisBlock *Block
	var stateRoot []byte
	var blocks []*Block

	if err == nil && headHash != nil {
		// Ya existe una blockchain, cargar desde DB
		fmt.Println("📂 Cargando blockchain existente desde disco...")

		// Obtener el número del head block
		headNumber, err := rawdb.ReadHeaderNumber(db, headHash)
		if err != nil {
			return nil, fmt.Errorf("error obteniendo número del head block: %v", err)
		}

		// Cargar TODOS los bloques desde el génesis hasta el head
		fmt.Printf("📥 Cargando %d bloques desde disco...\n", headNumber+1)

		blocks = make([]*Block, 0, headNumber+1)

		// Cargar cada bloque en orden (0 hasta headNumber)
		for i := uint64(0); i <= headNumber; i++ {
			// Obtener hash del bloque en esta altura
			blockHash, err := rawdb.ReadCanonicalHash(db, i)
			if err != nil || blockHash == nil {
				return nil, fmt.Errorf("no se encontró hash canónico para bloque #%d: %v", i, err)
			}

			// Leer el bloque
			header, body, err := rawdb.ReadBlock(db, blockHash, i)
			if err != nil {
				return nil, fmt.Errorf("error cargando bloque #%d: %v", i, err)
			}

			// Convertir a nuestro formato
			block := headerToBlock(header, body)
			blocks = append(blocks, block)

			if i%100 == 0 || i == headNumber {
				fmt.Printf("   ✅ Cargados %d/%d bloques...\n", i+1, headNumber+1)
			}
		}

		genesisBlock = blocks[0]
		stateRoot = blocks[len(blocks)-1].StateRoot

		fmt.Printf("✅ Blockchain cargada: %d bloques (altura: %d)\n", len(blocks), headNumber)
	} else {
		// Nueva blockchain, crear génesis
		fmt.Println("🆕 Creando nueva blockchain con persistencia...")

		genesisBlock = NewGenesisBlock()

		// Crear StateDB vacío
		stateDatabase := state.NewDatabase(db)
		stateDB, err := state.New(nil, stateDatabase)
		if err != nil {
			return nil, fmt.Errorf("error creando StateDB: %v", err)
		}

		// Inicializar cuentas génesis si es necesario (opcional)
		// Por ejemplo, dar balance inicial a una cuenta

		// Calcular state root inicial
		genesisBlock.StateRoot, err = stateDB.Commit()
		if err != nil {
			return nil, fmt.Errorf("error calculando state root: %v", err)
		}

		// Minar el bloque génesis
		genesisBlock.MineBlock(difficulty)

		// Persistir bloque génesis
		if err := rawdb.WriteBlock(db, blockToHeader(genesisBlock), blockToBody(genesisBlock)); err != nil {
			return nil, fmt.Errorf("error persistiendo bloque génesis: %v", err)
		}

		// Marcar como head block (convertir hash hex a bytes)
		hashBytes, err := hex.DecodeString(genesisBlock.Hash)
		if err != nil {
			return nil, fmt.Errorf("error decodificando hash: %v", err)
		}
		// Escribir hash canónico para el génesis (altura 0 -> hash)
		rawdb.WriteCanonicalHash(db, hashBytes, 0)
		// Actualizar head block
		rawdb.WriteHeadBlockHash(db, hashBytes)

		stateRoot = genesisBlock.StateRoot
		blocks = []*Block{genesisBlock}
	}

	// Crear StateDB con el root del bloque génesis
	stateDatabase := state.NewDatabase(db)
	stateDB, err := state.New(stateRoot, stateDatabase)
	if err != nil {
		return nil, fmt.Errorf("error creando StateDB: %v", err)
	}

	// Crear la blockchain
	bc := &Blockchain{
		Blocks:       blocks,
		Difficulty:   difficulty,
		AccountState: NewAccountState(), // Mantener por compatibilidad
		PendingTxs:   []*Transaction{},
		Contracts:    make(map[string]*evm.Contract), // Mantener por compatibilidad
		db:           db,
		stateDB:      stateDB,
	}

	// Si cargamos desde disco, re-ejecutar transacciones para reconstruir AccountState
	if len(blocks) > 1 {
		fmt.Printf("💼 Re-ejecutando transacciones para reconstruir estado...\n")
		totalTxs := 0
		for i, block := range blocks {
			if i == 0 {
				continue // Saltar génesis
			}
			for _, tx := range block.Transactions {
				if err := tx.Execute(bc.AccountState, bc); err != nil {
					fmt.Printf("   ⚠️  Error re-ejecutando tx en bloque #%d: %v\n", i, err)
				}
				totalTxs++
			}
		}
		if totalTxs > 0 {
			fmt.Printf("✅ Estado reconstruido (%d transacciones procesadas)\n", totalTxs)
		}
	}

	fmt.Printf("✅ Blockchain inicializada (dificultad: %d)\n", difficulty)
	fmt.Printf("   State Root: %x\n", stateRoot[:16])

	return bc, nil
}

// AddTransaction añade una transacción al mempool (pendientes)
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	// Validar la transacción
	if err := tx.Validate(bc.AccountState, bc); err != nil {
		return err
	}

	// Añadir al mempool
	bc.PendingTxs = append(bc.PendingTxs, tx)

	fmt.Printf("✅ Transacción añadida al mempool (total: %d pendientes)\n", len(bc.PendingTxs))

	return nil
}

// MineBlock mina un nuevo bloque con las transacciones pendientes
func (bc *Blockchain) MineBlock() {
	if len(bc.PendingTxs) == 0 {
		fmt.Println("\n⚠️  No hay transacciones pendientes para minar")
		return
	}

	prevBlock := bc.Blocks[len(bc.Blocks)-1]

	// Crear nuevo bloque
	newBlock := &Block{
		Index:        len(bc.Blocks),
		Timestamp:    time.Now(),
		Transactions: bc.PendingTxs,
		PreviousHash: prevBlock.Hash,
		Nonce:        0,
	}

	// ====================================
	// FASE 1: EJECUTAR TRANSACCIONES
	// ====================================
	fmt.Println("\n💼 Ejecutando transacciones del bloque...")

	for i, tx := range bc.PendingTxs {
		fmt.Printf("\n📝 Transacción %d/%d:\n", i+1, len(bc.PendingTxs))

		// Mostrar tipo de transacción
		if tx.IsContractDeployment() {
			fmt.Println("   Tipo: DESPLIEGUE DE CONTRATO")
		} else if tx.IsContractCall(bc) {
			fmt.Println("   Tipo: LLAMADA A CONTRATO")
		} else {
			fmt.Printf("   Tipo: TRANSFERENCIA (%s → %s: %.2f MTC)\n",
				tx.From[:16]+"...", tx.To[:16]+"...", tx.Amount)
		}

		// Ejecutar en modo legacy (AccountState)
		if err := tx.Execute(bc.AccountState, bc); err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}

		// Si tenemos StateDB, actualizar también ahí
		if bc.stateDB != nil {
			// TODO: Sincronizar cambios de AccountState a StateDB
			// Por ahora, solo ejecutar en AccountState
		}

		if tx.Amount > 0 {
			fmt.Printf("   ✅ Fondos transferidos\n")
		}
	}

	// ====================================
	// FASE 2: CALCULAR MERKLE ROOTS
	// ====================================
	if bc.stateDB != nil {
		// Calcular State Root
		stateRoot, err := bc.stateDB.Commit()
		if err != nil {
			fmt.Printf("⚠️  Error calculando state root: %v\n", err)
			newBlock.StateRoot = make([]byte, 32)
		} else {
			newBlock.StateRoot = stateRoot
			fmt.Printf("   📊 State Root: %x...\n", stateRoot[:8])
		}

		// TODO: Calcular TxRoot y ReceiptRoot
		newBlock.TxRoot = make([]byte, 32)
		newBlock.ReceiptRoot = make([]byte, 32)
	} else {
		// Modo legacy sin persistencia
		newBlock.StateRoot = make([]byte, 32)
		newBlock.TxRoot = make([]byte, 32)
		newBlock.ReceiptRoot = make([]byte, 32)
	}

	// ====================================
	// FASE 3: MINAR EL BLOQUE (Proof of Work)
	// ====================================
	fmt.Printf("\n⛏️  Minando bloque %d (dificultad: %d, %d transacciones)...\n",
		newBlock.Index, bc.Difficulty, len(bc.PendingTxs))

	newBlock.MineBlock(bc.Difficulty)

	// ====================================
	// FASE 4: PERSISTIR EN BASE DE DATOS
	// ====================================
	if bc.db != nil {
		if err := rawdb.WriteBlock(bc.db, blockToHeader(newBlock), blockToBody(newBlock)); err != nil {
			fmt.Printf("⚠️  Error persistiendo bloque: %v\n", err)
		} else {
			// Convertir hash hex a bytes
			hashBytes, err := hex.DecodeString(newBlock.Hash)
			if err == nil {
				// Escribir hash canónico (altura -> hash)
				rawdb.WriteCanonicalHash(bc.db, hashBytes, uint64(newBlock.Index))
				// Actualizar head block
				rawdb.WriteHeadBlockHash(bc.db, hashBytes)
				fmt.Println("   💾 Bloque persistido en disco")
			}
		}
	}

	// ====================================
	// FASE 5: AÑADIR A CADENA EN MEMORIA
	// ====================================
	bc.Blocks = append(bc.Blocks, newBlock)

	// Limpiar transacciones pendientes
	bc.PendingTxs = []*Transaction{}

	fmt.Printf("\n✅ Bloque %d minado exitosamente!\n", newBlock.Index)
	fmt.Printf("   Hash: %s\n", newBlock.Hash)
}

// GetBalance obtiene el saldo de una cuenta
func (bc *Blockchain) GetBalance(address string) float64 {
	return bc.AccountState.GetBalance(address)
}

// GetNonce obtiene el nonce actual de una cuenta
func (bc *Blockchain) GetNonce(address string) int {
	return bc.AccountState.GetAccount(address).Nonce
}

// IsValid verifica que toda la blockchain sea válida
func (bc *Blockchain) IsValid() bool {
	// Primero verificar el bloque génesis (índice 0)
	if len(bc.Blocks) > 0 {
		genesisBlock := bc.Blocks[0]
		if !genesisBlock.IsValid(bc.Difficulty) {
			fmt.Printf("❌ Bloque génesis (#0) es inválido\n")
			return false
		}
	}

	// Luego verificar el resto de bloques y sus enlaces
	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		previousBlock := bc.Blocks[i-1]

		// 1. Verificar que el bloque en sí sea válido
		if !currentBlock.IsValid(bc.Difficulty) {
			fmt.Printf("❌ Bloque #%d es inválido\n", i)
			return false
		}

		// 2. Verificar que el hash anterior coincida
		if currentBlock.PreviousHash != previousBlock.Hash {
			fmt.Printf("❌ Cadena rota en bloque #%d\n", i)
			fmt.Printf("   PreviousHash del bloque: %s\n", currentBlock.PreviousHash)
			fmt.Printf("   Hash del bloque anterior: %s\n", previousBlock.Hash)
			return false
		}
	}

	return true
}

// Print muestra toda la blockchain
func (bc *Blockchain) Print() {
	fmt.Println("\n" + "╔════════════════════════════════════════╗")
	fmt.Printf("║      BLOCKCHAIN (Dificultad: %d)       ║\n", bc.Difficulty)
	fmt.Printf("║      Total bloques: %d                  ║\n", len(bc.Blocks))
	fmt.Println("╚════════════════════════════════════════╝")

	for _, block := range bc.Blocks {
		block.Print()
	}
}

// PrintPendingTransactions muestra las transacciones pendientes
func (bc *Blockchain) PrintPendingTransactions() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║      TRANSACCIONES PENDIENTES          ║")
	fmt.Println("╚════════════════════════════════════════╝")

	if len(bc.PendingTxs) == 0 {
		fmt.Println("\n   (No hay transacciones pendientes)")
		return
	}

	for i, tx := range bc.PendingTxs {
		fmt.Printf("\n%d. From: %s\n", i+1, tx.From[:16]+"...")

		// Determinar tipo de transacción
		if tx.IsContractDeployment() {
			fmt.Println("   To: (CONTRATO - DEPLOYMENT)")
			fmt.Printf("   Monto: %.2f MTC\n", tx.Amount)
			fmt.Printf("   Data: %d bytes\n", len(tx.Data))
		} else if tx.To == "" {
			fmt.Println("   To: (Sin destinatario)")
		} else if len(tx.To) >= 8 {
			fmt.Printf("   To: %s\n", tx.To[:16]+"...")
			fmt.Printf("   Monto: %.2f MTC\n", tx.Amount)
			if len(tx.Data) > 0 {
				fmt.Printf("   Data: %d bytes (LLAMADA A CONTRATO)\n", len(tx.Data))
			}
		} else {
			fmt.Printf("   To: %s\n", tx.To)
			fmt.Printf("   Monto: %.2f MTC\n", tx.Amount)
		}

		fmt.Printf("   Nonce: %d\n", tx.Nonce)
		fmt.Printf("   Firmada: %v\n", tx.Signature != "")
	}
}

// DeployContract despliega un contrato en la blockchain
func (bc *Blockchain) DeployContract(owner string, bytecode []byte) (*evm.Contract, error) {
	// Crear el contrato
	contract := evm.NewContract(owner, bytecode)

	// Guardar en la blockchain
	bc.Contracts[contract.Address] = contract

	fmt.Printf("\n📜 Contrato desplegado en: %s\n", contract.Address)

	return contract, nil
}

// GetContract obtiene un contrato por su dirección
func (bc *Blockchain) GetContract(address string) (*evm.Contract, error) {
	contract, exists := bc.Contracts[address]
	if !exists {
		return nil, fmt.Errorf("contrato no encontrado: %s", address)
	}
	return contract, nil
}

// ExecuteContract ejecuta un contrato
func (bc *Blockchain) ExecuteContract(address string, gas uint64) error {
	contract, err := bc.GetContract(address)
	if err != nil {
		return err
	}

	fmt.Printf("\n⚙️  Ejecutando contrato %s...\n", address[:16]+"...")

	remainingGas, err := contract.Execute(gas)
	if err != nil {
		return fmt.Errorf("error ejecutando contrato: %v", err)
	}

	fmt.Printf("✅ Contrato ejecutado. Gas usado: %d\n", gas-remainingGas)

	return nil
}

// ListContracts muestra todos los contratos desplegados
func (bc *Blockchain) ListContracts() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║      CONTRATOS DESPLEGADOS             ║")
	fmt.Println("╚════════════════════════════════════════╝")

	if len(bc.Contracts) == 0 {
		fmt.Println("   (No hay contratos desplegados)")
		return
	}

	i := 1
	for address, contract := range bc.Contracts {
		fmt.Printf("\n%d. %s\n", i, address)
		fmt.Printf("   Owner:    %s\n", contract.Owner[:16]+"...")
		fmt.Printf("   Bytecode: %d bytes\n", len(contract.Bytecode))
		fmt.Printf("   Storage:  %d keys\n", len(contract.Storage.Data))
		i++
	}
}

// ==================== FUNCIONES AUXILIARES DE CONVERSIÓN ====================

// blockToHeader convierte nuestro Block al formato BlockHeader de ChainDB
func blockToHeader(block *Block) *rawdb.BlockHeader {
	// Convertir hashes de hex string a []byte
	parentHash, _ := hex.DecodeString(block.PreviousHash)
	if parentHash == nil {
		parentHash = []byte(block.PreviousHash) // Para el bloque génesis que tiene "0"
	}

	hashBytes, _ := hex.DecodeString(block.Hash)

	return &rawdb.BlockHeader{
		ParentHash:  parentHash,
		Number:      uint64(block.Index),
		StateRoot:   block.StateRoot,
		TxRoot:      block.TxRoot,
		ReceiptRoot: block.ReceiptRoot,
		Timestamp:   block.Timestamp.Unix(),
		Difficulty:  0, // La dificultad se almacena en Blockchain, no en Block
		Nonce:       block.Nonce,
		Hash:        hashBytes,
	}
}

// blockToBody convierte nuestro Block al formato BlockBody de ChainDB
func blockToBody(block *Block) *rawdb.BlockBody {
	// Serializar cada transacción a JSON
	txBytes := make([][]byte, len(block.Transactions))
	for i, tx := range block.Transactions {
		txData, err := json.Marshal(tx)
		if err != nil {
			fmt.Printf("⚠️  Error serializando transacción %d: %v\n", i, err)
			continue
		}
		txBytes[i] = txData
	}

	return &rawdb.BlockBody{
		Transactions: txBytes,
	}
}

// headerToBlock convierte rawdb.BlockHeader a nuestro Block
func headerToBlock(header *rawdb.BlockHeader, body *rawdb.BlockBody) *Block {
	// Deserializar transacciones desde JSON
	transactions := make([]*Transaction, 0, len(body.Transactions))
	for i, txData := range body.Transactions {
		var tx Transaction
		if err := json.Unmarshal(txData, &tx); err != nil {
			fmt.Printf("⚠️  Error deserializando transacción %d: %v\n", i, err)
			continue
		}
		transactions = append(transactions, &tx)
	}

	return &Block{
		Index:        int(header.Number),
		Timestamp:    time.Unix(header.Timestamp, 0),
		Transactions: transactions,
		PreviousHash: hex.EncodeToString(header.ParentHash),
		Hash:         hex.EncodeToString(header.Hash),
		Nonce:        header.Nonce,
		StateRoot:    header.StateRoot,
		TxRoot:       header.TxRoot,
		ReceiptRoot:  header.ReceiptRoot,
	}
}

// Close cierra la base de datos
func (bc *Blockchain) Close() error {
	if bc.db != nil {
		return bc.db.Close()
	}
	return nil
}
