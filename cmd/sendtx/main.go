package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Transaction representa una transacción para enviar
type Transaction struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Data   string  `json:"data"` // Hex string opcional
}

func main() {
	// Parsear argumentos
	from := flag.String("from", "", "Dirección origen (requerido)")
	to := flag.String("to", "", "Dirección destino (requerido)")
	amount := flag.Float64("amount", 0, "Cantidad a enviar")
	data := flag.String("data", "", "Data en hex (opcional)")
	rpcURL := flag.String("rpc", "http://localhost:8545", "URL del RPC del nodo")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso: %s [opciones]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Envía una transacción a un nodo de Minichain\n\n")
		fmt.Fprintf(os.Stderr, "Opciones:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEjemplos:\n")
		fmt.Fprintf(os.Stderr, "  # Transferencia simple\n")
		fmt.Fprintf(os.Stderr, "  %s --from Alice --to Bob --amount 10\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Despliegue de contrato\n")
		fmt.Fprintf(os.Stderr, "  %s --from Alice --to \"\" --data 608060405234...\n\n", os.Args[0])
	}

	flag.Parse()

	// Validar argumentos requeridos
	if *from == "" {
		fmt.Fprintln(os.Stderr, "❌ Error: --from es requerido")
		flag.Usage()
		os.Exit(1)
	}

	if *to == "" && *data == "" {
		fmt.Fprintln(os.Stderr, "❌ Error: --to es requerido (usa \"\" para despliegue de contrato)")
		flag.Usage()
		os.Exit(1)
	}

	// Crear transacción
	tx := Transaction{
		From:   *from,
		To:     *to,
		Amount: *amount,
		Data:   *data,
	}

	// Mostrar información
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           📤 ENVIANDO TRANSACCIÓN A MINICHAIN             ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("From:   %s\n", tx.From)

	if tx.To == "" {
		fmt.Println("To:     (DESPLIEGUE DE CONTRATO)")
	} else {
		fmt.Printf("To:     %s\n", tx.To)
	}

	fmt.Printf("Amount: %.2f MTC\n", tx.Amount)

	if tx.Data != "" {
		dataLen := len(strings.TrimPrefix(tx.Data, "0x"))
		fmt.Printf("Data:   %d bytes\n", dataLen/2)
	}

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
