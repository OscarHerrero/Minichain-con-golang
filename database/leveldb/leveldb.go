package leveldb

import (
	"fmt"
	"sync"

	"minichain/database"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Database es un wrapper de LevelDB que implementa database.Database
// Basado en ethdb/leveldb de go-ethereum
type Database struct {
	fn string      // Ruta del archivo de la base de datos
	db *leveldb.DB // Instancia de LevelDB

	// Métricas y estadísticas
	quitLock sync.Mutex      // Protege acceso a quit channel
	quitChan chan chan error // Canal para cerrar de forma segura

	log Logger // Logger para debugging (puede ser nil)
}

// Logger define interfaz de logging (opcional)
type Logger interface {
	Info(msg string, ctx ...interface{})
	Warn(msg string, ctx ...interface{})
	Error(msg string, ctx ...interface{})
	Debug(msg string, ctx ...interface{})
}

// New crea una nueva instancia de base de datos LevelDB
// file: ruta al directorio de la base de datos
// cache: tamaño de cache en MB (0 = default 16MB)
// handles: número de file handles (0 = default 16)
// namespace: prefijo opcional para todas las keys
// readonly: abrir en modo solo lectura
func New(file string, cache int, handles int, namespace string, readonly bool) (*Database, error) {
	return NewCustom(file, namespace, func(options *opt.Options) {
		// Configuración por defecto
		if cache < 16 {
			cache = 16
		}
		if handles < 16 {
			handles = 16
		}

		// Aplicar configuración
		options.OpenFilesCacheCapacity = handles
		options.BlockCacheCapacity = cache / 2 * opt.MiB
		options.WriteBuffer = cache / 4 * opt.MiB // Dos buffers de escritura
		options.Filter = filter.NewBloomFilter(10)

		// Modo read-only
		if readonly {
			options.ReadOnly = true
		}
	})
}

// NewCustom crea una base de datos con opciones personalizadas
func NewCustom(file string, namespace string, customize func(options *opt.Options)) (*Database, error) {
	// Opciones por defecto
	options := &opt.Options{
		OpenFilesCacheCapacity: 16,
		BlockCacheCapacity:     16 * opt.MiB,
		WriteBuffer:            8 * opt.MiB,
		Filter:                 filter.NewBloomFilter(10),
	}

	// Aplicar personalización
	if customize != nil {
		customize(options)
	}

	// Abrir LevelDB
	db, err := leveldb.OpenFile(file, options)
	if _, iscorrupted := err.(*errors.ErrCorrupted); iscorrupted {
		// Intentar recuperar base de datos corrupta
		db, err = leveldb.RecoverFile(file, nil)
	}
	if err != nil {
		return nil, err
	}

	// Crear wrapper
	ldb := &Database{
		fn:       file,
		db:       db,
		quitChan: make(chan chan error),
	}

	return ldb, nil
}

// Close cierra la base de datos
func (db *Database) Close() error {
	db.quitLock.Lock()
	defer db.quitLock.Unlock()

	if db.quitChan != nil {
		// Cerrar canal de métricas si existe
		close(db.quitChan)
		db.quitChan = nil
	}

	if db.db != nil {
		return db.db.Close()
	}

	return nil
}

// Has verifica si una key existe
func (db *Database) Has(key []byte) (bool, error) {
	return db.db.Has(key, nil)
}

// Get obtiene el valor de una key
func (db *Database) Get(key []byte) ([]byte, error) {
	dat, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// Put inserta o actualiza una key
func (db *Database) Put(key []byte, value []byte) error {
	return db.db.Put(key, value, nil)
}

// Delete elimina una key
func (db *Database) Delete(key []byte) error {
	return db.db.Delete(key, nil)
}

// NewBatch crea un nuevo batch
func (db *Database) NewBatch() database.Batch {
	return &batch{
		db: db.db,
		b:  new(leveldb.Batch),
	}
}

// NewBatchWithSize crea un batch con capacidad inicial
func (db *Database) NewBatchWithSize(size int) database.Batch {
	return &batch{
		db: db.db,
		b:  new(leveldb.Batch),
	}
}

// NewIterator crea un iterador
func (db *Database) NewIterator(prefix []byte, start []byte) database.Iterator {
	return db.NewIteratorWithRange(prefix, start, nil)
}

// NewIteratorWithRange crea un iterador con rango
func (db *Database) NewIteratorWithRange(prefix []byte, start []byte, end []byte) database.Iterator {
	// Construir rango
	r := util.BytesPrefix(prefix)
	if start != nil {
		r.Start = append(prefix, start...)
	}
	if end != nil {
		r.Limit = append(prefix, end...)
	}

	return &iter{
		iter: db.db.NewIterator(r, nil),
	}
}

// Stat retorna estadísticas de la base de datos
func (db *Database) Stat(property string) (string, error) {
	return db.db.GetProperty(property)
}

// Compact compacta un rango de keys
func (db *Database) Compact(start []byte, limit []byte) error {
	return db.db.CompactRange(util.Range{Start: start, Limit: limit})
}

// Path retorna la ruta de la base de datos
func (db *Database) Path() string {
	return db.fn
}

// batch implementa database.Batch usando leveldb.Batch
type batch struct {
	db   *leveldb.DB
	b    *leveldb.Batch
	size int
}

func (b *batch) Put(key, value []byte) error {
	b.b.Put(key, value)
	b.size += len(key) + len(value)
	return nil
}

func (b *batch) Delete(key []byte) error {
	b.b.Delete(key)
	b.size += len(key)
	return nil
}

func (b *batch) ValueSize() int {
	return b.size
}

func (b *batch) Write() error {
	return b.db.Write(b.b, nil)
}

func (b *batch) Reset() {
	b.b.Reset()
	b.size = 0
}

func (b *batch) Replay(w database.KeyValueWriter) error {
	return b.b.Replay(&replayer{writer: w})
}

// replayer adapta database.KeyValueWriter para leveldb.Batch.Replay
type replayer struct {
	writer database.KeyValueWriter
	err    error
}

func (r *replayer) Put(key, value []byte) {
	if r.err != nil {
		return
	}
	r.err = r.writer.Put(key, value)
}

func (r *replayer) Delete(key []byte) {
	if r.err != nil {
		return
	}
	r.err = r.writer.Delete(key)
}

// iter adapta leveldb.Iterator a database.Iterator
type iter struct {
	iter iterator.Iterator
}

func (it *iter) Next() bool {
	return it.iter.Next()
}

func (it *iter) Error() error {
	return it.iter.Error()
}

func (it *iter) Key() []byte {
	return it.iter.Key()
}

func (it *iter) Value() []byte {
	return it.iter.Value()
}

func (it *iter) Release() {
	it.iter.Release()
}

// String retorna información de la base de datos
func (db *Database) String() string {
	return fmt.Sprintf("LevelDB: %s", db.fn)
}
