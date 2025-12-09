package evm

import (
	"fmt"
	"math/big"
)

// Storage es el almacenamiento persistente del contrato
// Los datos aquí NO se borran (como el disco duro)
type Storage struct {
	Data map[string]*big.Int // key -> value
}

// NewStorage crea un nuevo storage vacío
func NewStorage() *Storage {
	return &Storage{
		Data: make(map[string]*big.Int),
	}
}

// Store guarda un valor en el storage
func (s *Storage) Store(key, value *big.Int) {
	// Convertir la key a string para usar como índice del map
	keyStr := key.String()

	// Si el valor es 0, eliminar la entrada (ahorrar espacio)
	if value.Cmp(big.NewInt(0)) == 0 {
		delete(s.Data, keyStr)
	} else {
		s.Data[keyStr] = new(big.Int).Set(value)
	}
}

// Load carga un valor del storage
func (s *Storage) Load(key *big.Int) *big.Int {
	keyStr := key.String()

	value, exists := s.Data[keyStr]
	if !exists {
		// Si no existe, devolver 0
		return big.NewInt(0)
	}

	return new(big.Int).Set(value)
}

// Print muestra el contenido del storage
func (s *Storage) Print() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║      STORAGE (ALMACENAMIENTO)          ║")
	fmt.Println("╚════════════════════════════════════════╝")

	if len(s.Data) == 0 {
		fmt.Println("   (vacío)")
		return
	}

	for key, value := range s.Data {
		fmt.Printf("Key [%s] = %s\n", key, value.String())
	}
}
