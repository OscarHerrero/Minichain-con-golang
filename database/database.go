package database

import "io"

// Database es la interfaz base para almacenamiento clave-valor
// Basada en ethdb.Database de go-ethereum
type Database interface {
	// Reader interface
	KeyValueReader
	// Writer interface
	KeyValueWriter
	// Batcher interface
	Batcher
	// Iteratee interface
	Iteratee
	// Stater interface
	Stater
	// Compacter interface
	Compacter
	// Closer
	io.Closer
}

// KeyValueReader define operaciones de lectura
type KeyValueReader interface {
	// Has verifica si una key existe en la base de datos
	Has(key []byte) (bool, error)

	// Get obtiene el valor de una key
	// Retorna error si la key no existe
	Get(key []byte) ([]byte, error)
}

// KeyValueWriter define operaciones de escritura
type KeyValueWriter interface {
	// Put inserta o actualiza una key con un valor
	Put(key []byte, value []byte) error

	// Delete elimina una key de la base de datos
	Delete(key []byte) error
}

// Batcher permite agrupar múltiples operaciones
type Batcher interface {
	// NewBatch crea un batch para escrituras atómicas
	NewBatch() Batch

	// NewBatchWithSize crea un batch con capacidad pre-asignada
	NewBatchWithSize(size int) Batch
}

// Batch representa un conjunto de operaciones que se ejecutan atómicamente
type Batch interface {
	KeyValueWriter

	// ValueSize retorna el tamaño total de datos en el batch
	ValueSize() int

	// Write ejecuta todas las operaciones del batch
	Write() error

	// Reset limpia el batch para reutilizarlo
	Reset()

	// Replay ejecuta las operaciones del batch en otro KeyValueWriter
	Replay(w KeyValueWriter) error
}

// Iteratee permite iterar sobre rangos de keys
type Iteratee interface {
	// NewIterator crea un iterador sobre un rango de keys
	// prefix: prefijo para filtrar keys (puede ser nil para todas)
	// start: key inicial del rango (puede ser nil para el inicio)
	NewIterator(prefix []byte, start []byte) Iterator
}

// Iterator permite recorrer pares clave-valor
type Iterator interface {
	// Next mueve el iterador al siguiente elemento
	// Retorna false cuando no hay más elementos
	Next() bool

	// Error retorna cualquier error de iteración
	Error() error

	// Key retorna la key actual
	Key() []byte

	// Value retorna el valor actual
	Value() []byte

	// Release libera recursos del iterador
	Release()
}

// Stater proporciona estadísticas de la base de datos
type Stater interface {
	// Stat retorna una propiedad específica de la base de datos
	Stat(property string) (string, error)

	// Compact compacta la base de datos en un rango de keys
	// Compact(start []byte, limit []byte) error
}

// Compacter permite compactar rangos de la base de datos
type Compacter interface {
	// Compact compacta la base de datos en el rango [start, limit)
	Compact(start []byte, limit []byte) error
}

// AncientReader define lectura de datos "antiguos" (freezer)
// Para Minichain no lo usaremos inicialmente, pero lo dejamos para compatibilidad
type AncientReader interface {
	// HasAncient verifica si un elemento antiguo existe
	HasAncient(kind string, number uint64) (bool, error)

	// Ancient retorna un elemento antiguo
	Ancient(kind string, number uint64) ([]byte, error)

	// AncientRange retorna múltiples elementos antiguos
	AncientRange(kind string, start, count, maxBytes uint64) ([][]byte, error)

	// Ancients retorna el número total de elementos antiguos
	Ancients() (uint64, error)

	// AncientSize retorna el tamaño de datos antiguos
	AncientSize(kind string) (uint64, error)
}

// AncientWriter define escritura de datos antiguos
type AncientWriter interface {
	// ModifyAncients ejecuta una función de escritura en el freezer
	ModifyAncients(func(AncientWriteOp) error) (int64, error)

	// TruncateHead descarta datos antiguos más recientes que el número dado
	TruncateHead(n uint64) error

	// TruncateTail descarta datos antiguos más viejos que el número dado
	TruncateTail(n uint64) error

	// Sync asegura que todos los datos estén escritos a disco
	Sync() error
}

// AncientWriteOp representa una operación de escritura en el freezer
type AncientWriteOp interface {
	// Append añade un elemento al freezer
	Append(kind string, number uint64, item interface{}) error

	// AppendRaw añade datos crudos al freezer
	AppendRaw(kind string, number uint64, item []byte) error
}

// AncientStore combina lectura y escritura de datos antiguos
type AncientStore interface {
	AncientReader
	AncientWriter
	io.Closer
}
