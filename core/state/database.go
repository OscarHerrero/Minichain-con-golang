package state

import (
	"minichain/database"
	"minichain/trie"
)

// Database es la interfaz para acceder al state database
// Basado en go-ethereum/core/state/database.go
type Database interface {
	// TrieDB retorna la trie database
	TrieDB() *trie.Database

	// ContractCode obtiene el código de un contrato por su hash
	ContractCode(codeHash []byte) ([]byte, error)

	// ContractCodeWrite guarda el código de un contrato
	ContractCodeWrite(codeHash []byte, code []byte) error
}

// cachingDB implementa Database usando una base de datos clave-valor
type cachingDB struct {
	db      database.Database // Base de datos backing
	trieDB  *trie.Database    // Trie database
}

// NewDatabase crea una nueva state database
func NewDatabase(db database.Database) Database {
	return &cachingDB{
		db:     db,
		trieDB: trie.NewDatabase(db),
	}
}

// TrieDB retorna la trie database
func (db *cachingDB) TrieDB() *trie.Database {
	return db.trieDB
}

// ContractCode obtiene el código de un contrato
func (db *cachingDB) ContractCode(codeHash []byte) ([]byte, error) {
	// Prefijo 'c' para contract code (como en Geth)
	key := append([]byte("c"), codeHash...)
	return db.db.Get(key)
}

// ContractCodeWrite guarda el código de un contrato
func (db *cachingDB) ContractCodeWrite(codeHash []byte, code []byte) error {
	// Prefijo 'c' para contract code
	key := append([]byte("c"), codeHash...)
	return db.db.Put(key, code)
}
