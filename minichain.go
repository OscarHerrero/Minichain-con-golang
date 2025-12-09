package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"minichain/blockchain"
	"minichain/compiler" // ← AÑADIR
	"minichain/crypto"   // ← AÑADIR
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║                                          ║")
	fmt.Println("║          🔗 MINICHAIN v2.0 🔗           ║")
	fmt.Println("║   Blockchain con Transacciones          ║")
	fmt.Println("║                                          ║")
	fmt.Println("╚══════════════════════════════════════════╝")

	// Crear la blockchain con dificultad 3
	fmt.Println("\n🚀 Creando blockchain...")
	bc := blockchain.NewBlockchain(3)

	// Crear una wallet para gestionar cuentas
	wallet := crypto.NewWallet()

	// Crear 3 cuentas de ejemplo y darles saldo inicial
	fmt.Println("\n💼 Creando cuentas de ejemplo...")

	account1, _ := wallet.CreateAccount()
	bc.AccountState.AddBalance(account1, 100.0)

	account2, _ := wallet.CreateAccount()
	bc.AccountState.AddBalance(account2, 50.0)

	account3, _ := wallet.CreateAccount()
	bc.AccountState.AddBalance(account3, 75.0)

	fmt.Println("\n💰 Saldos iniciales asignados:")
	fmt.Printf("   Cuenta 1: 100 MTC\n")
	fmt.Printf("   Cuenta 2: 50 MTC\n")
	fmt.Printf("   Cuenta 3: 75 MTC\n")

	// Menú interactivo
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║              MENÚ PRINCIPAL            ║")
		fmt.Println("╠════════════════════════════════════════╣")
		fmt.Println("║ 1. Ver cuentas en wallet               ║")
		fmt.Println("║ 2. Crear nueva cuenta                  ║")
		fmt.Println("║ 3. Ver estado de cuentas               ║")
		fmt.Println("║ 4. Crear transacción                   ║")
		fmt.Println("║ 5. Ver transacciones pendientes        ║")
		fmt.Println("║ 6. Minar bloque                        ║")
		fmt.Println("║ 7. Ver blockchain completa             ║")
		fmt.Println("║ 8. Verificar integridad                ║")
		fmt.Println("║ --- CONTRATOS INTELIGENTES ---         ║")
		fmt.Println("║ 10. Desplegar contrato (directo)       ║")
		fmt.Println("║ 11. Listar contratos                   ║")
		fmt.Println("║ 12. Ejecutar contrato (directo)        ║")
		fmt.Println("║ 13. Ver estado de contrato             ║")
		fmt.Println("║ --- TRANSACCIONES DE CONTRATOS ---     ║")
		fmt.Println("║ 14. TX: Desplegar contrato             ║")
		fmt.Println("║ 15. TX: Llamar a contrato              ║")
		fmt.Println("║ --- SALIR ---                          ║")
		fmt.Println("║ 9. Salir                               ║")
		fmt.Println("╚════════════════════════════════════════╝")
		fmt.Print("\n👉 Selecciona una opción: ")

		scanner.Scan()
		option := strings.TrimSpace(scanner.Text())

		switch option {
		case "1":
			// Ver cuentas en wallet
			wallet.ListAccounts()

		case "2":
			// Crear nueva cuenta
			address, _ := wallet.CreateAccount()
			fmt.Printf("\n✨ Cuenta creada: %s\n", address)
			fmt.Print("💰 ¿Asignar saldo inicial? (cantidad o Enter para 0): ")
			scanner.Scan()
			amountStr := strings.TrimSpace(scanner.Text())
			if amountStr != "" {
				amount, err := strconv.ParseFloat(amountStr, 64)
				if err == nil && amount > 0 {
					bc.AccountState.AddBalance(address, amount)
					fmt.Printf("✅ Saldo asignado: %.2f MTC\n", amount)
				}
			}

		case "3":
			// Ver estado de cuentas
			bc.AccountState.Print()

		case "4":
			// Crear transacción
			fmt.Println("\n💸 CREAR TRANSACCIÓN")

			// Listar cuentas
			fmt.Println("\nCuentas disponibles:")
			accounts := []string{}
			i := 1
			for address := range wallet.KeyPairs {
				fmt.Printf("%d. %s (Balance: %.2f MTC, Nonce: %d)\n",
					i, address[:16]+"...",
					bc.GetBalance(address),
					bc.GetNonce(address))
				accounts = append(accounts, address)
				i++
			}

			// Seleccionar remitente
			fmt.Print("\n👤 Número de cuenta remitente: ")
			scanner.Scan()
			fromIdx, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || fromIdx < 1 || fromIdx > len(accounts) {
				fmt.Println("❌ Cuenta inválida")
				continue
			}
			fromAddress := accounts[fromIdx-1]

			// Seleccionar destinatario
			fmt.Print("👤 Número de cuenta destinatario: ")
			scanner.Scan()
			toIdx, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || toIdx < 1 || toIdx > len(accounts) {
				fmt.Println("❌ Cuenta inválida")
				continue
			}
			toAddress := accounts[toIdx-1]

			if fromAddress == toAddress {
				fmt.Println("❌ No puedes enviar a ti mismo")
				continue
			}

			// Cantidad
			fmt.Print("💰 Cantidad a enviar: ")
			scanner.Scan()
			amount, err := strconv.ParseFloat(strings.TrimSpace(scanner.Text()), 64)
			if err != nil || amount <= 0 {
				fmt.Println("❌ Cantidad inválida")
				continue
			}

			// Obtener nonce actual
			nonce := bc.GetNonce(fromAddress)

			// Crear transacción
			tx := blockchain.NewTransaction(fromAddress, toAddress, amount, nonce)

			// Firmar transacción
			keyPair, err := wallet.GetKeyPair(fromAddress)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}

			if err := tx.Sign(keyPair); err != nil {
				fmt.Printf("❌ Error firmando: %v\n", err)
				continue
			}

			// Mostrar transacción
			tx.Print()

			// Añadir al mempool
			if err := bc.AddTransaction(tx); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}

		case "5":
			// Ver transacciones pendientes
			bc.PrintPendingTransactions()

		case "6":
			// Minar bloque
			fmt.Println("\n⛏️  MINAR BLOQUE")

			if len(bc.PendingTxs) == 0 {
				fmt.Println("❌ No hay transacciones pendientes para minar")
				continue
			}

			fmt.Printf("📊 Transacciones a incluir: %d\n", len(bc.PendingTxs))
			fmt.Print("⚠️  Esto puede tardar unos segundos. ¿Continuar? (s/n): ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) != "s" {
				continue
			}

			bc.MineBlock()
			fmt.Printf("✅ Bloque minado y añadido a la blockchain (total bloques: %d)\n", len(bc.Blocks))
		case "7":
			// Ver blockchain
			bc.Print()

		case "8":
			// Verificar integridad
			fmt.Println("\n🔍 Verificando integridad de la blockchain...")
			if bc.IsValid() {
				fmt.Println("✅ ¡Blockchain válida! Todos los bloques están intactos.")
			} else {
				fmt.Println("❌ ¡Blockchain corrupta! Se detectaron alteraciones.")
			}

		case "9":
			// Salir
			fmt.Println("\n👋 ¡Gracias por usar MiniChain!")
			return

		case "10":
			// Desplegar contrato
			fmt.Println("\n📜 DESPLEGAR CONTRATO")

			fmt.Println("\n¿Cómo quieres crear el contrato?")
			fmt.Println("1. Escribir assembly")
			fmt.Println("2. Bytecode directo")
			fmt.Print("Opción: ")

			// Crear nuevo scanner para esta sección
			var opcion string
			fmt.Scanln(&opcion)

			var bytecode []byte
			var err error

			if opcion == "1" {
				// Assembly
				fmt.Println("\nEscribe el código assembly (escribe 'FIN' para terminar):")
				fmt.Println("Ejemplo:")
				fmt.Println("  PUSH1 100")
				fmt.Println("  PUSH1 0")
				fmt.Println("  SSTORE")
				fmt.Println("  STOP")
				fmt.Println("  FIN")
				fmt.Println()

				var lines []string
				inputScanner := bufio.NewScanner(os.Stdin)

				// Leer líneas hasta "FIN"
				for inputScanner.Scan() {
					line := inputScanner.Text()
					trimmed := strings.TrimSpace(line)

					if strings.ToUpper(trimmed) == "FIN" {
						break
					}

					// Solo añadir líneas no vacías
					if trimmed != "" {
						lines = append(lines, line)
					}
				}

				if len(lines) == 0 {
					fmt.Println("❌ No se escribió ningún código")
					continue
				}

				assemblyCode := strings.Join(lines, "\n")

				// DEBUG: Mostrar el código que se va a compilar
				fmt.Println("\n📝 Código a compilar:")
				fmt.Println("─────────────────────")
				fmt.Println(assemblyCode)
				fmt.Println("─────────────────────")

				assembler := compiler.NewAssembler()
				bytecode, err = assembler.Assemble(assemblyCode)
				if err != nil {
					fmt.Printf("❌ Error compilando: %v\n", err)
					continue
				}

				// DEBUG: Mostrar bytecode generado
				fmt.Printf("\n✅ Bytecode generado: %x (%d bytes)\n", bytecode, len(bytecode))

			} else {
				// Bytecode directo
				fmt.Print("\nBytecode (hex): ")
				var hexStr string
				fmt.Scanln(&hexStr)
				hexStr = strings.TrimSpace(hexStr)
				bytecode, err = hex.DecodeString(hexStr)
				if err != nil {
					fmt.Printf("❌ Error: %v\n", err)
					continue
				}
			}

			// Seleccionar owner
			fmt.Println("\nCuentas disponibles:")
			accounts := []string{}
			i := 1
			for address := range wallet.KeyPairs {
				fmt.Printf("%d. %s\n", i, address[:16]+"...")
				accounts = append(accounts, address)
				i++
			}

			fmt.Print("\nNúmero de cuenta owner: ")
			var ownerIdxStr string
			fmt.Scanln(&ownerIdxStr)
			ownerIdx, err := strconv.Atoi(strings.TrimSpace(ownerIdxStr))
			if err != nil || ownerIdx < 1 || ownerIdx > len(accounts) {
				fmt.Println("❌ Cuenta inválida")
				continue
			}
			ownerAddress := accounts[ownerIdx-1]

			// Desplegar
			contract, err := bc.DeployContract(ownerAddress, bytecode)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}

			contract.Print()

		case "11":
			// Listar contratos
			bc.ListContracts()

		case "12":
			// Ejecutar contrato
			fmt.Println("\n⚙️  EJECUTAR CONTRATO")

			if len(bc.Contracts) == 0 {
				fmt.Println("❌ No hay contratos desplegados")
				continue
			}

			// Listar contratos
			fmt.Println("\nContratos disponibles:")
			contractAddrs := []string{}
			i := 1
			for address := range bc.Contracts {
				fmt.Printf("%d. %s\n", i, address[:16]+"...")
				contractAddrs = append(contractAddrs, address)
				i++
			}

			fmt.Print("\nNúmero de contrato: ")
			scanner.Scan()
			contractIdx, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || contractIdx < 1 || contractIdx > len(contractAddrs) {
				fmt.Println("❌ Contrato inválido")
				continue
			}
			contractAddr := contractAddrs[contractIdx-1]

			// Ejecutar con gas suficiente
			if err := bc.ExecuteContract(contractAddr, 1000000); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}

		case "13":
			// Ver estado de contrato
			fmt.Println("\n📊 ESTADO DE CONTRATO")

			if len(bc.Contracts) == 0 {
				fmt.Println("❌ No hay contratos desplegados")
				continue
			}

			// Listar contratos
			fmt.Println("\nContratos disponibles:")
			contractAddrs := []string{}
			i := 1
			for address := range bc.Contracts {
				fmt.Printf("%d. %s\n", i, address[:16]+"...")
				contractAddrs = append(contractAddrs, address)
				i++
			}

			fmt.Print("\nNúmero de contrato: ")
			scanner.Scan()
			contractIdx, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || contractIdx < 1 || contractIdx > len(contractAddrs) {
				fmt.Println("❌ Contrato inválido")
				continue
			}
			contractAddr := contractAddrs[contractIdx-1]

			contract, _ := bc.GetContract(contractAddr)
			contract.Print()

		case "14":
			// Crear transacción de despliegue de contrato
			fmt.Println("\n📜 CREAR TRANSACCIÓN DE DESPLIEGUE")

			// Seleccionar cuenta
			fmt.Println("\nCuentas disponibles:")
			accounts := []string{}
			i := 1
			for address := range wallet.KeyPairs {
				balance := bc.GetBalance(address)
				nonce := bc.GetNonce(address)
				fmt.Printf("%d. %s (Balance: %.2f MTC, Nonce: %d)\n",
					i, address[:16]+"...", balance, nonce)
				accounts = append(accounts, address)
				i++
			}

			fmt.Print("\nNúmero de cuenta: ")
			scanner.Scan()
			accountIdx, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || accountIdx < 1 || accountIdx > len(accounts) {
				fmt.Println("❌ Cuenta inválida")
				continue
			}
			fromAddress := accounts[accountIdx-1]

			// Escribir bytecode
			fmt.Println("\n¿Cómo quieres crear el contrato?")
			fmt.Println("1. Escribir assembly")
			fmt.Println("2. Bytecode directo")
			fmt.Print("Opción: ")
			scanner.Scan()

			var bytecode []byte

			if scanner.Text() == "1" {
				// Assembly
				fmt.Println("\nEscribe el código assembly (escribe 'FIN' para terminar):")
				fmt.Println("Ejemplo:")
				fmt.Println("  PUSH1 0")
				fmt.Println("  SLOAD")
				fmt.Println("  PUSH1 1")
				fmt.Println("  ADD")
				fmt.Println("  PUSH1 0")
				fmt.Println("  SSTORE")
				fmt.Println("  STOP")
				fmt.Println("  FIN")
				fmt.Println()

				var lines []string

				for scanner.Scan() {
					line := scanner.Text()
					trimmed := strings.TrimSpace(line)

					if strings.ToUpper(trimmed) == "FIN" {
						break
					}

					if trimmed != "" {
						lines = append(lines, line)
					}
				}

				if len(lines) == 0 {
					fmt.Println("❌ No se escribió ningún código")
					continue
				}

				assemblyCode := strings.Join(lines, "\n")
				assembler := compiler.NewAssembler()
				bytecode, err = assembler.Assemble(assemblyCode)
				if err != nil {
					fmt.Printf("❌ Error compilando: %v\n", err)
					continue
				}

				fmt.Printf("✅ Bytecode: %x\n", bytecode)

			} else {
				// Bytecode directo
				fmt.Print("\nBytecode (hex): ")
				scanner.Scan()
				hexStr := strings.TrimSpace(scanner.Text())
				bytecode, err = hex.DecodeString(hexStr)
				if err != nil {
					fmt.Printf("❌ Error: %v\n", err)
					continue
				}
			}

			// Crear transacción
			nonce := bc.GetNonce(fromAddress)
			tx := blockchain.NewContractDeploymentTx(fromAddress, bytecode, nonce)

			// Firmar
			keyPair, err := wallet.GetKeyPair(fromAddress)
			if err != nil {
				fmt.Printf("❌ Error obteniendo keypair: %v\n", err)
				continue
			}

			if err := tx.Sign(keyPair); err != nil {
				fmt.Printf("❌ Error firmando: %v\n", err)
				continue
			}

			// Añadir al mempool
			if err := bc.AddTransaction(tx); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}

			fmt.Println("✅ Transacción de despliegue añadida al mempool")
			fmt.Println("💡 Usa la opción 6 para minar y desplegar el contrato")

		case "15":
			// Crear transacción de llamada a contrato
			fmt.Println("\n⚙️  CREAR TRANSACCIÓN DE LLAMADA")

			if len(bc.Contracts) == 0 {
				fmt.Println("❌ No hay contratos desplegados")
				continue
			}

			// Seleccionar cuenta
			fmt.Println("\nCuentas disponibles:")
			accounts := []string{}
			i := 1
			for address := range wallet.KeyPairs {
				balance := bc.GetBalance(address)
				nonce := bc.GetNonce(address)
				fmt.Printf("%d. %s (Balance: %.2f MTC, Nonce: %d)\n",
					i, address[:16]+"...", balance, nonce)
				accounts = append(accounts, address)
				i++
			}

			fmt.Print("\nNúmero de cuenta: ")
			scanner.Scan()
			accountIdx, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || accountIdx < 1 || accountIdx > len(accounts) {
				fmt.Println("❌ Cuenta inválida")
				continue
			}
			fromAddress := accounts[accountIdx-1]

			// Seleccionar contrato
			fmt.Println("\nContratos disponibles:")
			contractAddrs := []string{}
			i = 1
			for address := range bc.Contracts {
				fmt.Printf("%d. %s\n", i, address[:16]+"...")
				contractAddrs = append(contractAddrs, address)
				i++
			}

			fmt.Print("\nNúmero de contrato: ")
			scanner.Scan()
			contractIdx, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || contractIdx < 1 || contractIdx > len(contractAddrs) {
				fmt.Println("❌ Contrato inválido")
				continue
			}
			contractAddr := contractAddrs[contractIdx-1]

			// Por ahora, calldata vacío (ejecuta todo el contrato)
			calldata := []byte{}

			// Crear transacción
			nonce := bc.GetNonce(fromAddress)
			tx := blockchain.NewContractCallTx(fromAddress, contractAddr, calldata, nonce)

			// Firmar
			keyPair, err := wallet.GetKeyPair(fromAddress)
			if err != nil {
				fmt.Printf("❌ Error obteniendo keypair: %v\n", err)
				continue
			}
			if err := tx.Sign(keyPair); err != nil {
				fmt.Printf("❌ Error firmando: %v\n", err)
				continue
			}

			// Añadir al mempool
			if err := bc.AddTransaction(tx); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}

			fmt.Println("✅ Transacción de llamada añadida al mempool")
			fmt.Println("💡 Usa la opción 6 para minar y ejecutar el contrato")

		default:
			fmt.Println("\n❌ Opción inválida")
		}
	}
}
