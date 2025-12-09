package evm

import (
	"fmt"
	"math/big"
)

// VM es la máquina virtual de Ethereum
type VM struct {
	Stack   *Stack   // Pila de operaciones
	Memory  *Memory  // Memoria temporal
	Storage *Storage // Almacenamiento permanente
	Code    []byte   // Bytecode a ejecutar
	PC      int      // Program Counter (posición actual)
	Gas     uint64   // Gas disponible
	Stopped bool     // Si la ejecución terminó
}

// NewVM crea una nueva instancia de la VM
func NewVM(code []byte, gas uint64) *VM {
	return &VM{
		Stack:   NewStack(),
		Memory:  NewMemory(),
		Storage: NewStorage(),
		Code:    code,
		PC:      0,
		Gas:     gas,
		Stopped: false,
	}
}

// Run ejecuta el bytecode
func (vm *VM) Run() error {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║         EJECUTANDO BYTECODE            ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("📝 Bytecode: %x\n", vm.Code)
	fmt.Printf("⛽ Gas disponible: %d\n\n", vm.Gas)

	step := 0

	for vm.PC < len(vm.Code) && !vm.Stopped {
		step++

		// Leer el opcode actual
		op := OpCode(vm.Code[vm.PC])

		fmt.Printf("━━━ Paso %d ━━━\n", step)
		fmt.Printf("PC: %d | Opcode: %s (0x%02x) | Gas: %d\n", vm.PC, op.String(), byte(op), vm.Gas)

		// Verificar gas
		gasCost := op.GetGasCost()
		if vm.Gas < gasCost {
			return fmt.Errorf("out of gas: necesita %d, tiene %d", gasCost, vm.Gas)
		}
		vm.Gas -= gasCost

		// Ejecutar el opcode
		if err := vm.executeOpcode(op); err != nil {
			return fmt.Errorf("error en PC=%d: %v", vm.PC, err)
		}

		// Avanzar el PC (algunas instrucciones lo modifican)
		if !vm.Stopped {
			vm.PC++
		}

		fmt.Println()
	}

	fmt.Println("✅ Ejecución completada")
	fmt.Printf("⛽ Gas restante: %d\n", vm.Gas)

	return nil
}

// executeOpcode ejecuta un opcode específico
func (vm *VM) executeOpcode(op OpCode) error {
	switch op {
	case STOP:
		return vm.opStop()
	case ADD:
		return vm.opAdd()
	case MUL:
		return vm.opMul()
	case SUB:
		return vm.opSub()
	case DIV:
		return vm.opDiv()
	case MOD:
		return vm.opMod()
	case LT:
		return vm.opLT()
	case GT:
		return vm.opGT()
	case EQ:
		return vm.opEQ()
	case POP:
		return vm.opPop()
	case MLOAD:
		return vm.opMLoad()
	case MSTORE:
		return vm.opMStore()
	case SLOAD:
		return vm.opSLoad()
	case SSTORE:
		return vm.opSStore()
	case PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH32:
		return vm.opPush(op)
	case DUP1, DUP2:
		return vm.opDup(op)
	case SWAP1, SWAP2:
		return vm.opSwap(op)
	default:
		return fmt.Errorf("opcode no implementado: %s (0x%02x)", op.String(), op)
	}
}

// Implementación de cada opcode

func (vm *VM) opStop() error {
	fmt.Println("→ STOP: Deteniendo ejecución")
	vm.Stopped = true
	return nil
}

func (vm *VM) opAdd() error {
	a, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	b, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	result := new(big.Int).Add(a, b)
	vm.Stack.Push(result)

	fmt.Printf("→ ADD: %s + %s = %s\n", a.String(), b.String(), result.String())
	return nil
}

func (vm *VM) opMul() error {
	a, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	b, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	result := new(big.Int).Mul(a, b)
	vm.Stack.Push(result)

	fmt.Printf("→ MUL: %s * %s = %s\n", a.String(), b.String(), result.String())
	return nil
}

func (vm *VM) opSub() error {
	a, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	b, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	result := new(big.Int).Sub(a, b)
	vm.Stack.Push(result)

	fmt.Printf("→ SUB: %s - %s = %s\n", a.String(), b.String(), result.String())
	return nil
}

func (vm *VM) opDiv() error {
	a, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	b, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	if b.Cmp(big.NewInt(0)) == 0 {
		// División por cero → resultado 0 en Ethereum
		vm.Stack.Push(big.NewInt(0))
		fmt.Println("→ DIV: División por cero, resultado = 0")
		return nil
	}

	result := new(big.Int).Div(a, b)
	vm.Stack.Push(result)

	fmt.Printf("→ DIV: %s / %s = %s\n", a.String(), b.String(), result.String())
	return nil
}

func (vm *VM) opMod() error {
	a, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	b, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	if b.Cmp(big.NewInt(0)) == 0 {
		vm.Stack.Push(big.NewInt(0))
		fmt.Println("→ MOD: Módulo por cero, resultado = 0")
		return nil
	}

	result := new(big.Int).Mod(a, b)
	vm.Stack.Push(result)

	fmt.Printf("→ MOD: %s %% %s = %s\n", a.String(), b.String(), result.String())
	return nil
}

func (vm *VM) opLT() error {
	a, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	b, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	var result *big.Int
	if a.Cmp(b) < 0 {
		result = big.NewInt(1) // true
	} else {
		result = big.NewInt(0) // false
	}
	vm.Stack.Push(result)

	fmt.Printf("→ LT: %s < %s = %s\n", a.String(), b.String(), result.String())
	return nil
}

func (vm *VM) opGT() error {
	a, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	b, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	var result *big.Int
	if a.Cmp(b) > 0 {
		result = big.NewInt(1) // true
	} else {
		result = big.NewInt(0) // false
	}
	vm.Stack.Push(result)

	fmt.Printf("→ GT: %s > %s = %s\n", a.String(), b.String(), result.String())
	return nil
}

func (vm *VM) opEQ() error {
	a, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	b, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	var result *big.Int
	if a.Cmp(b) == 0 {
		result = big.NewInt(1) // true
	} else {
		result = big.NewInt(0) // false
	}
	vm.Stack.Push(result)

	fmt.Printf("→ EQ: %s == %s = %s\n", a.String(), b.String(), result.String())
	return nil
}

func (vm *VM) opPop() error {
	value, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	fmt.Printf("→ POP: Descartado %s\n", value.String())
	return nil
}

func (vm *VM) opMLoad() error {
	offset, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	// Cargar 32 bytes de memoria
	data, err := vm.Memory.Load(int(offset.Int64()), 32)
	if err != nil {
		return err
	}

	value := new(big.Int).SetBytes(data)
	vm.Stack.Push(value)

	fmt.Printf("→ MLOAD: memory[%s] = %s\n", offset.String(), value.String())
	return nil
}

func (vm *VM) opMStore() error {
	offset, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	value, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	// Guardar 32 bytes en memoria
	data := value.Bytes()
	// Pad to 32 bytes
	if len(data) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(data):], data)
		data = padded
	}

	vm.Memory.Store(int(offset.Int64()), data)

	fmt.Printf("→ MSTORE: memory[%s] = %s\n", offset.String(), value.String())
	return nil
}

func (vm *VM) opSLoad() error {
	key, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	value := vm.Storage.Load(key)
	vm.Stack.Push(value)

	fmt.Printf("→ SLOAD: storage[%s] = %s\n", key.String(), value.String())
	return nil
}

func (vm *VM) opSStore() error {
	key, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	value, err := vm.Stack.Pop()
	if err != nil {
		return err
	}

	vm.Storage.Store(key, value)

	fmt.Printf("→ SSTORE: storage[%s] = %s\n", key.String(), value.String())
	return nil
}

func (vm *VM) opPush(op OpCode) error {
	// Cuántos bytes vamos a empujar
	size := op.PushSize()

	// Verificar que hay suficientes bytes
	if vm.PC+size >= len(vm.Code) {
		return fmt.Errorf("PUSH fuera de rango")
	}

	// Leer los bytes siguientes
	data := vm.Code[vm.PC+1 : vm.PC+1+size]

	// Convertir a big.Int
	value := new(big.Int).SetBytes(data)
	vm.Stack.Push(value)

	fmt.Printf("→ %s: Push %s (bytes: %x)\n", op.String(), value.String(), data)

	// Avanzar el PC para saltar los bytes que leímos
	vm.PC += size

	return nil
}

func (vm *VM) opDup(op OpCode) error {
	// DUP1 duplica el 1er elemento (índice 0 desde el tope)
	// DUP2 duplica el 2do elemento (índice 1 desde el tope)
	depth := int(op - DUP1)

	if vm.Stack.Len() <= depth {
		return fmt.Errorf("stack underflow en DUP")
	}

	// Obtener el elemento sin sacarlo
	value := vm.Stack.data[vm.Stack.Len()-1-depth]

	// Empujar una copia
	vm.Stack.Push(new(big.Int).Set(value))

	fmt.Printf("→ %s: Duplicado %s\n", op.String(), value.String())
	return nil
}

func (vm *VM) opSwap(op OpCode) error {
	// SWAP1 intercambia posición 0 y 1
	// SWAP2 intercambia posición 0 y 2
	depth := int(op - SWAP1 + 1)

	if vm.Stack.Len() <= depth {
		return fmt.Errorf("stack underflow en SWAP")
	}

	// Intercambiar
	topIdx := vm.Stack.Len() - 1
	swapIdx := topIdx - depth

	vm.Stack.data[topIdx], vm.Stack.data[swapIdx] = vm.Stack.data[swapIdx], vm.Stack.data[topIdx]

	fmt.Printf("→ %s: Intercambiado posiciones %d y %d\n", op.String(), 0, depth)
	return nil
}

// PrintState muestra el estado completo de la VM
func (vm *VM) PrintState() {
	vm.Stack.Print()
	vm.Memory.Print()
	vm.Storage.Print()
}
