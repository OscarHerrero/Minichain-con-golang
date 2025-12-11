package memorydb

import (
	"bytes"
	"errors"
	"minichain/database"
	"sort"
	"sync"
)

// Database es una base de datos en memoria para testing
// Basada en ethdb/memorydb de go-ethereum
type Database struct {
	db   map[string][]byte
	lock sync.RWMutex
}

// New crea una nueva base de datos en memoria
func New() *Database {
	return &Database{
		db: make(map[string][]byte),
	}
}

// NewWithCap crea una base de datos con capacidad inicial
func NewWithCap(size int) *Database {
	return &Database{
		db: make(map[string][]byte, size),
	}
}

// Close cierra la base de datos (no hace nada en memoria)
func (db *Database) Close() error {
	return nil
}

// Has verifica si una key existe
func (db *Database) Has(key []byte) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	_, exists := db.db[string(key)]
	return exists, nil
}

// Get obtiene el valor de una key
func (db *Database) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if entry, exists := db.db[string(key)]; exists {
		// Retornar copia para evitar modificaciones
		result := make([]byte, len(entry))
		copy(result, entry)
		return result, nil
	}
	return nil, errors.New("not found")
}

// Put inserta o actualiza una key
func (db *Database) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	// Crear copia del valor
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	db.db[string(key)] = valueCopy
	return nil
}

// Delete elimina una key
func (db *Database) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	delete(db.db, string(key))
	return nil
}

// NewBatch crea un nuevo batch
func (db *Database) NewBatch() database.Batch {
	return &batch{
		db: db,
	}
}

// NewBatchWithSize crea un batch con capacidad inicial
func (db *Database) NewBatchWithSize(size int) database.Batch {
	return &batch{
		db: db,
	}
}

// NewIterator crea un iterador
func (db *Database) NewIterator(prefix []byte, start []byte) database.Iterator {
	db.lock.RLock()
	defer db.lock.RUnlock()

	// Construir lista de keys que cumplen con el filtro
	var keys []string
	for key := range db.db {
		if bytes.HasPrefix([]byte(key), prefix) {
			if start == nil || bytes.Compare([]byte(key), append(prefix, start...)) >= 0 {
				keys = append(keys, key)
			}
		}
	}

	// Ordenar keys
	sort.Strings(keys)

	// Copiar valores
	values := make(map[string][]byte, len(keys))
	for _, key := range keys {
		valueCopy := make([]byte, len(db.db[key]))
		copy(valueCopy, db.db[key])
		values[key] = valueCopy
	}

	return &iterator{
		keys:   keys,
		values: values,
		index:  -1,
	}
}

// Stat retorna estadísticas (no implementado en memoria)
func (db *Database) Stat(property string) (string, error) {
	return "", errors.New("not supported")
}

// Compact no hace nada en memoria
func (db *Database) Compact(start []byte, limit []byte) error {
	return nil
}

// Len retorna el número de elementos en la base de datos
func (db *Database) Len() int {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return len(db.db)
}

// batch implementa database.Batch para base de datos en memoria
type batch struct {
	db     *Database
	writes []keyvalue
	size   int
}

type keyvalue struct {
	key    []byte
	value  []byte
	delete bool
}

func (b *batch) Put(key, value []byte) error {
	// Copiar key y value
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	b.writes = append(b.writes, keyvalue{keyCopy, valueCopy, false})
	b.size += len(key) + len(value)
	return nil
}

func (b *batch) Delete(key []byte) error {
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	b.writes = append(b.writes, keyvalue{keyCopy, nil, true})
	b.size += len(key)
	return nil
}

func (b *batch) ValueSize() int {
	return b.size
}

func (b *batch) Write() error {
	b.db.lock.Lock()
	defer b.db.lock.Unlock()

	for _, kv := range b.writes {
		if kv.delete {
			delete(b.db.db, string(kv.key))
		} else {
			b.db.db[string(kv.key)] = kv.value
		}
	}
	return nil
}

func (b *batch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}

func (b *batch) Replay(w database.KeyValueWriter) error {
	for _, kv := range b.writes {
		if kv.delete {
			if err := w.Delete(kv.key); err != nil {
				return err
			}
		} else {
			if err := w.Put(kv.key, kv.value); err != nil {
				return err
			}
		}
	}
	return nil
}

// iterator implementa database.Iterator para base de datos en memoria
type iterator struct {
	keys   []string
	values map[string][]byte
	index  int
}

func (it *iterator) Next() bool {
	it.index++
	return it.index < len(it.keys)
}

func (it *iterator) Error() error {
	return nil
}

func (it *iterator) Key() []byte {
	if it.index < 0 || it.index >= len(it.keys) {
		return nil
	}
	return []byte(it.keys[it.index])
}

func (it *iterator) Value() []byte {
	if it.index < 0 || it.index >= len(it.keys) {
		return nil
	}
	return it.values[it.keys[it.index]]
}

func (it *iterator) Release() {
	it.keys = nil
	it.values = nil
	it.index = -1
}
