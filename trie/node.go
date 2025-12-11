package trie

import (
	"fmt"
	"io"
	"minichain/rlp"
)

// EncWriter es una interfaz simplificada para encoding RLP
// Para compatibilidad con nuestro paquete rlp
type EncWriter interface {
	EncodeString([]byte) error
	EncodeList(func() error) error
}

// node es la interfaz que implementan todos los tipos de nodos del trie
// Basado en go-ethereum/trie/node.go
type node interface {
	// cache devuelve el hash cacheado del nodo (nil si no está cacheado)
	cache() (hashNode, bool)

	// encode escribe el nodo en RLP a un writer
	encode(w EncWriter) error
}

// Los 4 tipos de nodos en Ethereum Merkle Patricia Trie:
// 1. fullNode (Branch Node): tiene 16 hijos + 1 valor opcional
// 2. shortNode: Extension o Leaf node
// 3. hashNode: referencia a otro nodo por su hash
// 4. valueNode: valor final (datos del leaf)

// fullNode representa un branch node con 16 hijos
// Cada hijo corresponde a un nibble hex (0-15)
type fullNode struct {
	Children [17]node // 16 hijos + 1 valor en posición 16
	flags    nodeFlag
}

// shortNode representa un extension o leaf node
// Key: nibbles compartidos o restantes
// Val: hashNode (extension) o valueNode (leaf)
type shortNode struct {
	Key   []byte // Nibbles en hex/compact encoding
	Val   node
	flags nodeFlag
}

// hashNode es una referencia a otro nodo por su hash (32 bytes)
// Permite cargar nodos bajo demanda desde la base de datos
type hashNode []byte

// valueNode es el valor final almacenado en un leaf
// Típicamente datos RLP encoded
type valueNode []byte

// nodeFlag mantiene información de caché del nodo
type nodeFlag struct {
	hash  hashNode // Hash cacheado del nodo
	dirty bool     // Si el nodo fue modificado desde la última serialización
}

// Implementación de interfaz node para fullNode
func (n *fullNode) cache() (hashNode, bool) {
	return n.flags.hash, n.flags.dirty
}

func (n *fullNode) encode(w EncWriter) error {
	// Un fullNode se codifica como una lista de 17 elementos
	return w.EncodeList(func() error {
		// Codificar los 16 hijos
		for i := 0; i < 16; i++ {
			if n.Children[i] != nil {
				if err := n.Children[i].encode(w); err != nil {
					return err
				}
			} else {
				// Hijo vacío = string vacío en RLP
				if err := w.EncodeString(nil); err != nil {
					return err
				}
			}
		}
		// Codificar el valor en posición 16
		if n.Children[16] != nil {
			if err := n.Children[16].encode(w); err != nil {
				return err
			}
		} else {
			if err := w.EncodeString(nil); err != nil {
				return err
			}
		}
		return nil
	})
}

// Implementación de interfaz node para shortNode
func (n *shortNode) cache() (hashNode, bool) {
	return n.flags.hash, n.flags.dirty
}

func (n *shortNode) encode(w EncWriter) error {
	// Un shortNode se codifica como [key, value]
	return w.EncodeList(func() error {
		// Codificar key en compact encoding
		key := compactEncode(n.Key)
		if err := w.EncodeString(key); err != nil {
			return err
		}
		// Codificar value
		return n.Val.encode(w)
	})
}

// Implementación de interfaz node para hashNode
func (n hashNode) cache() (hashNode, bool) {
	return n, true
}

func (n hashNode) encode(w EncWriter) error {
	// Un hashNode se codifica como sus bytes directamente
	return w.EncodeString([]byte(n))
}

// Implementación de interfaz node para valueNode
func (n valueNode) cache() (hashNode, bool) {
	return nil, true
}

func (n valueNode) encode(w EncWriter) error {
	// Un valueNode se codifica como sus bytes directamente
	return w.EncodeString([]byte(n))
}

// encBuffer implementa EncWriter usando nuestro RLP
type encBuffer struct {
	buf []byte
}

func (w *encBuffer) EncodeString(b []byte) error {
	encoded, err := rlp.Encode(b)
	if err != nil {
		return err
	}
	w.buf = append(w.buf, encoded...)
	return nil
}

func (w *encBuffer) EncodeList(f func() error) error {
	// Guardar posición inicial
	start := len(w.buf)

	// Reservar espacio para header
	w.buf = append(w.buf, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	// Ejecutar función que codifica elementos
	contentStart := len(w.buf)
	if err := f(); err != nil {
		return err
	}

	// Calcular tamaño del contenido
	contentSize := len(w.buf) - contentStart

	// Escribir header correcto
	if contentSize < 56 {
		// Lista corta
		w.buf[start] = byte(0xc0 + contentSize)
		copy(w.buf[start+1:], w.buf[contentStart:])
		w.buf = w.buf[:start+1+contentSize]
	} else {
		// Lista larga
		lenLen := putIntLen(contentSize)
		w.buf[start] = byte(0xf7 + lenLen)
		copy(w.buf[start+1:], intToBytes(contentSize, lenLen))
		headerSize := 1 + lenLen
		copy(w.buf[start+headerSize:], w.buf[contentStart:])
		w.buf = w.buf[:start+headerSize+contentSize]
	}

	return nil
}

func putIntLen(n int) int {
	if n < 256 {
		return 1
	}
	if n < 65536 {
		return 2
	}
	if n < 16777216 {
		return 3
	}
	return 4
}

func intToBytes(n int, bytes int) []byte {
	b := make([]byte, bytes)
	for i := bytes - 1; i >= 0; i-- {
		b[i] = byte(n)
		n >>= 8
	}
	return b
}

// mustDecodeNode decodifica un nodo desde bytes RLP
func mustDecodeNode(hash, buf []byte) node {
	n, err := decodeNode(hash, buf)
	if err != nil {
		panic(fmt.Sprintf("node decode error: %v", err))
	}
	return n
}

// decodeNode decodifica un nodo desde bytes RLP
func decodeNode(hash, buf []byte) (node, error) {
	if len(buf) == 0 {
		return nil, io.ErrUnexpectedEOF
	}

	// Primer byte determina el tipo
	elems, _, err := splitList(buf)
	if err != nil {
		// No es una lista, es un valor directo
		return decodeShort(hash, buf)
	}

	// Contar elementos
	count := 0
	for {
		_, rest, err := splitString(elems)
		if err != nil {
			break
		}
		count++
		elems = rest
	}

	switch count {
	case 2:
		// shortNode (extension o leaf)
		return decodeShort(hash, buf)
	case 17:
		// fullNode (branch)
		return decodeFull(hash, buf)
	default:
		return nil, fmt.Errorf("invalid number of list elements: %d", count)
	}
}

func decodeShort(hash, buf []byte) (node, error) {
	elems, _, err := splitList(buf)
	if err != nil {
		return nil, fmt.Errorf("not a list: %w", err)
	}

	// Primer elemento: key
	keyBytes, rest, err := splitString(elems)
	if err != nil {
		return nil, err
	}

	// Decodificar key de compact encoding
	key := compactDecode(keyBytes)

	// Segundo elemento: value
	valBytes, _, err := splitString(rest)
	if err != nil {
		return nil, err
	}

	// Si value es un hash, crear hashNode
	// Si no, es un valueNode
	var val node
	if len(valBytes) == 32 {
		val = hashNode(valBytes)
	} else {
		val = valueNode(valBytes)
	}

	return &shortNode{Key: key, Val: val}, nil
}

func decodeFull(hash, buf []byte) (node, error) {
	n := &fullNode{}
	elems, _, err := splitList(buf)
	if err != nil {
		return nil, err
	}

	// Decodificar 17 elementos
	for i := 0; i < 17; i++ {
		childBytes, rest, err := splitString(elems)
		if err != nil {
			return nil, err
		}

		if len(childBytes) > 0 {
			if len(childBytes) == 32 {
				n.Children[i] = hashNode(childBytes)
			} else {
				n.Children[i] = valueNode(childBytes)
			}
		}

		elems = rest
	}

	return n, nil
}

// splitList divide un buffer RLP en su contenido de lista y resto
func splitList(buf []byte) (content, rest []byte, err error) {
	if len(buf) == 0 {
		return nil, nil, io.ErrUnexpectedEOF
	}

	b := buf[0]
	if b < 0xc0 {
		return nil, nil, fmt.Errorf("not a list")
	}

	if b < 0xf8 {
		// Lista corta
		size := int(b - 0xc0)
		if len(buf) < 1+size {
			return nil, nil, io.ErrUnexpectedEOF
		}
		return buf[1 : 1+size], buf[1+size:], nil
	}

	// Lista larga
	lenLen := int(b - 0xf7)
	if len(buf) < 1+lenLen {
		return nil, nil, io.ErrUnexpectedEOF
	}

	size := 0
	for i := 0; i < lenLen; i++ {
		size = size<<8 | int(buf[1+i])
	}

	start := 1 + lenLen
	if len(buf) < start+size {
		return nil, nil, io.ErrUnexpectedEOF
	}

	return buf[start : start+size], buf[start+size:], nil
}

// splitString divide un buffer RLP en su string y resto
func splitString(buf []byte) (content, rest []byte, err error) {
	if len(buf) == 0 {
		return nil, nil, io.ErrUnexpectedEOF
	}

	b := buf[0]

	if b < 0x80 {
		// Byte único
		return buf[:1], buf[1:], nil
	}

	if b < 0xb8 {
		// String corto
		size := int(b - 0x80)
		if len(buf) < 1+size {
			return nil, nil, io.ErrUnexpectedEOF
		}
		return buf[1 : 1+size], buf[1+size:], nil
	}

	if b < 0xc0 {
		// String largo
		lenLen := int(b - 0xb7)
		if len(buf) < 1+lenLen {
			return nil, nil, io.ErrUnexpectedEOF
		}

		size := 0
		for i := 0; i < lenLen; i++ {
			size = size<<8 | int(buf[1+i])
		}

		start := 1 + lenLen
		if len(buf) < start+size {
			return nil, nil, io.ErrUnexpectedEOF
		}

		return buf[start : start+size], buf[start+size:], nil
	}

	return nil, nil, fmt.Errorf("not a string")
}
