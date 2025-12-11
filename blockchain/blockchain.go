package blockchain

import (
	"fmt"
	"minichain/evm"
	"time"
)

// Blockchain es la cadena completa de bloques
type Blockchain struct {
	Blocks       []*Block                 // Array de bloques
	Difficulty   int                      // Dificultad del minado (ej: 3 = "000...")
	AccountState *AccountState            // Estado de todas las cuentas
	PendingTxs   []*Transaction           // Transacciones pendientes (mempool)
	Contracts    map[string]*evm.Contract // Contratos desplegados
}

// NewBlockchain crea una nueva blockchain con el bloque génesis
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

	// Minar el bloque
	fmt.Printf("\n⛏️  Minando bloque %d (dificultad: %d, %d transacciones)...\n",
		newBlock.Index, bc.Difficulty, len(bc.PendingTxs))

	newBlock.MineBlock(bc.Difficulty)

	// EJECUTAR TRANSACCIONES (incluye contratos)
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

		// Ejecutar (incluye contratos si aplica)
		if err := tx.Execute(bc.AccountState, bc); err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}

		if tx.Amount > 0 {
			fmt.Printf("   ✅ Fondos transferidos\n")
		}
	}

	// Añadir bloque a la cadena
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
