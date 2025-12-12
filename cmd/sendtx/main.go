package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"minichain/crypto"
	"net/http"
	"os"
	"strings"
)

// Transaction representa una transacción para enviar
type Transaction struct {
	From       string   `json:"from"`
	To         string   `json:"to"`
	Amount     float64  `json:"amount"`
	Nonce      int      `json:"nonce"`
	Data       string   `json:"data"` // Hex string opcional
	Signature  string   `json:"signature"`
	PublicKeyX *big.Int `json:"publicKeyX"`
	PublicKeyY *big.Int `json:"publicKeyY"`
}

// WalletFile representa el formato de archivo de wallet
type WalletFile struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

func main() {
	// Parsear argumentos
	to := flag.String("to", "", "Dirección destino (requerido)")
	amount := flag.Float64("amount", 0, "Cantidad a enviar")
	data := flag.String("data", "", "Data en hex (opcional)")
	privateKey := flag.String("key", "", "Clave privada en hex")
	walletFile := flag.String("wallet", "", "Archivo de wallet (ej: alice.json)")
	rpcURL := flag.String("rpc", "http://localhost:8545", "URL del RPC del nodo")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso: %s [opciones]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Envía una transacción firmada a un nodo de Minichain\n\n")
		fmt.Fprintf(os.Stderr, "Opciones:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEjemplos:\n")
		fmt.Fprintf(os.Stderr, "  # Con wallet file\n")
		fmt.Fprintf(os.Stderr, "  %s --wallet alice.json --to <dirección> --amount 10\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Con clave privada directa\n")
		fmt.Fprintf(os.Stderr, "  %s --key <hex> --to <dirección> --amount 10\n\n", os.Args[0])
	}

	flag.Parse()

	// Validar que se proporcionó wallet o key
	if *privateKey == "" && *walletFile == "" {
		fmt.Fprintln(os.Stderr, "❌ Error: Debes proporcionar --wallet o --key")
		flag.Usage()
		os.Exit(1)
	}

	// Validar destino
	if *to == "" && *data == "" {
		fmt.Fprintln(os.Stderr, "❌ Error: --to es requerido")
		flag.Usage()
		os.Exit(1)
	}

	// Cargar KeyPair
	var keyPair *crypto.KeyPair
	var err error

	if *walletFile != "" {
		// Cargar desde archivo
		keyPair, err = loadWalletFile(*walletFile)
		if err != nil {
			fmt.Printf("❌ Error cargando wallet: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Cargar desde clave privada
		keyPair, err = crypto.LoadFromPrivateKeyHex(*privateKey)
		if err != nil {
			fmt.Printf("❌ Error cargando clave privada: %v\n", err)
			os.Exit(1)
		}
	}

	from := keyPair.GetAddress()

	// Crear transacción
	tx := Transaction{
		From:       from,
		To:         *to,
		Amount:     *amount,
		Nonce:      0, // TODO: Obtener nonce actual del servidor
		Data:       *data,
		PublicKeyX: keyPair.PublicKey.X,
		PublicKeyY: keyPair.PublicKey.Y,
	}

	// Firmar transacción
	txData := fmt.Sprintf("%s%s%.2f%d%s", tx.From, tx.To, tx.Amount, tx.Nonce, tx.Data)
	signature, err := keyPair.SignData([]byte(txData))
	if err != nil {
		fmt.Printf("❌ Error firmando transacción: %v\n", err)
		os.Exit(1)
	}
	tx.Signature = signature

	// Mostrar información
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           📤 ENVIANDO TRANSACCIÓN A MINICHAIN             ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("From:      %s\n", tx.From)

	if tx.To == "" {
		fmt.Println("To:        (DESPLIEGUE DE CONTRATO)")
	} else {
		fmt.Printf("To:        %s\n", tx.To)
	}

	fmt.Printf("Amount:    %.2f MTC\n", tx.Amount)
	fmt.Printf("Nonce:     %d\n", tx.Nonce)

	if tx.Data != "" {
		dataLen := len(strings.TrimPrefix(tx.Data, "0x"))
		fmt.Printf("Data:      %d bytes\n", dataLen/2)
	}

	fmt.Printf("Signature: %s...\n", tx.Signature[:32])
	fmt.Println()

	// Serializar a JSON
	jsonData, err := json.Marshal(tx)
	if err != nil {
		fmt.Printf("❌ Error serializando transacción: %v\n", err)
		os.Exit(1)
	}

	// Enviar al nodo
	fmt.Printf("🔄 Enviando a %s/tx ...\n", *rpcURL)

	resp, err := http.Post(*rpcURL+"/tx", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		fmt.Printf("❌ Error enviando transacción: %v\n", err)
		fmt.Println()
		fmt.Println("💡 Asegúrate de que el nodo esté corriendo con RPC habilitado")
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Error leyendo respuesta: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode == 200 {
		fmt.Println("✅ Transacción enviada exitosamente!")
		fmt.Printf("   %s\n", string(body))
	} else {
		fmt.Printf("❌ Error del servidor (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	fmt.Println()
}

func loadWalletFile(filename string) (*crypto.KeyPair, error) {
	// Leer archivo
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %v", err)
	}

	// Deserializar
	var wallet WalletFile
	if err := json.Unmarshal(data, &wallet); err != nil {
		return nil, fmt.Errorf("error parseando wallet: %v", err)
	}

	// Cargar KeyPair desde clave privada
	return crypto.LoadFromPrivateKeyHex(wallet.PrivateKey)
}
