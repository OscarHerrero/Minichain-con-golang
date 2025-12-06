package blockchain

import (
	"fmt"
)

// Blockchain es la cadena completa de bloques
type Blockchain struct {
	Blocks       []*Block       // Array de bloques
	Difficulty   int            // Dificultad del minado (ej: 3 = "000...")
	AccountState *AccountState  // Estado de todas las cuentas
	PendingTxs   []*Transaction // Transacciones pendientes (mempool)
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
	}

	return bc
}

// AddTransaction añade una transacción al mempool (pendientes)
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	// Validar la transacción
	if err := tx.Validate(bc.AccountState); err != nil {
		return fmt.Errorf("transacción inválida: %v", err)
	}

	// Añadir al mempool
	bc.PendingTxs = append(bc.PendingTxs, tx)

	fmt.Printf("✅ Transacción añadida al mempool (total: %d pendientes)\n", len(bc.PendingTxs))

	return nil
}

// MineBlock mina un nuevo bloque con las transacciones pendientes
func (bc *Blockchain) MineBlock() error {
	if len(bc.PendingTxs) == 0 {
		return fmt.Errorf("no hay transacciones pendientes")
	}

	// Obtener el último bloque de la cadena
	previousBlock := bc.Blocks[len(bc.Blocks)-1]

	// Crear nuevo bloque con las transacciones pendientes
	newBlock := NewBlock(
		previousBlock.Index+1,
		bc.PendingTxs,
		previousBlock.Hash,
	)

	// Minar el nuevo bloque (encontrar un hash válido)
	newBlock.MineBlock(bc.Difficulty)

	// Ejecutar todas las transacciones del bloque
	for _, tx := range newBlock.Transactions {
		if err := tx.Execute(bc.AccountState); err != nil {
			// Esto no debería pasar si validamos antes
			return fmt.Errorf("error ejecutando transacción: %v", err)
		}
	}

	// Añadir bloque a la cadena
	bc.Blocks = append(bc.Blocks, newBlock)

	// Limpiar el mempool
	bc.PendingTxs = []*Transaction{}

	fmt.Printf("\n✨ Bloque minado y añadido a la cadena (total: %d bloques)\n", len(bc.Blocks))

	return nil
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
		fmt.Println("   (No hay transacciones pendientes)")
		return
	}

	for i, tx := range bc.PendingTxs {
		fmt.Printf("\n%d. %.2f MTC: %s → %s\n",
			i+1, tx.Amount, tx.From[:8]+"...", tx.To[:8]+"...")
	}
}
