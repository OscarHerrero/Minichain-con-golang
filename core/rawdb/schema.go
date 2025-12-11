package rawdb

import (
	"encoding/binary"
)

// Prefijos de keys en LevelDB (como en go-ethereum)
// Basado en go-ethereum/core/rawdb/schema.go

var (
	// headerPrefix + num (uint64 big endian) + hash -> header
	headerPrefix = []byte("h")

	// headerHashPrefix + num (uint64 big endian) -> hash
	headerHashPrefix = []byte("H")

	// headerNumberPrefix + hash -> num (uint64 big endian)
	headerNumberPrefix = []byte("l") // 'l' = lookup

	// bodyPrefix + num (uint64 big endian) + hash -> block body
	bodyPrefix = []byte("b")

	// txLookupPrefix + hash -> transaction/receipt lookup metadata
	txLookupPrefix = []byte("t")

	// Metadata keys
	headHeaderKey = []byte("LastHeader")
	headBlockKey  = []byte("LastBlock")
)

// encodeBlockNumber codifica un número de bloque en 8 bytes big endian
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

// headerKey = headerPrefix + num (uint64 big endian) + hash
func headerKey(number uint64, hash []byte) []byte {
	return append(append(headerPrefix, encodeBlockNumber(number)...), hash...)
}

// headerHashKey = headerHashPrefix + num (uint64 big endian)
func headerHashKey(number uint64) []byte {
	return append(headerHashPrefix, encodeBlockNumber(number)...)
}

// headerNumberKey = headerNumberPrefix + hash
func headerNumberKey(hash []byte) []byte {
	return append(headerNumberPrefix, hash...)
}

// bodyKey = bodyPrefix + num (uint64 big endian) + hash
func bodyKey(number uint64, hash []byte) []byte {
	return append(append(bodyPrefix, encodeBlockNumber(number)...), hash...)
}

// txLookupKey = txLookupPrefix + hash
func txLookupKey(hash []byte) []byte {
	return append(txLookupPrefix, hash...)
}
