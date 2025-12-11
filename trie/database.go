package trie

import (
	"minichain/database"
	"sync"
)

// Database es un intermediario entre el trie y la base de datos de almacenamiento
// Provee caché y batch writes para eficiencia
// Basado en go-ethereum/trie/database.go
type Database struct {
	db database.Database // Base de datos backing

	// Caché de nodos en memoria
	nodes map[string][]byte
	lock  sync.RWMutex

	// Estadísticas
	nodesSize int // Tamaño total de nodos en caché
}

// Config contiene la configuración para la trie database
type Config struct {
	Cache int // Tamaño de caché en MB (0 = sin límite)
}

// NewDatabase crea una nueva trie database
func NewDatabase(db database.Database) *Database {
	return &Database{
		db:    db,
		nodes: make(map[string][]byte),
	}
}

// Node obtiene un nodo codificado por su hash
func (db *Database) Node(hash []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	// Buscar en caché primero
	if enc, ok := db.nodes[string(hash)]; ok {
		return enc, nil
	}

	// Si no está en caché, cargar desde disco
	return db.db.Get(hash)
}

// Insert inserta un nodo en la caché
func (db *Database) Insert(hash []byte, blob []byte) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.nodes[string(hash)] = blob
	db.nodesSize += len(blob)
}

// Commit escribe todos los nodos en caché a disco
func (db *Database) Commit() error {
	db.lock.Lock()
	defer db.lock.Unlock()

	batch := db.db.NewBatch()

	// Escribir todos los nodos del caché
	for hash, blob := range db.nodes {
		if err := batch.Put([]byte(hash), blob); err != nil {
			return err
		}
	}

	// Ejecutar batch
	if err := batch.Write(); err != nil {
		return err
	}

	// Limpiar caché después de commit
	db.nodes = make(map[string][]byte)
	db.nodesSize = 0

	return nil
}

// Size retorna el tamaño del caché en bytes
func (db *Database) Size() int {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return db.nodesSize
}

// Cap limita el tamaño del caché
// Remueve nodos más viejos si excede el límite
func (db *Database) Cap(limit int) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.nodesSize <= limit {
		return nil
	}

	// Estrategia simple: limpiar todo si excede límite
	// En producción, se usaría LRU u otra política
	batch := db.db.NewBatch()

	for hash, blob := range db.nodes {
		if err := batch.Put([]byte(hash), blob); err != nil {
			return err
		}
	}

	if err := batch.Write(); err != nil {
		return err
	}

	db.nodes = make(map[string][]byte)
	db.nodesSize = 0

	return nil
}

// Reference NO hace nada en nuestra implementación simplificada
// En Geth, esto maneja reference counting para garbage collection
func (db *Database) Reference(child []byte, parent []byte) {}

// Dereference NO hace nada en nuestra implementación simplificada
func (db *Database) Dereference(root []byte) {}
