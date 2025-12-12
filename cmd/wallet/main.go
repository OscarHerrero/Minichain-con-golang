package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"minichain/crypto"
	"os"
	"path/filepath"
)

// WalletFile representa el formato de archivo de wallet
type WalletFile struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

func main() {
	// Parsear argumentos
	output := flag.String("output", "", "Archivo donde guardar la wallet (ej: alice.json)")
	load := flag.String("load", "", "Cargar wallet existente desde archivo")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso: %s [opciones]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Gestión de wallets para Minichain\n\n")
		fmt.Fprintf(os.Stderr, "Opciones:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEjemplos:\n")
		fmt.Fprintf(os.Stderr, "  # Generar nueva wallet\n")
		fmt.Fprintf(os.Stderr, "  %s --output alice.json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Ver wallet existente\n")
		fmt.Fprintf(os.Stderr, "  %s --load alice.json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generar wallet sin guardar\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
	}

	flag.Parse()

	if *load != "" {
		// Cargar wallet existente
		loadWallet(*load)
	} else {
		// Generar nueva wallet
		generateWallet(*output)
	}
}

func generateWallet(outputFile string) {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              🔐 GENERADOR DE WALLETS - MINICHAIN          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Generar par de claves ECDSA
	fmt.Println("🔑 Generando par de claves ECDSA...")
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Printf("❌ Error generando par de claves: %v\n", err)
		os.Exit(1)
	}

	// Obtener dirección
	address := keyPair.GetAddress()

	// Obtener clave privada en formato hex
	privateKeyHex := keyPair.GetPrivateKeyHex()

	// Mostrar información
	fmt.Println()
	fmt.Println("✅ Wallet generada exitosamente!")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📍 DIRECCIÓN (para recibir fondos):")
	fmt.Println()
	fmt.Printf("   %s\n", address)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔒 CLAVE PRIVADA (mantén esto en secreto):")
	fmt.Println()
	fmt.Printf("   %s\n", privateKeyHex)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Guardar en archivo si se especificó
	if outputFile != "" {
		wallet := WalletFile{
			Address:    address,
			PrivateKey: privateKeyHex,
		}

		jsonData, err := json.MarshalIndent(wallet, "", "  ")
		if err != nil {
			fmt.Printf("❌ Error serializando wallet: %v\n", err)
			os.Exit(1)
		}

		// Crear directorio si no existe
		dir := filepath.Dir(outputFile)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("❌ Error creando directorio: %v\n", err)
				os.Exit(1)
			}
		}

		// Guardar archivo
		if err := ioutil.WriteFile(outputFile, jsonData, 0600); err != nil {
			fmt.Printf("❌ Error guardando wallet: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("💾 Wallet guardada en: %s\n", outputFile)
		fmt.Println()
	}

	// Advertencias de seguridad
	fmt.Println("⚠️  IMPORTANTE - SEGURIDAD:")
	fmt.Println("   • NUNCA compartas tu clave privada")
	fmt.Println("   • Guarda tu clave privada en lugar seguro")
	fmt.Println("   • Si pierdes tu clave privada, pierdes acceso a tus fondos")
	fmt.Println("   • Usa esta wallet solo para testing/desarrollo")
	fmt.Println()
}

func loadWallet(filename string) {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              🔍 CARGAR WALLET - MINICHAIN                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Leer archivo
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("❌ Error leyendo archivo: %v\n", err)
		os.Exit(1)
	}

	// Deserializar
	var wallet WalletFile
	if err := json.Unmarshal(data, &wallet); err != nil {
		fmt.Printf("❌ Error parseando wallet: %v\n", err)
		os.Exit(1)
	}

	// Mostrar información
	fmt.Printf("📁 Archivo: %s\n", filename)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📍 DIRECCIÓN:")
	fmt.Println()
	fmt.Printf("   %s\n", wallet.Address)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔒 CLAVE PRIVADA:")
	fmt.Println()
	fmt.Printf("   %s\n", wallet.PrivateKey)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}
