package blockchain

import (
	"fmt"
)

// Account representa una cuenta con saldo
type Account struct {
	Address string  // Dirección de la cuenta
	Balance float64 // Saldo en la cuenta
	Nonce   int     // Contador de transacciones (previene replay attacks)
}

// AccountState mantiene el estado global de todas las cuentas
type AccountState struct {
	Accounts map[string]*Account // address -> Account
}

// NewAccountState crea un nuevo estado de cuentas vacío
func NewAccountState() *AccountState {
	return &AccountState{
		Accounts: make(map[string]*Account),
	}
}

// GetAccount obtiene una cuenta (la crea si no existe)
func (as *AccountState) GetAccount(address string) *Account {
	account, exists := as.Accounts[address]
	if !exists {
		// Crear cuenta nueva con saldo 0
		account = &Account{
			Address: address,
			Balance: 0,
			Nonce:   0,
		}
		as.Accounts[address] = account
	}
	return account
}

// GetBalance obtiene el saldo de una cuenta
func (as *AccountState) GetBalance(address string) float64 {
	return as.GetAccount(address).Balance
}

// AddBalance añade saldo a una cuenta
func (as *AccountState) AddBalance(address string, amount float64) {
	account := as.GetAccount(address)
	account.Balance += amount
}

// SubtractBalance resta saldo de una cuenta
func (as *AccountState) SubtractBalance(address string, amount float64) error {
	account := as.GetAccount(address)
	if account.Balance < amount {
		return fmt.Errorf("saldo insuficiente: tiene %.2f, necesita %.2f", account.Balance, amount)
	}
	account.Balance -= amount
	return nil
}

// IncrementNonce incrementa el nonce de una cuenta
func (as *AccountState) IncrementNonce(address string) {
	account := as.GetAccount(address)
	account.Nonce++
}

// StateSnapshot guarda un snapshot del estado de cuentas
type StateSnapshot struct {
	Accounts map[string]*Account
}

// CreateSnapshot crea un snapshot del estado actual
func (as *AccountState) CreateSnapshot() *StateSnapshot {
	snapshot := &StateSnapshot{
		Accounts: make(map[string]*Account),
	}

	// Copiar todas las cuentas
	for address, account := range as.Accounts {
		snapshot.Accounts[address] = &Account{
			Address: account.Address,
			Balance: account.Balance,
			Nonce:   account.Nonce,
		}
	}

	return snapshot
}

// RevertToSnapshot revierte el estado a un snapshot
func (as *AccountState) RevertToSnapshot(snapshot *StateSnapshot) {
	// Restaurar cuentas
	for address, account := range snapshot.Accounts {
		as.Accounts[address] = &Account{
			Address: account.Address,
			Balance: account.Balance,
			Nonce:   account.Nonce,
		}
	}
}

// Print muestra el estado de todas las cuentas
func (as *AccountState) Print() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║         ESTADO DE CUENTAS              ║")
	fmt.Println("╚════════════════════════════════════════╝")

	if len(as.Accounts) == 0 {
		fmt.Println("   (No hay cuentas)")
		return
	}

	for address, account := range as.Accounts {
		fmt.Printf("\n📍 %s\n", address)
		fmt.Printf("   💰 Saldo: %.2f MTC\n", account.Balance)
		fmt.Printf("   🔢 Nonce: %d\n", account.Nonce)
	}
}
