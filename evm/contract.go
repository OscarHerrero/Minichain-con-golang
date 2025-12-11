package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"minichain/utils"
)

// Contract representa un contrato inteligente desplegado
type Contract struct {
	Address  string   // Dirección del contrato (0xabc...)
	Owner    string   // Dirección del creador
	Bytecode []byte   // Código del contrato
	Storage  *Storage // Estado persistente del contrato
	Balance  float64  // Saldo del contrato (puede recibir fondos)
}

// NewContract crea un nuevo contrato
func NewContract(owner string, bytecode []byte) *Contract {
	// Generar dirección del contrato (hash del bytecode + owner)
	data := fmt.Sprintf("%s:%x", owner, bytecode)
	address := utils.CalculateHash(data)[:40] // Tomar primeros 40 caracteres

	return &Contract{
		Address:  address,
		Owner:    owner,
		Bytecode: bytecode,
		Storage:  NewStorage(),
		Balance:  0,
	}
}

// Execute ejecuta el bytecode del contrato usando el intérprete global
func (c *Contract) Execute(gas uint64) (uint64, error) {
	// Crear contexto de ejecución
	ctx := &ExecutionContext{
		Stack:    NewStack(),
		Memory:   NewMemory(),
		Storage:  c.Storage,  // Referencia al storage del contrato
		Code:     c.Bytecode,
		PC:       0,
		Gas:      gas,
		Stopped:  false,
		Verbose:  true,
		Contract: c,
	}
	
	// Ejecutar con el intérprete global
	if err := GlobalInterpreter.Run(ctx); err != nil {
		return 0, err
	}
	
	// Devolver gas restante
	return ctx.Gas, nil
}

// Call simula llamar a una función del contrato con datos
func (c *Contract) Call(calldata []byte, gas uint64) (uint64, error) {
	// Crear contexto de ejecución
	ctx := &ExecutionContext{
		Stack:    NewStack(),
		Memory:   NewMemory(),
		Storage:  c.Storage,
		Code:     c.Bytecode,
		PC:       0,
		Gas:      gas,
		Stopped:  false,
		Verbose:  true,
		Contract: c,
	}
	
	// Ejecutar con el intérprete global
	if err := GlobalInterpreter.Run(ctx); err != nil {
		return 0, err
	}
	
	return ctx.Gas, nil
}

// GetStorageValue obtiene un valor del storage del contrato
func (c *Contract) GetStorageValue(key *big.Int) *big.Int {
	return c.Storage.Load(key)
}

// Print muestra información del contrato
func (c *Contract) Print() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║         SMART CONTRACT                 ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("📍 Address:  %s\n", c.Address)
	fmt.Printf("👤 Owner:    %s\n", c.Owner[:16]+"...")
	fmt.Printf("💰 Balance:  %.2f MTC\n", c.Balance)
	fmt.Printf("📝 Bytecode: %d bytes (%s...)\n", len(c.Bytecode), hex.EncodeToString(c.Bytecode[:min(8, len(c.Bytecode))]))
	fmt.Printf("💾 Storage:  %d keys\n", len(c.Storage.Data))

	if len(c.Storage.Data) > 0 {
		fmt.Println("\n📊 Storage State:")
		c.Storage.Print()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
