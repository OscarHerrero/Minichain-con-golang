package evm

import (
	"fmt"
	"math/big"
)

// ExecutionContext representa el contexto de ejecución de un contrato
type ExecutionContext struct {
	Stack    *Stack
	Memory   *Memory
	Storage  *Storage // Referencia al storage del contrato
	Code     []byte
	PC       int
	Gas      uint64
	Stopped  bool
	Verbose  bool
	Contract *Contract // Referencia al contrato
}

// EVMInterpreter es el intérprete singleton de la EVM
type EVMInterpreter struct {
	GasTable map[OpCode]uint64
}

// Instancia global del intérprete (singleton)
var GlobalInterpreter = NewEVMInterpreter()

// NewEVMInterpreter crea un nuevo intérprete
func NewEVMInterpreter() *EVMInterpreter {
	return &EVMInterpreter{
		GasTable: gasCosts,
	}
}

// Run ejecuta el bytecode en un contexto dado
func (interp *EVMInterpreter) Run(ctx *ExecutionContext) error {
	// Imprimir header solo si verbose
	if ctx.Verbose {
		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║         EJECUTANDO BYTECODE            ║")
		fmt.Println("╚════════════════════════════════════════╝")
		fmt.Printf("📝 Bytecode: %x\n", ctx.Code)
		fmt.Printf("⛽ Gas disponible: %d\n", ctx.Gas)
	}

	stepCount := 0

	for ctx.PC < len(ctx.Code) && !ctx.Stopped {
		// Leer el opcode actual
		if ctx.PC >= len(ctx.Code) {
			break
		}

		op := OpCode(ctx.Code[ctx.PC])

		// Imprimir paso solo si verbose
		if ctx.Verbose {
			stepCount++
			fmt.Printf("\n━━━ Paso %d ━━━\n", stepCount)
			fmt.Printf("PC: %d | Opcode: %s (0x%02x) | Gas: %d\n",
				ctx.PC, op.String(), byte(op), ctx.Gas)
		}

		// Verificar gas
		gasCost := interp.GetGasCost(op)
		if ctx.Gas < gasCost {
			return fmt.Errorf("out of gas en PC=%d: necesita %d, tiene %d", ctx.PC, gasCost, ctx.Gas)
		}
		ctx.Gas -= gasCost

		// Ejecutar opcode
		if err := interp.ExecuteOpcode(op, ctx); err != nil {
			return fmt.Errorf("error en PC=%d: %v", ctx.PC, err)
		}

		// Avanzar PC (si no fue modificado por JUMP)
		if !op.IsJump() {
			ctx.PC++
		}
	}

	if ctx.Verbose {
		fmt.Printf("\n✅ Ejecución completada\n")
		fmt.Printf("⛽ Gas restante: %d\n", ctx.Gas)
	}

	return nil
}

// GetGasCost devuelve el costo de gas de un opcode
func (interp *EVMInterpreter) GetGasCost(op OpCode) uint64 {
	if cost, exists := interp.GasTable[op]; exists {
		return cost
	}
	return 3 // Costo por defecto
}

// ExecuteOpcode ejecuta un opcode específico
func (interp *EVMInterpreter) ExecuteOpcode(op OpCode, ctx *ExecutionContext) error {
	switch op {
	case STOP:
		return interp.opStop(ctx)
	case ADD:
		return interp.opAdd(ctx)
	case MUL:
		return interp.opMul(ctx)
	case SUB:
		return interp.opSub(ctx)
	case DIV:
		return interp.opDiv(ctx)
	case MOD:
		return interp.opMod(ctx)
	case LT:
		return interp.opLt(ctx)
	case GT:
		return interp.opGt(ctx)
	case EQ:
		return interp.opEq(ctx)
	case POP:
		return interp.opPop(ctx)
	case MLOAD:
		return interp.opMload(ctx)
	case MSTORE:
		return interp.opMstore(ctx)
	case SLOAD:
		return interp.opSload(ctx)
	case SSTORE:
		return interp.opSstore(ctx)
	case PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH32:
		return interp.opPush(op, ctx)
	case DUP1, DUP2:
		return interp.opDup(op, ctx)
	case SWAP1, SWAP2:
		return interp.opSwap(op, ctx)
	default:
		return fmt.Errorf("opcode no implementado: %s (0x%02x)", op.String(), byte(op))
	}
}

// ============================================
// IMPLEMENTACIÓN DE OPCODES
// ============================================

func (interp *EVMInterpreter) opStop(ctx *ExecutionContext) error {
	if ctx.Verbose {
		fmt.Println("→ STOP: Deteniendo ejecución")
	}
	ctx.Stopped = true
	return nil
}

func (interp *EVMInterpreter) opAdd(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 2 {
		return fmt.Errorf("stack underflow: ADD necesita 2 valores")
	}

	a, _ := ctx.Stack.Pop()
	b, _ := ctx.Stack.Pop()
	result := new(big.Int).Add(a, b)
	ctx.Stack.Push(result)

	if ctx.Verbose {
		fmt.Printf("→ ADD: %s + %s = %s\n", a.String(), b.String(), result.String())
	}

	return nil
}

func (interp *EVMInterpreter) opMul(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 2 {
		return fmt.Errorf("stack underflow")
	}

	a, _ := ctx.Stack.Pop()
	b, _ := ctx.Stack.Pop()
	result := new(big.Int).Mul(a, b)
	ctx.Stack.Push(result)

	if ctx.Verbose {
		fmt.Printf("→ MUL: %s * %s = %s\n", a.String(), b.String(), result.String())
	}

	return nil
}

func (interp *EVMInterpreter) opSub(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 2 {
		return fmt.Errorf("stack underflow")
	}

	a, _ := ctx.Stack.Pop()
	b, _ := ctx.Stack.Pop()
	result := new(big.Int).Sub(a, b)
	ctx.Stack.Push(result)

	if ctx.Verbose {
		fmt.Printf("→ SUB: %s - %s = %s\n", a.String(), b.String(), result.String())
	}

	return nil
}

func (interp *EVMInterpreter) opDiv(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 2 {
		return fmt.Errorf("stack underflow")
	}

	a, _ := ctx.Stack.Pop()
	b, _ := ctx.Stack.Pop()

	if b.Sign() == 0 {
		// División por cero → resultado 0 (según especificación EVM)
		ctx.Stack.Push(big.NewInt(0))
	} else {
		result := new(big.Int).Div(a, b)
		ctx.Stack.Push(result)
	}

	if ctx.Verbose {
		fmt.Printf("→ DIV: %s / %s\n", a.String(), b.String())
	}

	return nil
}

func (interp *EVMInterpreter) opMod(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 2 {
		return fmt.Errorf("stack underflow")
	}

	a, _ := ctx.Stack.Pop()
	b, _ := ctx.Stack.Pop()

	if b.Sign() == 0 {
		ctx.Stack.Push(big.NewInt(0))
	} else {
		result := new(big.Int).Mod(a, b)
		ctx.Stack.Push(result)
	}

	if ctx.Verbose {
		fmt.Printf("→ MOD: %s %% %s\n", a.String(), b.String())
	}

	return nil
}

func (interp *EVMInterpreter) opLt(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 2 {
		return fmt.Errorf("stack underflow")
	}

	a, _ := ctx.Stack.Pop()
	b, _ := ctx.Stack.Pop()

	if a.Cmp(b) < 0 {
		ctx.Stack.Push(big.NewInt(1))
	} else {
		ctx.Stack.Push(big.NewInt(0))
	}

	if ctx.Verbose {
		fmt.Printf("→ LT: %s < %s\n", a.String(), b.String())
	}

	return nil
}

func (interp *EVMInterpreter) opGt(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 2 {
		return fmt.Errorf("stack underflow")
	}

	a, _ := ctx.Stack.Pop()
	b, _ := ctx.Stack.Pop()

	if a.Cmp(b) > 0 {
		ctx.Stack.Push(big.NewInt(1))
	} else {
		ctx.Stack.Push(big.NewInt(0))
	}

	if ctx.Verbose {
		fmt.Printf("→ GT: %s > %s\n", a.String(), b.String())
	}

	return nil
}

func (interp *EVMInterpreter) opEq(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 2 {
		return fmt.Errorf("stack underflow")
	}

	a, _ := ctx.Stack.Pop()
	b, _ := ctx.Stack.Pop()

	if a.Cmp(b) == 0 {
		ctx.Stack.Push(big.NewInt(1))
	} else {
		ctx.Stack.Push(big.NewInt(0))
	}

	if ctx.Verbose {
		fmt.Printf("→ EQ: %s == %s\n", a.String(), b.String())
	}

	return nil
}

func (interp *EVMInterpreter) opPop(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 1 {
		return fmt.Errorf("stack underflow")
	}

	ctx.Stack.Pop()

	if ctx.Verbose {
		fmt.Println("→ POP: Eliminado del stack")
	}

	return nil
}

func (interp *EVMInterpreter) opMload(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 1 {
		return fmt.Errorf("stack underflow")
	}

	offset, _ := ctx.Stack.Pop()
	value, _ := ctx.Memory.Load(int(offset.Int64()), 32)
	ctx.Stack.Push(new(big.Int).SetBytes(value))

	if ctx.Verbose {
		fmt.Printf("→ MLOAD: memory[%d]\n", offset.Int64())
	}

	return nil
}

func (interp *EVMInterpreter) opMstore(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 2 {
		return fmt.Errorf("stack underflow")
	}

	offset, _ := ctx.Stack.Pop()
	value, _ := ctx.Stack.Pop()

	ctx.Memory.Store(int(offset.Int64()), value.Bytes())

	if ctx.Verbose {
		fmt.Printf("→ MSTORE: memory[%d] = %s\n", offset.Int64(), value.String())
	}

	return nil
}

func (interp *EVMInterpreter) opSload(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 1 {
		return fmt.Errorf("stack underflow")
	}

	key, _ := ctx.Stack.Pop()
	value := ctx.Storage.Load(key)
	ctx.Stack.Push(value)

	if ctx.Verbose {
		fmt.Printf("→ SLOAD: storage[%s] = %s\n", key.String(), value.String())
	}

	return nil
}

func (interp *EVMInterpreter) opSstore(ctx *ExecutionContext) error {
	if ctx.Stack.Len() < 2 {
		return fmt.Errorf("stack underflow")
	}

	key, _ := ctx.Stack.Pop()
	value, _ := ctx.Stack.Pop()

	ctx.Storage.Store(key, value)

	if ctx.Verbose {
		fmt.Printf("→ SSTORE: storage[%s] = %s\n", key.String(), value.String())
	}

	return nil
}

func (interp *EVMInterpreter) opPush(op OpCode, ctx *ExecutionContext) error {
	pushSize := op.PushSize()

	if ctx.PC+pushSize >= len(ctx.Code) {
		return fmt.Errorf("código incompleto para PUSH")
	}

	valueBytes := ctx.Code[ctx.PC+1 : ctx.PC+1+pushSize]
	value := new(big.Int).SetBytes(valueBytes)
	ctx.Stack.Push(value)

	if ctx.Verbose {
		fmt.Printf("→ %s: Push %d (bytes: %x)\n", op.String(), value.Int64(), valueBytes)
	}

	ctx.PC += pushSize
	return nil
}

func (interp *EVMInterpreter) opDup(op OpCode, ctx *ExecutionContext) error {
	n := int(op - DUP1 + 1)

	if ctx.Stack.Len() < n {
		return fmt.Errorf("stack underflow")
	}

	value := ctx.Stack.data[ctx.Stack.Len()-n]
	ctx.Stack.Push(new(big.Int).Set(value))

	if ctx.Verbose {
		fmt.Printf("→ %s: Duplicado posición %d\n", op.String(), n)
	}

	return nil
}

func (interp *EVMInterpreter) opSwap(op OpCode, ctx *ExecutionContext) error {
	n := int(op - SWAP1 + 1)

	if ctx.Stack.Len() < n+1 {
		return fmt.Errorf("stack underflow")
	}

	top := ctx.Stack.Len() - 1
	ctx.Stack.data[top], ctx.Stack.data[top-n] = ctx.Stack.data[top-n], ctx.Stack.data[top]

	if ctx.Verbose {
		fmt.Printf("→ %s: Intercambiado posiciones\n", op.String())
	}

	return nil
}
