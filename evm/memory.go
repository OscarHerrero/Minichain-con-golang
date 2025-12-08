package evm

import (
	"fmt"
)

// Memory es la memoria temporal de la EVM
// Se borra después de cada ejecución de contrato
type Memory struct {
	data []byte // Array de bytes
}

// NewMemory crea una nueva memoria vacía
func NewMemory() *Memory {
	return &Memory{
		data: make([]byte, 0),
	}
}

// Store guarda datos en una posición de memoria
func (m *Memory) Store(offset int, value []byte) error {
	// Expandir memoria si es necesario
	requiredSize := offset + len(value)
	if requiredSize > len(m.data) {
		// Ethereum cobra gas por expandir memoria
		// Aquí simplemente expandimos
		newData := make([]byte, requiredSize)
		copy(newData, m.data)
		m.data = newData
	}
	
	// Copiar el valor en la posición
	copy(m.data[offset:], value)
	
	return nil
}

// Load carga datos desde una posición de memoria
func (m *Memory) Load(offset, size int) ([]byte, error) {
	// Verificar que no se lea fuera de la memoria
	if offset+size > len(m.data) {
		return nil, fmt.Errorf("memoria fuera de rango")
	}
	
	// Copiar los datos
	result := make([]byte, size)
	copy(result, m.data[offset:offset+size])
	
	return result, nil
}

// Size devuelve el tamaño actual de la memoria
func (m *Memory) Size() int {
	return len(m.data)
}

// Print muestra el contenido de la memoria
func (m *Memory) Print() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║           MEMORY (MEMORIA)             ║")
	fmt.Println("╚════════════════════════════════════════╝")
	
	if len(m.data) == 0 {
		fmt.Println("   (vacía)")
		return
	}
	
	fmt.Printf("Tamaño: %d bytes\n", len(m.data))
	
	// Mostrar en grupos de 32 bytes (como Ethereum)
	for i := 0; i < len(m.data); i += 32 {
		end := i + 32
		if end > len(m.data) {
			end = len(m.data)
		}
		fmt.Printf("[%04d-%04d] %x\n", i, end-1, m.data[i:end])
	}
}
/*

## ¿QUÉ HACE ESTE ARCHIVO?

### **Memory (Memoria Temporal)**

Como la **RAM** de un ordenador:
- Se usa durante la ejecución
- Se borra cuando termina el contrato
- Rápida y barata (poco gas)
```
Guardar "Hola":
Offset 0: [H e l l o]

Guardar "Mundo":
Offset 10: [M u n d o]

Memoria completa:
[H e l l o _ _ _ _ _ M u n d o]
 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14
 */