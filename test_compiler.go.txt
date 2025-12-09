package main

import (
	"bufio"
	"fmt"
	"minichain/compiler"
	"minichain/evm"
	"os"
	"strings"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║                                          ║")
	fmt.Println("║       🔨 COMPILADOR EVM v1.0 🔨         ║")
	fmt.Println("║    Assembly → Bytecode → Ejecución      ║")
	fmt.Println("║                                          ║")
	fmt.Println("╚══════════════════════════════════════════╝")

	// Menú de ejemplos
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║        PROGRAMAS DE EJEMPLO            ║")
	fmt.Println("╠════════════════════════════════════════╣")
	fmt.Println("║ 1. Suma: 5 + 3                         ║")
	fmt.Println("║ 2. Fibonacci: fib(5)                   ║")
	fmt.Println("║ 3. Contador en storage                 ║")
	fmt.Println("║ 4. Escribir código custom             ║")
	fmt.Println("╚════════════════════════════════════════╝")

	fmt.Print("\n👉 Selecciona una opción: ")
	var choice int
	fmt.Scan(&choice)

	var assemblyCode string

	switch choice {
	case 1:
		// Suma simple
		assemblyCode = `
// Programa: Sumar 5 + 3
PUSH1 5
PUSH1 3
ADD
STOP
`

	case 2:
		// Fibonacci
		assemblyCode = `
// Programa: Calcular Fibonacci(5)
// Fib(0) = 0, Fib(1) = 1, Fib(n) = Fib(n-1) + Fib(n-2)

PUSH1 0      // a = 0
PUSH1 1      // b = 1
PUSH1 5      // contador = 5

// Loop
DUP1         // duplicar contador
PUSH1 0
EQ           // ¿contador == 0?

// Si no es 0, continuar
SWAP2        // intercambiar a y b
DUP2         // duplicar b
ADD          // a + b
SWAP1        // preparar para siguiente iteración
PUSH1 1
SWAP2
SUB          // contador--

STOP
`

	case 3:
		// Contador en storage
		assemblyCode = `
// Programa: Incrementar contador en storage

// Leer contador actual
PUSH1 0      // key = 0
SLOAD        // cargar storage[0]

// Incrementar
PUSH1 1
ADD

// Guardar de vuelta
PUSH1 0      // key = 0
SSTORE       // storage[0] = valor

STOP
`

	case 4:
		// Código custom
		fmt.Println("\n📝 Escribe tu código assembly (escribe 'FIN' para terminar):")
		fmt.Println("Ejemplo:")
		fmt.Println("  PUSH1 10")
		fmt.Println("  PUSH1 20")
		fmt.Println("  ADD")
		fmt.Println("  STOP")
		fmt.Println("  FIN")
		fmt.Println()

		scanner := bufio.NewScanner(os.Stdin)
		var lines []string

		// Consumir el Enter pendiente del Scan anterior
		scanner.Scan()

		// Leer líneas hasta "FIN"
		for scanner.Scan() {
			line := scanner.Text()
			if strings.ToUpper(strings.TrimSpace(line)) == "FIN" {
				break
			}
			lines = append(lines, line)
		}

		if len(lines) == 0 {
			fmt.Println("❌ No se escribió ningún código")
			return
		}

		assemblyCode = strings.Join(lines, "\n")

	default:
		fmt.Println("❌ Opción inválida")
		return
	}

	// Mostrar el código assembly
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║          CÓDIGO ASSEMBLY               ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println(assemblyCode)

	// Compilar
	fmt.Println("\n🔨 Compilando...")
	assembler := compiler.NewAssembler()

	bytecode, err := assembler.Assemble(assemblyCode)
	if err != nil {
		fmt.Printf("❌ Error de compilación: %v\n", err)
		return
	}

	fmt.Println("✅ Compilación exitosa")

	// Mostrar bytecode
	compiler.PrintBytecode(bytecode)

	// Desensamblar (ingeniería inversa)
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║         DESENSAMBLADO                  ║")
	fmt.Println("╚════════════════════════════════════════╝")
	disassembly := assembler.Disassemble(bytecode)
	fmt.Println(disassembly)

	// Preguntar si ejecutar
	fmt.Print("\n⚡ ¿Ejecutar el bytecode? (s/n): ")
	var execute string
	fmt.Scan(&execute)

	if strings.ToLower(execute) != "s" {
		return
	}

	// Ejecutar
	fmt.Println("\n" + strings.Repeat("═", 50))
	vm := evm.NewVM(bytecode, 1000000)

	if err := vm.Run(); err != nil {
		fmt.Printf("\n❌ Error de ejecución: %v\n", err)
		return
	}

	// Mostrar estado final
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║          ESTADO FINAL                  ║")
	fmt.Println("╚════════════════════════════════════════╝")
	vm.PrintState()
}
