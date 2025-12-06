package blockchain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"minichain/crypto"
)

// Transaction representa una transferencia de fondos
type Transaction struct {
	From      string  // Dirección del remitente
	To        string  // Dirección del destinatario
	Amount    float64 // Cantidad a transferir
	Nonce     int     // Contador de transacciones del remitente
	Signature string  // Firma digital del remitente

	// Coordenadas de la clave pública (para verificar firma)
	PublicKeyX *big.Int
	PublicKeyY *big.Int
}

// NewTransaction crea una nueva transacción (sin firmar)
func NewTransaction(from, to string, amount float64, nonce int) *Transaction {
	return &Transaction{
		From:   from,
		To:     to,
		Amount: amount,
		Nonce:  nonce,
	}
}

// Sign firma la transacción con un par de claves
func (tx *Transaction) Sign(keyPair *crypto.KeyPair) error {
	// Verificar que la dirección coincide con el par de claves
	if tx.From != keyPair.GetAddress() {
		return fmt.Errorf("la dirección From no coincide con el par de claves")
	}

	// Guardar la clave pública (necesaria para verificar la firma)
	tx.PublicKeyX = keyPair.PublicKey.X
	tx.PublicKeyY = keyPair.PublicKey.Y

	// Crear los datos a firmar (sin la firma misma)
	dataToSign := tx.getDataToSign()

	// Firmar los datos
	signature, err := keyPair.SignData(dataToSign)
	if err != nil {
		return fmt.Errorf("error firmando transacción: %v", err)
	}

	tx.Signature = signature

	return nil
}

// getDataToSign obtiene los datos que se firman
// No incluye la firma misma (obvio, no puedes firmar la firma)
func (tx *Transaction) getDataToSign() []byte {
	data := fmt.Sprintf("%s:%s:%.2f:%d", tx.From, tx.To, tx.Amount, tx.Nonce)
	return []byte(data)
}

// VerifySignature verifica que la firma sea válida
func (tx *Transaction) VerifySignature() bool {
	if tx.Signature == "" {
		return false
	}

	if tx.PublicKeyX == nil || tx.PublicKeyY == nil {
		return false
	}

	// Obtener los datos que fueron firmados
	dataToSign := tx.getDataToSign()

	// Verificar la firma
	return crypto.VerifySignature(tx.PublicKeyX, tx.PublicKeyY, dataToSign, tx.Signature)
}

// Validate verifica que la transacción sea válida
func (tx *Transaction) Validate(state *AccountState) error {
	// 1. Verificar que tenga firma
	if tx.Signature == "" {
		return fmt.Errorf("transacción sin firmar")
	}

	// 2. Verificar que la firma sea válida
	if !tx.VerifySignature() {
		return fmt.Errorf("firma inválida")
	}

	// 3. Verificar que el monto sea positivo
	if tx.Amount <= 0 {
		return fmt.Errorf("monto debe ser positivo: %.2f", tx.Amount)
	}

	// 4. Obtener la cuenta del remitente
	account := state.GetAccount(tx.From)

	// 5. Verificar el nonce (debe ser EXACTAMENTE el siguiente)
	if tx.Nonce != account.Nonce {
		return fmt.Errorf("nonce inválido: esperado %d, recibido %d", account.Nonce, tx.Nonce)
	}

	// 6. Verificar que tenga saldo suficiente
	if account.Balance < tx.Amount {
		return fmt.Errorf("saldo insuficiente: tiene %.2f, necesita %.2f", account.Balance, tx.Amount)
	}

	return nil
}

// Execute ejecuta la transacción (transfiere los fondos)
func (tx *Transaction) Execute(state *AccountState) error {
	// Validar antes de ejecutar
	if err := tx.Validate(state); err != nil {
		return err
	}

	// Restar del remitente
	if err := state.SubtractBalance(tx.From, tx.Amount); err != nil {
		return err
	}

	// Sumar al destinatario
	state.AddBalance(tx.To, tx.Amount)

	// Incrementar el nonce del remitente
	state.IncrementNonce(tx.From)

	fmt.Printf("✅ Transacción ejecutada: %.2f MTC de %s a %s\n",
		tx.Amount,
		tx.From[:8]+"...",
		tx.To[:8]+"...")

	return nil
}

// ToJSON convierte la transacción a JSON
func (tx *Transaction) ToJSON() (string, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Print muestra la transacción de forma bonita
func (tx *Transaction) Print() {
	fmt.Println("\n┌────────────────────────────────────────┐")
	fmt.Println("│          💸 TRANSACCIÓN                │")
	fmt.Println("└────────────────────────────────────────┘")
	fmt.Printf("📤 From:      %s\n", tx.From[:16]+"...")
	fmt.Printf("📥 To:        %s\n", tx.To[:16]+"...")
	fmt.Printf("💰 Amount:    %.2f MTC\n", tx.Amount)
	fmt.Printf("🔢 Nonce:     %d\n", tx.Nonce)

	if tx.Signature != "" {
		fmt.Printf("✍️  Signature: %s...\n", tx.Signature[:16])
		fmt.Printf("✅ Firmada:   Sí\n")
		if tx.VerifySignature() {
			fmt.Printf("🔐 Válida:    Sí\n")
		} else {
			fmt.Printf("❌ Válida:    No\n")
		}
	} else {
		fmt.Printf("⚠️  Signature: (sin firmar)\n")
	}
}
