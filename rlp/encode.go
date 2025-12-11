package rlp

import (
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"reflect"
)

// RLP (Recursive Length Prefix) encoding
// Especificación: https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp/
//
// Reglas de encoding:
// 1. String de 0-55 bytes: [0x80 + len, ...data...]
// 2. String de 56+ bytes: [0xb7 + len(len), ...len..., ...data...]
// 3. Lista de 0-55 bytes total: [0xc0 + len, ...items...]
// 4. Lista de 56+ bytes total: [0xf7 + len(len), ...len..., ...items...]
// 5. Byte único [0x00, 0x7f]: se representa como sí mismo

const (
	// Constantes RLP
	stringShort = 0x80 // [0x80, 0xb7] - string de 0-55 bytes
	stringLong  = 0xb7 // [0xb8, 0xbf] - string de 56+ bytes
	listShort   = 0xc0 // [0xc0, 0xf7] - lista de 0-55 bytes
	listLong    = 0xf7 // [0xf8, 0xff] - lista de 56+ bytes
)

var (
	ErrNegativeBigInt = errors.New("rlp: cannot encode negative *big.Int")
	ErrNilValue       = errors.New("rlp: cannot encode nil value")
)

// Encode codifica un valor a RLP
func Encode(val interface{}) ([]byte, error) {
	w := &encBuffer{}
	if err := encode(w, reflect.ValueOf(val)); err != nil {
		return nil, err
	}
	return w.toBytes(), nil
}

// EncodeToWriter codifica un valor a un Writer
func EncodeToWriter(w io.Writer, val interface{}) error {
	buf := &encBuffer{}
	if err := encode(buf, reflect.ValueOf(val)); err != nil {
		return err
	}
	_, err := w.Write(buf.toBytes())
	return err
}

// encBuffer es un buffer para construir output RLP
type encBuffer struct {
	str []byte   // Datos codificados
	lh  lhStack  // Stack de list headers
}

func (w *encBuffer) toBytes() []byte {
	return w.str
}

func (w *encBuffer) Write(b []byte) (int, error) {
	w.str = append(w.str, b...)
	return len(b), nil
}

func (w *encBuffer) WriteByte(b byte) error {
	w.str = append(w.str, b)
	return nil
}

// lhStack maneja el stack de list headers
type lhStack struct {
	stack []listHead
}

type listHead struct {
	offset int // Posición en el buffer
	size   int // Tamaño del contenido
}

func (s *lhStack) push(offset, size int) {
	s.stack = append(s.stack, listHead{offset, size})
}

func (s *lhStack) pop() (int, int) {
	if len(s.stack) == 0 {
		return 0, 0
	}
	last := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return last.offset, last.size
}

// encode es el codificador principal
func encode(w *encBuffer, val reflect.Value) error {
	// Manejar nil
	if !val.IsValid() {
		w.str = append(w.str, 0x80) // String vacío
		return nil
	}

	// Resolver interfaces y pointers
	for val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		if val.IsNil() {
			w.str = append(w.str, 0x80)
			return nil
		}
		val = val.Elem()
	}

	// Codificar según tipo
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			w.str = append(w.str, 0x01)
		} else {
			w.str = append(w.str, 0x80) // false = string vacío
		}
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return encodeUint(w, val.Uint())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int() < 0 {
			return errors.New("rlp: cannot encode negative integer")
		}
		return encodeUint(w, uint64(val.Int()))

	case reflect.String:
		return encodeString(w, []byte(val.String()))

	case reflect.Slice, reflect.Array:
		return encodeList(w, val)

	case reflect.Struct:
		return encodeStruct(w, val)

	default:
		// Tipos especiales
		if val.Type() == reflect.TypeOf(big.Int{}) {
			return encodeBigInt(w, val.Addr().Interface().(*big.Int))
		}
		return errors.New("rlp: unsupported type " + val.Type().String())
	}
}

// encodeUint codifica un unsigned integer
func encodeUint(w *encBuffer, i uint64) error {
	if i == 0 {
		w.str = append(w.str, 0x80) // String vacío
		return nil
	}
	if i < 0x80 {
		w.str = append(w.str, byte(i)) // Byte único
		return nil
	}

	// Convertir a bytes (big-endian)
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], i)

	// Quitar ceros al inicio
	start := 0
	for start < 8 && b[start] == 0 {
		start++
	}

	return encodeString(w, b[start:])
}

// encodeBigInt codifica un *big.Int
func encodeBigInt(w *encBuffer, i *big.Int) error {
	if i == nil {
		return ErrNilValue
	}
	if i.Sign() < 0 {
		return ErrNegativeBigInt
	}
	if i.Sign() == 0 {
		w.str = append(w.str, 0x80) // String vacío
		return nil
	}

	// Convertir a bytes
	b := i.Bytes()
	return encodeString(w, b)
}

// encodeString codifica un byte slice (string)
func encodeString(w *encBuffer, b []byte) error {
	if len(b) == 1 && b[0] < 0x80 {
		// Byte único menor a 0x80
		w.str = append(w.str, b[0])
		return nil
	}

	if len(b) < 56 {
		// String corto: [0x80 + len] + data
		w.str = append(w.str, byte(0x80+len(b)))
		w.str = append(w.str, b...)
	} else {
		// String largo: [0xb7 + len(len)] + len + data
		lenLen := putIntLen(len(b))
		w.str = append(w.str, byte(0xb7+lenLen))
		w.str = append(w.str, intToBytes(len(b), lenLen)...)
		w.str = append(w.str, b...)
	}
	return nil
}

// encodeList codifica un slice o array
func encodeList(w *encBuffer, val reflect.Value) error {
	// Para []byte, tratar como string
	if val.Type().Elem().Kind() == reflect.Uint8 {
		b := val.Bytes()
		return encodeString(w, b)
	}

	// Guardar posición inicial
	startPos := len(w.str)

	// Reservar espacio para header (lo escribiremos después)
	w.str = append(w.str, 0, 0, 0, 0, 0, 0, 0, 0, 0) // Max 9 bytes de header

	// Codificar elementos
	contentStart := len(w.str)
	for i := 0; i < val.Len(); i++ {
		if err := encode(w, val.Index(i)); err != nil {
			return err
		}
	}

	// Calcular tamaño del contenido
	contentSize := len(w.str) - contentStart

	// Escribir header correcto
	if contentSize < 56 {
		// Lista corta: [0xc0 + len] + content
		w.str[startPos] = byte(0xc0 + contentSize)
		// Mover contenido 1 byte atrás (solo necesitamos 1 byte de header)
		copy(w.str[startPos+1:], w.str[contentStart:])
		w.str = w.str[:startPos+1+contentSize]
	} else {
		// Lista larga: [0xf7 + len(len)] + len + content
		lenLen := putIntLen(contentSize)
		w.str[startPos] = byte(0xf7 + lenLen)
		copy(w.str[startPos+1:], intToBytes(contentSize, lenLen))
		// Mover contenido
		headerSize := 1 + lenLen
		copy(w.str[startPos+headerSize:], w.str[contentStart:])
		w.str = w.str[:startPos+headerSize+contentSize]
	}

	return nil
}

// encodeStruct codifica una struct como lista
func encodeStruct(w *encBuffer, val reflect.Value) error {
	// Guardar posición inicial
	startPos := len(w.str)
	w.str = append(w.str, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	// Codificar campos
	contentStart := len(w.str)
	for i := 0; i < val.NumField(); i++ {
		// Ignorar campos no exportados
		if !val.Type().Field(i).IsExported() {
			continue
		}
		if err := encode(w, val.Field(i)); err != nil {
			return err
		}
	}

	// Calcular tamaño y escribir header
	contentSize := len(w.str) - contentStart
	if contentSize < 56 {
		w.str[startPos] = byte(0xc0 + contentSize)
		copy(w.str[startPos+1:], w.str[contentStart:])
		w.str = w.str[:startPos+1+contentSize]
	} else {
		lenLen := putIntLen(contentSize)
		w.str[startPos] = byte(0xf7 + lenLen)
		copy(w.str[startPos+1:], intToBytes(contentSize, lenLen))
		headerSize := 1 + lenLen
		copy(w.str[startPos+headerSize:], w.str[contentStart:])
		w.str = w.str[:startPos+headerSize+contentSize]
	}

	return nil
}

// putIntLen retorna cuántos bytes se necesitan para representar n
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

// intToBytes convierte un int a bytes big-endian
func intToBytes(n int, bytes int) []byte {
	b := make([]byte, bytes)
	for i := bytes - 1; i >= 0; i-- {
		b[i] = byte(n)
		n >>= 8
	}
	return b
}
