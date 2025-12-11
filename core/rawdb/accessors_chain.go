package rawdb

import (
	"encoding/binary"
	"fmt"
	"minichain/database"
	"minichain/rlp"
)

// BlockHeader representa el header de un bloque
// Lo definimos aquí para evitar dependencias circulares
// En la integración final usaremos el Block de blockchain/
type BlockHeader struct {
	ParentHash  []byte
	Number      uint64
	StateRoot   []byte
	TxRoot      []byte
	ReceiptRoot []byte
	Timestamp   int64
	Difficulty  int
	Nonce       int
	Hash        []byte
}

// BlockBody representa el body de un bloque (transacciones)
type BlockBody struct {
	Transactions [][]byte // Transacciones RLP encoded
}

// ReadCanonicalHash obtiene el hash canónico de un número de bloque
func ReadCanonicalHash(db database.KeyValueReader, number uint64) ([]byte, error) {
	data, err := db.Get(headerHashKey(number))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// WriteCanonicalHash escribe el hash canónico de un número de bloque
func WriteCanonicalHash(db database.KeyValueWriter, hash []byte, number uint64) error {
	return db.Put(headerHashKey(number), hash)
}

// ReadHeaderNumber obtiene el número de bloque de un hash
func ReadHeaderNumber(db database.KeyValueReader, hash []byte) (uint64, error) {
	data, err := db.Get(headerNumberKey(hash))
	if err != nil {
		return 0, err
	}
	if len(data) != 8 {
		return 0, fmt.Errorf("invalid header number data")
	}
	return binary.BigEndian.Uint64(data), nil
}

// WriteHeaderNumber escribe el número de bloque de un hash
func WriteHeaderNumber(db database.KeyValueWriter, hash []byte, number uint64) error {
	return db.Put(headerNumberKey(hash), encodeBlockNumber(number))
}

// ReadHeader lee un header de bloque
func ReadHeader(db database.KeyValueReader, hash []byte, number uint64) (*BlockHeader, error) {
	data, err := db.Get(headerKey(number, hash))
	if err != nil {
		return nil, err
	}

	header := new(BlockHeader)
	if err := rlp.Decode(data, header); err != nil {
		return nil, err
	}

	return header, nil
}

// WriteHeader escribe un header de bloque
func WriteHeader(db database.KeyValueWriter, header *BlockHeader) error {
	data, err := rlp.Encode(header)
	if err != nil {
		return err
	}

	// Escribir header
	if err := db.Put(headerKey(header.Number, header.Hash), data); err != nil {
		return err
	}

	// Escribir número de bloque lookup
	if err := WriteHeaderNumber(db, header.Hash, header.Number); err != nil {
		return err
	}

	return nil
}

// ReadBody lee el body de un bloque
func ReadBody(db database.KeyValueReader, hash []byte, number uint64) (*BlockBody, error) {
	data, err := db.Get(bodyKey(number, hash))
	if err != nil {
		return nil, err
	}

	body := new(BlockBody)
	if err := rlp.Decode(data, body); err != nil {
		return nil, err
	}

	return body, nil
}

// WriteBody escribe el body de un bloque
func WriteBody(db database.KeyValueWriter, hash []byte, number uint64, body *BlockBody) error {
	data, err := rlp.Encode(body)
	if err != nil {
		return err
	}

	return db.Put(bodyKey(number, hash), data)
}

// ReadBlock lee un bloque completo (header + body)
func ReadBlock(db database.KeyValueReader, hash []byte, number uint64) (*BlockHeader, *BlockBody, error) {
	header, err := ReadHeader(db, hash, number)
	if err != nil {
		return nil, nil, err
	}

	body, err := ReadBody(db, hash, number)
	if err != nil {
		return nil, nil, err
	}

	return header, body, nil
}

// WriteBlock escribe un bloque completo
func WriteBlock(db database.KeyValueWriter, header *BlockHeader, body *BlockBody) error {
	// Escribir header
	if err := WriteHeader(db, header); err != nil {
		return err
	}

	// Escribir body
	if err := WriteBody(db, header.Hash, header.Number, body); err != nil {
		return err
	}

	return nil
}

// ReadHeadHeaderHash obtiene el hash del header más reciente
func ReadHeadHeaderHash(db database.KeyValueReader) ([]byte, error) {
	return db.Get(headHeaderKey)
}

// WriteHeadHeaderHash escribe el hash del header más reciente
func WriteHeadHeaderHash(db database.KeyValueWriter, hash []byte) error {
	return db.Put(headHeaderKey, hash)
}

// ReadHeadBlockHash obtiene el hash del bloque más reciente
func ReadHeadBlockHash(db database.KeyValueReader) ([]byte, error) {
	return db.Get(headBlockKey)
}

// WriteHeadBlockHash escribe el hash del bloque más reciente
func WriteHeadBlockHash(db database.KeyValueWriter, hash []byte) error {
	return db.Put(headBlockKey, hash)
}

// DeleteHeader elimina un header de bloque
func DeleteHeader(db database.KeyValueWriter, hash []byte, number uint64) error {
	if err := db.Delete(headerKey(number, hash)); err != nil {
		return err
	}
	return db.Delete(headerNumberKey(hash))
}

// DeleteBody elimina el body de un bloque
func DeleteBody(db database.KeyValueWriter, hash []byte, number uint64) error {
	return db.Delete(bodyKey(number, hash))
}

// DeleteBlock elimina un bloque completo
func DeleteBlock(db database.KeyValueWriter, hash []byte, number uint64) error {
	if err := DeleteHeader(db, hash, number); err != nil {
		return err
	}
	return DeleteBody(db, hash, number)
}

// DeleteCanonicalHash elimina el hash canónico de un número
func DeleteCanonicalHash(db database.KeyValueWriter, number uint64) error {
	return db.Delete(headerHashKey(number))
}
