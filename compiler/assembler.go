package compiler

import (
	"encoding/hex"
	"fmt"
	"minichain/evm"
	"strconv"
	"strings"
)

// Assembler convierte código assembly a bytecode
type Assembler struct {
	opcodeMap map[string]evm.OpCode // Nombre → OpCode
}

// NewAssembler crea un nuevo assembler
func NewAssembler() *Assembler {
	return &Assembler{
		opcodeMap: map[string]evm.OpCode{
			"STOP":   evm.STOP,
			"ADD":    evm.ADD,
			"MUL":    evm.MUL,
			"SUB":    evm.SUB,
			"DIV":    evm.DIV,
			"MOD":    evm.MOD,
			"LT":     evm.LT,
			"GT":     evm.GT,
			"EQ":     evm.EQ,
			"POP":    evm.POP,
			"MLOAD":  evm.MLOAD,
			"MSTORE": evm.MSTORE,
			"SLOAD":  evm.SLOAD,
			"SSTORE": evm.SSTORE,
			"JUMP":   evm.JUMP,
			"JUMPI":  evm.JUMPI,
			"PC":     evm.PC,
			"PUSH1":  evm.PUSH1,
			"PUSH2":  evm.PUSH2,
			"PUSH3":  evm.PUSH3,
			"PUSH4":  evm.PUSH4,
			"PUSH5":  evm.PUSH5,
			"PUSH32": evm.PUSH32,
			"DUP1":   evm.DUP1,
			"DUP2":   evm.DUP2,
			"SWAP1":  evm.SWAP1,
			"SWAP2":  evm.SWAP2,
			"RETURN": evm.RETURN,
		},
	}
}

// Assemble convierte código assembly a bytecode
func (a *Assembler) Assemble(code string) ([]byte, error) {
	// Limpiar y separar en líneas
	lines := strings.Split(code, "\n")

	bytecode := []byte{}

	for lineNum, line := range lines {
		// Limpiar espacios y comentarios
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Separar por espacios
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		instruction := strings.ToUpper(parts[0])

		// Verificar si es un opcode conocido
		opcode, exists := a.opcodeMap[instruction]
		if !exists {
			return nil, fmt.Errorf("línea %d: opcode desconocido '%s'", lineNum+1, instruction)
		}

		// Añadir el opcode
		bytecode = append(bytecode, byte(opcode))

		// Si es PUSH, necesitamos el valor
		if opcode.IsPush() {
			if len(parts) < 2 {
				return nil, fmt.Errorf("línea %d: PUSH requiere un valor", lineNum+1)
			}

			// Parsear el valor
			valueStr := parts[1]
			value, err := parseValue(valueStr)
			if err != nil {
				return nil, fmt.Errorf("línea %d: error parseando valor '%s': %v", lineNum+1, valueStr, err)
			}

			// Obtener el tamaño del PUSH
			pushSize := opcode.PushSize()

			// Verificar que el valor cabe en el tamaño
			maxValue := int64(1) << uint(pushSize*8) // 2^(pushSize*8)
			if value >= maxValue {
				return nil, fmt.Errorf("línea %d: valor %d demasiado grande para %s (máx: %d)",
					lineNum+1, value, instruction, maxValue-1)
			}

			// Convertir a bytes (big-endian)
			valueBytes := intToBytes(value, pushSize)
			bytecode = append(bytecode, valueBytes...)
		}
	}

	return bytecode, nil
}

// parseValue parsea un valor (decimal o hexadecimal)
func parseValue(s string) (int64, error) {
	s = strings.TrimSpace(s)

	// ¿Es hexadecimal? (0x...)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		value, err := strconv.ParseInt(s[2:], 16, 64)
		if err != nil {
			return 0, fmt.Errorf("valor hexadecimal inválido: %s", s)
		}
		return value, nil
	}

	// Es decimal
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("valor decimal inválido: %s", s)
	}

	return value, nil
}

// intToBytes convierte un int64 a bytes (big-endian)
func intToBytes(value int64, size int) []byte {
	bytes := make([]byte, size)

	// Convertir de derecha a izquierda (big-endian)
	for i := size - 1; i >= 0; i-- {
		bytes[i] = byte(value & 0xFF)
		value >>= 8
	}

	return bytes
}

// Disassemble convierte bytecode a assembly legible
func (a *Assembler) Disassemble(bytecode []byte) string {
	var output strings.Builder

	pc := 0
	for pc < len(bytecode) {
		op := evm.OpCode(bytecode[pc])

		// Escribir el opcode
		output.WriteString(fmt.Sprintf("%04d: %s", pc, op.String()))

		// Si es PUSH, mostrar el valor
		if op.IsPush() {
			pushSize := op.PushSize()
			if pc+pushSize < len(bytecode) {
				valueBytes := bytecode[pc+1 : pc+1+pushSize]
				output.WriteString(fmt.Sprintf(" 0x%s", hex.EncodeToString(valueBytes)))
				pc += pushSize
			}
		}

		output.WriteString("\n")
		pc++
	}

	return output.String()
}

// PrintBytecode muestra el bytecode de forma bonita
func PrintBytecode(bytecode []byte) {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║            BYTECODE                    ║")
	fmt.Println("╚════════════════════════════════════════╝")

	fmt.Printf("Tamaño: %d bytes\n", len(bytecode))
	fmt.Printf("Hex: %x\n", bytecode)

	// Mostrar en grupos de 16 bytes
	for i := 0; i < len(bytecode); i += 16 {
		end := i + 16
		if end > len(bytecode) {
			end = len(bytecode)
		}
		fmt.Printf("%04d: %x\n", i, bytecode[i:end])
	}
}

/*

## ¿QUÉ HACE ESTE ARCHIVO?

### 1. **Assemble()** - Assembly → Bytecode

Convierte código legible a bytes ejecutables:
```
INPUT (Assembly):
PUSH1 5
PUSH1 3
ADD
STOP

OUTPUT (Bytecode):
[0x60, 0x05, 0x60, 0x03, 0x01, 0x00]
```

**Proceso:**
```
1. Lee línea por línea
2. "PUSH1" → Busca en mapa → 0x60
3. "5" → Convierte a byte → 0x05
4. Concatena todo
```

### 2. **Disassemble()** - Bytecode → Assembly

Lo contrario, convierte bytes en código legible:
```
INPUT (Bytecode):
[0x60, 0x05, 0x60, 0x03, 0x01]

OUTPUT (Assembly):
0000: PUSH1 0x05
0002: PUSH1 0x03
0004: ADD
*/
