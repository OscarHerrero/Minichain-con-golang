package evm

import (
	"fmt"
	"math/big"
)

// Stack es una pila LIFO (Last In, First Out)
// Funciona como una pila de platos: el último en entrar es el primero en salir
type Stack struct {
	data []*big.Int // Usamos big.Int para números de 256 bits (como Ethereum)
}

// NewStack crea una nueva pila vacía
func NewStack() *Stack {
	return &Stack{
		data: make([]*big.Int, 0),
	}
}

// Push añade un valor al tope de la pila
func (s *Stack) Push(value *big.Int) error {
	// Ethereum limita la pila a 1024 elementos
	if len(s.data) >= 1024 {
		return fmt.Errorf("stack overflow: máximo 1024 elementos")
	}

	s.data = append(s.data, value)
	return nil
}

// Pop saca y devuelve el valor del tope de la pila
func (s *Stack) Pop() (*big.Int, error) {
	if len(s.data) == 0 {
		return nil, fmt.Errorf("stack underflow: pila vacía")
	}

	// Obtener el último elemento
	value := s.data[len(s.data)-1]

	// Remover el último elemento
	s.data = s.data[:len(s.data)-1]

	return value, nil
}

// Peek mira el valor del tope SIN sacarlo
func (s *Stack) Peek() (*big.Int, error) {
	if len(s.data) == 0 {
		return nil, fmt.Errorf("stack vacía")
	}

	return s.data[len(s.data)-1], nil
}

// Len devuelve el tamaño actual de la pila
func (s *Stack) Len() int {
	return len(s.data)
}

// Print muestra el contenido de la pila
func (s *Stack) Print() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║              STACK (PILA)              ║")
	fmt.Println("╚════════════════════════════════════════╝")

	if len(s.data) == 0 {
		fmt.Println("   (vacía)")
		return
	}

	// Mostrar desde el tope hacia abajo
	for i := len(s.data) - 1; i >= 0; i-- {
		fmt.Printf("[%d] %s\n", i, s.data[i].String())
	}
}

/*

---

## ¿QUÉ HACE ESTE ARCHIVO?

### **Stack (Pila)**

Es como una **pila de platos**:
- Solo puedes añadir al TOPE (Push)
- Solo puedes sacar del TOPE (Pop)
- No puedes sacar del medio
```
PUSH 5:
┌─────┐
│  5  │ ← tope
└─────┘

PUSH 3:
┌─────┐
│  3  │ ← tope
├─────┤
│  5  │
└─────┘

POP:
┌─────┐
│  5  │ ← tope (3 fue removido)
└─────┘
```

### **¿Por qué big.Int?**

Ethereum usa números de **256 bits**:
- `int` de Go = 64 bits (máximo: 9,223,372,036,854,775,807)
- `big.Int` = ¡¡¡infinito!!! (bueno, limitado por memoria)

Ethereum puede manejar números gigantes como:
```
115792089237316195423570985008687907853269984665640564039457584007913129639935
*/
