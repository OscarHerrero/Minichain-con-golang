package crypto

import (
	"fmt"
)

// Wallet gestiona múltiples pares de claves
type Wallet struct {
	KeyPairs map[string]*KeyPair // address -> KeyPair
}

// NewWallet crea una nueva wallet vacía
func NewWallet() *Wallet {
	return &Wallet{
		KeyPairs: make(map[string]*KeyPair),
	}
}

// CreateAccount crea una nueva cuenta (par de claves)
func (w *Wallet) CreateAccount() (string, error) {
	// Generar nuevo par de claves
	keyPair, err := GenerateKeyPair()
	if err != nil {
		return "", err
	}
	
	// Obtener la dirección
	address := keyPair.GetAddress()
	
	// Guardar en la wallet
	w.KeyPairs[address] = keyPair
	
	fmt.Printf("\n✨ Nueva cuenta creada: %s\n", address)
	
	return address, nil
}

// GetKeyPair obtiene el par de claves de una dirección
func (w *Wallet) GetKeyPair(address string) (*KeyPair, error) {
	keyPair, exists := w.KeyPairs[address]
	if !exists {
		return nil, fmt.Errorf("dirección no encontrada en la wallet")
	}
	return keyPair, nil
}

// ListAccounts muestra todas las cuentas de la wallet
func (w *Wallet) ListAccounts() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║           CUENTAS EN WALLET            ║")
	fmt.Println("╚════════════════════════════════════════╝")
	
	if len(w.KeyPairs) == 0 {
		fmt.Println("   (No hay cuentas)")
		return
	}
	
	i := 1
	for address := range w.KeyPairs {
		fmt.Printf("%d. %s\n", i, address)
		i++
	}
}