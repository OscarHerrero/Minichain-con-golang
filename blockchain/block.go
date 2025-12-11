package blockchain

import (
	"fmt"
	"minichain/utils"
	"strconv"
	"strings"
	"time"
)

// Block representa un bloque en la blockchain
type Block struct {
	Index        int            // Posición del bloque en la cadena (0, 1, 2...)
	Timestamp    time.Time      // Cuándo se creó el bloque
	Transactions []*Transaction // Lista de transacciones en el bloque
	PreviousHash string         // Hash del bloque anterior (esto crea la "cadena")
	Hash         string         // Hash de ESTE bloque (su huella digital única)
	Nonce        int            // Número que se va probando hasta encontrar un hash válido

	// Merkle roots (estilo Ethereum)
	StateRoot   []byte // Root del árbol de estado (todas las cuentas y contratos)
	TxRoot      []byte // Root del árbol de transacciones
	ReceiptRoot []byte // Root del árbol de receipts (resultados de ejecución)
}

// NewBlock crea un nuevo bloque (sin minar todavía)
func NewBlock(index int, transactions []*Transaction, previousHash string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now(),
		Transactions: transactions,
		PreviousHash: previousHash,
		Nonce:        0, // Empieza en 0, se incrementará al minar
	}
	return block
}

// NewGenesisBlock crea el bloque génesis (bloque especial #0)
func NewGenesisBlock() *Block {
	return &Block{
		Index:        0,
		Timestamp:    time.Now(),
		Transactions: []*Transaction{}, // Sin transacciones
		PreviousHash: "0",
		Nonce:        0,
		StateRoot:    make([]byte, 32), // Root vacío (hash de trie vacío)
		TxRoot:       make([]byte, 32), // Sin transacciones
		ReceiptRoot:  make([]byte, 32), // Sin receipts
	}
}

// getTransactionsData convierte las transacciones a string para el hash
func (b *Block) getTransactionsData() string {
	if len(b.Transactions) == 0 {
		return ""
	}

	// Serializar transacciones a JSON para el hash
	var txData []string
	for _, tx := range b.Transactions {
		// Incluir TODOS los campos que definen la transacción
		txStr := fmt.Sprintf("from=%s|to=%s|amount=%.2f|nonce=%d|data=%x|sig=%s",
			tx.From,
			tx.To,
			tx.Amount,
			tx.Nonce,
			tx.Data,
			tx.Signature,
		)
		txData = append(txData, txStr)
	}

	return strings.Join(txData, "||")
}

// CalculateBlockHash calcula el hash del bloque
// Combina TODOS los datos del bloque en un solo string y hace hash
func (b *Block) CalculateBlockHash() string {
	// Concatenamos todos los datos del bloque
	record := strconv.Itoa(b.Index) +
		b.Timestamp.String() +
		b.getTransactionsData() +
		b.PreviousHash +
		strconv.Itoa(b.Nonce) +
		string(b.StateRoot) +
		string(b.TxRoot) +
		string(b.ReceiptRoot)

	// Calculamos el hash SHA-256 de todo eso
	return utils.CalculateHash(record)
}

// MineBlock realiza el "Proof of Work" - encuentra un hash válido
// difficulty = cuántos ceros debe tener al inicio el hash
func (b *Block) MineBlock(difficulty int) {
	fmt.Printf("\n⛏️  Minando bloque %d (dificultad: %d, %d transacciones)...\n",
		b.Index, difficulty, len(b.Transactions))

	// Probamos diferentes valores de Nonce hasta encontrar un hash válido
	for {
		// Calculamos el hash con el Nonce actual
		b.Hash = b.CalculateBlockHash()

		// ¿Cumple con la dificultad? (¿empieza con suficientes ceros?)
		if utils.MeetsTarget(b.Hash, difficulty) {
			// ¡Encontrado! Este bloque es válido
			fmt.Printf("✅ Bloque minado! Hash: %s (intentos: %d)\n", b.Hash, b.Nonce)
			break
		}

		// No funcionó, probamos con el siguiente número
		b.Nonce++

		// Mostrar progreso cada 100,000 intentos
		if b.Nonce%100000 == 0 {
			fmt.Printf("   Intentando... nonce=%d\n", b.Nonce)
		}
	}
}

// IsValid verifica si el bloque es válido
func (b *Block) IsValid(difficulty int) bool {
	// Recalculamos el hash
	calculatedHash := b.CalculateBlockHash()

	// Verificamos que:
	// 1. El hash almacenado coincide con el calculado
	// 2. El hash cumple con la dificultad
	return b.Hash == calculatedHash && utils.MeetsTarget(b.Hash, difficulty)
}

// Print muestra el bloque de forma bonita
func (b *Block) Print() {
	fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📦 BLOQUE #%d\n", b.Index)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("⏰ Timestamp:     %s\n", b.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("📊 Transacciones: %d\n", len(b.Transactions))

	// Mostrar transacciones si las hay
	if len(b.Transactions) > 0 {
		// Mostrar transacciones
		if len(b.Transactions) > 0 {
			for i, tx := range b.Transactions {
				fmt.Printf("\n📝 Transacción %d:\n", i+1)

				// From (siempre existe)
				if len(tx.From) >= 16 {
					fmt.Printf("   From: %s\n", tx.From[:16]+"...")
				} else {
					fmt.Printf("   From: %s\n", tx.From)
				}

				// To (depende del tipo)
				if tx.IsContractDeployment() {
					fmt.Println("   To: (CONTRATO - DEPLOYMENT)")
					if tx.ContractAddress != "" && len(tx.ContractAddress) >= 16 {
						fmt.Printf("   Contrato desplegado: %s\n", tx.ContractAddress[:16]+"...")
					} else if tx.ContractAddress != "" {
						fmt.Printf("   Contrato desplegado: %s\n", tx.ContractAddress)
					}
				} else if tx.To == "" {
					fmt.Println("   To: (vacío)")
				} else if len(tx.To) >= 16 {
					fmt.Printf("   To: %s\n", tx.To[:16]+"...")
					if len(tx.Data) > 0 {
						fmt.Println("   Tipo: LLAMADA A CONTRATO")
					}
				} else {
					fmt.Printf("   To: %s\n", tx.To)
				}

				// Resto de info
				fmt.Printf("   Monto: %.2f MTC\n", tx.Amount)
				fmt.Printf("   Nonce: %d\n", tx.Nonce)

				if tx.GasUsed > 0 {
					fmt.Printf("   Gas usado: %d\n", tx.GasUsed)
				}

				if len(tx.Data) > 0 && tx.IsContractDeployment() {
					fmt.Printf("   Bytecode: %d bytes\n", len(tx.Data))
				}
			}
		}
	}

	// Mostrar PreviousHash - verificar longitud primero
	if len(b.PreviousHash) <= 16 {
		fmt.Printf("🔗 Previous Hash: %s\n", b.PreviousHash)
	} else {
		fmt.Printf("🔗 Previous Hash: %s...\n", b.PreviousHash[:16])
	}

	// Mostrar Hash - verificar longitud primero
	if len(b.Hash) <= 16 {
		fmt.Printf("🔐 Hash:          %s\n", b.Hash)
	} else {
		fmt.Printf("🔐 Hash:          %s...\n", b.Hash[:16])
	}

	fmt.Printf("🎲 Nonce:         %d\n", b.Nonce)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
