package evm

// OpCode representa un código de operación de la EVM
type OpCode byte

// Definición de todos los opcodes
// Usamos los mismos valores que Ethereum para compatibilidad
const (
	// 0x0 range - Aritméticas
	STOP OpCode = 0x00 // Detener ejecución
	ADD  OpCode = 0x01 // Suma: a + b
	MUL  OpCode = 0x02 // Multiplicación: a * b
	SUB  OpCode = 0x03 // Resta: a - b
	DIV  OpCode = 0x04 // División: a / b
	MOD  OpCode = 0x06 // Módulo: a % b

	// 0x10 range - Comparaciones
	LT OpCode = 0x10 // Menor que: a < b
	GT OpCode = 0x11 // Mayor que: a > b
	EQ OpCode = 0x14 // Igual: a == b

	// 0x50 range - Stack, Memory, Storage
	POP    OpCode = 0x50 // Sacar de la pila
	MLOAD  OpCode = 0x51 // Cargar de memoria
	MSTORE OpCode = 0x52 // Guardar en memoria
	SLOAD  OpCode = 0x54 // Cargar de storage
	SSTORE OpCode = 0x55 // Guardar en storage
	JUMP   OpCode = 0x56 // Salto incondicional
	JUMPI  OpCode = 0x57 // Salto condicional
	PC     OpCode = 0x58 // Program counter (posición actual)

	// 0x60 range - Push
	PUSH1  OpCode = 0x60 // Push 1 byte
	PUSH2  OpCode = 0x61 // Push 2 bytes
	PUSH3  OpCode = 0x62 // Push 3 bytes
	PUSH4  OpCode = 0x63 // Push 4 bytes
	PUSH5  OpCode = 0x64 // Push 5 bytes
	PUSH32 OpCode = 0x7f // Push 32 bytes

	// 0x80 range - Duplicar
	DUP1 OpCode = 0x80 // Duplicar el 1er elemento
	DUP2 OpCode = 0x81 // Duplicar el 2do elemento

	// 0x90 range - Intercambiar
	SWAP1 OpCode = 0x90 // Intercambiar 1er y 2do elemento
	SWAP2 OpCode = 0x91 // Intercambiar 1er y 3er elemento

	// 0xf0 range - System
	RETURN OpCode = 0xf3 // Retornar datos
)

// opcodeNames mapea opcodes a nombres legibles
var opcodeNames = map[OpCode]string{
	STOP:   "STOP",
	ADD:    "ADD",
	MUL:    "MUL",
	SUB:    "SUB",
	DIV:    "DIV",
	MOD:    "MOD",
	LT:     "LT",
	GT:     "GT",
	EQ:     "EQ",
	POP:    "POP",
	MLOAD:  "MLOAD",
	MSTORE: "MSTORE",
	SLOAD:  "SLOAD",
	SSTORE: "SSTORE",
	JUMP:   "JUMP",
	JUMPI:  "JUMPI",
	PC:     "PC",
	PUSH1:  "PUSH1",
	PUSH2:  "PUSH2",
	PUSH3:  "PUSH3",
	PUSH4:  "PUSH4",
	PUSH5:  "PUSH5",
	PUSH32: "PUSH32",
	DUP1:   "DUP1",
	DUP2:   "DUP2",
	SWAP1:  "SWAP1",
	SWAP2:  "SWAP2",
	RETURN: "RETURN",
}

// String devuelve el nombre del opcode
func (op OpCode) String() string {
	if name, exists := opcodeNames[op]; exists {
		return name
	}
	return "UNKNOWN"
}

// IsPush verifica si un opcode es PUSH
func (op OpCode) IsPush() bool {
	return op >= PUSH1 && op <= PUSH32
}

// PushSize devuelve cuántos bytes empuja un PUSH
func (op OpCode) PushSize() int {
	if op >= PUSH1 && op <= PUSH32 {
		return int(op) - int(PUSH1) + 1
	}
	return 0
}

// IsJump verifica si el opcode es un salto
func (op OpCode) IsJump() bool {
	return op == JUMP || op == JUMPI
}

// gasCosts define el costo en gas de cada operación
var gasCosts = map[OpCode]uint64{
	STOP:   0,
	ADD:    3,
	MUL:    5,
	SUB:    3,
	DIV:    5,
	MOD:    5,
	LT:     3,
	GT:     3,
	EQ:     3,
	POP:    2,
	MLOAD:  3,
	MSTORE: 3,
	SLOAD:  200,   // Leer storage es caro
	SSTORE: 20000, // Escribir storage es MUY caro
	JUMP:   8,
	JUMPI:  10,
	PC:     2,
	PUSH1:  3,
	PUSH2:  3,
	PUSH3:  3,
	PUSH4:  3,
	PUSH5:  3,
	PUSH32: 3,
	DUP1:   3,
	DUP2:   3,
	SWAP1:  3,
	SWAP2:  3,
	RETURN: 0,
}

// GetGasCost devuelve el costo en gas de un opcode
func (op OpCode) GetGasCost() uint64 {
	if cost, exists := gasCosts[op]; exists {
		return cost
	}
	return 0
}
