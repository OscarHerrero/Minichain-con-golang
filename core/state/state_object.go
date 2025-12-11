package state

import (
	"bytes"
	"math/big"
	"minichain/trie"
)

// Account representa el estado de una cuenta en Ethereum
// Basado en go-ethereum/core/state/state_object.go
type Account struct {
	Nonce    uint64   // Número de transacciones enviadas
	Balance  *big.Int // Saldo en wei
	Root     []byte   // Root del storage trie (para contratos)
	CodeHash []byte   // Hash del código del contrato
}

// NewAccount crea una nueva cuenta vacía
func NewAccount() *Account {
	return &Account{
		Nonce:    0,
		Balance:  new(big.Int),
		Root:     trie.Keccak256(nil), // Empty trie root
		CodeHash: trie.Keccak256(nil), // Empty code hash
	}
}

// Copy crea una copia profunda de la cuenta
func (a *Account) Copy() *Account {
	cpy := &Account{
		Nonce: a.Nonce,
		Balance: new(big.Int),
	}
	if a.Balance != nil {
		cpy.Balance.Set(a.Balance)
	}
	if a.Root != nil {
		cpy.Root = make([]byte, len(a.Root))
		copy(cpy.Root, a.Root)
	}
	if a.CodeHash != nil {
		cpy.CodeHash = make([]byte, len(a.CodeHash))
		copy(cpy.CodeHash, a.CodeHash)
	}
	return cpy
}

// stateObject representa el estado mutable de una cuenta
// Wrappea Account y añade funcionalidad de tracking de cambios
type stateObject struct {
	address []byte   // Dirección de la cuenta
	data    Account  // Datos de la cuenta
	db      *StateDB // Referencia al StateDB

	// Storage trie para contratos
	storageTrie *trie.SecureTrie

	// Code del contrato
	code []byte

	// Tracking de cambios
	dirtyStorage map[string][]byte // Cambios pendientes en storage

	// Flags
	dirtyCode bool // Si el código cambió
	suicided  bool // Si la cuenta se autodestruyó
	deleted   bool // Si la cuenta fue eliminada
}

// newObject crea un nuevo state object
func newObject(db *StateDB, address []byte, data Account) *stateObject {
	return &stateObject{
		address:      address,
		data:         data,
		db:           db,
		dirtyStorage: make(map[string][]byte),
	}
}

// Address retorna la dirección de la cuenta
func (s *stateObject) Address() []byte {
	return s.address
}

// Balance retorna el saldo de la cuenta
func (s *stateObject) Balance() *big.Int {
	return s.data.Balance
}

// SetBalance establece el saldo de la cuenta
func (s *stateObject) SetBalance(amount *big.Int) {
	s.data.Balance = new(big.Int).Set(amount)
}

// AddBalance añade al saldo de la cuenta
func (s *stateObject) AddBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	s.SetBalance(new(big.Int).Add(s.Balance(), amount))
}

// SubBalance resta del saldo de la cuenta
func (s *stateObject) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	s.SetBalance(new(big.Int).Sub(s.Balance(), amount))
}

// Nonce retorna el nonce de la cuenta
func (s *stateObject) Nonce() uint64 {
	return s.data.Nonce
}

// SetNonce establece el nonce de la cuenta
func (s *stateObject) SetNonce(nonce uint64) {
	s.data.Nonce = nonce
}

// Code retorna el código del contrato
func (s *stateObject) Code() []byte {
	if s.code != nil {
		return s.code
	}

	// Si no está cargado y hay un codeHash, cargar desde DB
	if !bytes.Equal(s.data.CodeHash, trie.Keccak256(nil)) {
		code, err := s.db.db.ContractCode(s.data.CodeHash)
		if err == nil {
			s.code = code
		}
	}

	return s.code
}

// SetCode establece el código del contrato
func (s *stateObject) SetCode(code []byte) {
	s.code = code
	s.data.CodeHash = trie.Keccak256(code)
	s.dirtyCode = true
}

// GetState retorna un valor del storage
func (s *stateObject) GetState(key []byte) []byte {
	// Primero buscar en dirty storage
	if value, ok := s.dirtyStorage[string(key)]; ok {
		return value
	}

	// Cargar storage trie si es necesario
	if s.storageTrie == nil {
		var err error
		s.storageTrie, err = trie.NewSecure(s.data.Root, s.db.db.TrieDB())
		if err != nil {
			return nil
		}
	}

	// Buscar en el trie
	return s.storageTrie.Get(key)
}

// SetState establece un valor en el storage
func (s *stateObject) SetState(key, value []byte) {
	s.dirtyStorage[string(key)] = value
}

// updateStorageTrie escribe los cambios de storage al trie
func (s *stateObject) updateStorageTrie() error {
	// Cargar storage trie si es necesario
	if s.storageTrie == nil {
		var err error
		s.storageTrie, err = trie.NewSecure(s.data.Root, s.db.db.TrieDB())
		if err != nil {
			return err
		}
	}

	// Escribir cambios dirty al trie
	for key, value := range s.dirtyStorage {
		if len(value) == 0 {
			s.storageTrie.Delete([]byte(key))
		} else {
			s.storageTrie.Update([]byte(key), value)
		}
	}

	// Limpiar dirty storage
	s.dirtyStorage = make(map[string][]byte)

	// Actualizar storage root
	s.data.Root = s.storageTrie.Hash()

	return nil
}

// commit escribe el state object a la base de datos
func (s *stateObject) commit() error {
	// Actualizar storage trie si hay cambios
	if len(s.dirtyStorage) > 0 {
		if err := s.updateStorageTrie(); err != nil {
			return err
		}

		// Commit del storage trie
		if s.storageTrie != nil {
			_, err := s.storageTrie.Commit()
			if err != nil {
				return err
			}
		}
	}

	// Guardar código si cambió
	if s.dirtyCode && s.code != nil {
		s.db.db.ContractCodeWrite(s.data.CodeHash, s.code)
		s.dirtyCode = false
	}

	return nil
}

// empty retorna si la cuenta está vacía
func (s *stateObject) empty() bool {
	return s.data.Nonce == 0 &&
		s.data.Balance.Sign() == 0 &&
		bytes.Equal(s.data.CodeHash, trie.Keccak256(nil))
}
