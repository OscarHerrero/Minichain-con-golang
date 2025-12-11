package rlp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"
)

var (
	ErrUnexpectedEnd = errors.New("rlp: unexpected end of input")
	ErrTooLarge      = errors.New("rlp: value too large")
	ErrNonCanonical  = errors.New("rlp: non-canonical encoding")
	ErrListOverflow  = errors.New("rlp: list overflow")
)

// Decode decodifica datos RLP en val
func Decode(data []byte, val interface{}) error {
	// Para datos completos en memoria, usar decodificación simple
	// sin Stream para evitar problemas de EOF
	if val == nil {
		return errors.New("rlp: decode target is nil")
	}

	rval := reflect.ValueOf(val)
	if rval.Kind() != reflect.Ptr || rval.IsNil() {
		return errors.New("rlp: decode target must be a non-nil pointer")
	}

	// Crear un Stream limitado al tamaño del buffer
	s := &Stream{
		r:         bytes.NewReader(data),
		remaining: uint64(len(data)),
		limited:   true,
	}

	return s.Decode(val)
}

// DecodeFrom decodifica desde un Reader
func DecodeFrom(r io.Reader, val interface{}) error {
	s := &Stream{r: r}
	return s.Decode(val)
}

// Stream es un decoder RLP que lee desde un Reader
type Stream struct {
	r        io.Reader
	buf      []byte
	kind     Kind   // Tipo del ítem actual
	size     uint64 // Tamaño del ítem actual
	byteval  byte   // Valor del byte único
	kinderr  error  // Error al leer kind
	stack    []listPos
	limited  bool
	remaining uint64
}

type listPos struct {
	pos uint64
	size uint64
}

// Kind representa el tipo de un valor RLP
type Kind int

const (
	Byte Kind = iota
	String
	List
)

// NewStream crea un nuevo Stream
func NewStream(r io.Reader, inputLimit uint64) *Stream {
	s := &Stream{
		r:   r,
		buf: make([]byte, 9), // Tamaño máximo de header
	}
	if inputLimit != 0 {
		s.limited = true
		s.remaining = inputLimit
	}
	return s
}

// Kind retorna el tipo y tamaño del siguiente valor
func (s *Stream) Kind() (Kind, uint64, error) {
	if s.kinderr != nil {
		return 0, 0, s.kinderr
	}
	if s.kind != 0 {
		return s.kind, s.size, nil
	}

	// Leer primer byte
	b, err := s.readByte()
	if err != nil {
		s.kinderr = err
		return 0, 0, err
	}

	switch {
	case b < 0x80:
		// Byte único
		s.kind = Byte
		s.byteval = b
		s.size = 0

	case b < 0xb8:
		// String corto
		s.kind = String
		s.size = uint64(b - 0x80)

	case b < 0xc0:
		// String largo
		lenLen := int(b - 0xb7)
		if lenLen > 8 {
			return 0, 0, ErrTooLarge
		}
		size, err := s.readUint(lenLen)
		if err != nil {
			return 0, 0, err
		}
		if size < 56 {
			return 0, 0, ErrNonCanonical
		}
		s.kind = String
		s.size = size

	case b < 0xf8:
		// Lista corta
		s.kind = List
		s.size = uint64(b - 0xc0)

	default:
		// Lista larga
		lenLen := int(b - 0xf7)
		if lenLen > 8 {
			return 0, 0, ErrTooLarge
		}
		size, err := s.readUint(lenLen)
		if err != nil {
			return 0, 0, err
		}
		if size < 56 {
			return 0, 0, ErrNonCanonical
		}
		s.kind = List
		s.size = size
	}

	return s.kind, s.size, nil
}

// Decode decodifica el siguiente valor en val
func (s *Stream) Decode(val interface{}) error {
	if val == nil {
		return errors.New("rlp: decode target is nil")
	}

	rval := reflect.ValueOf(val)
	if rval.Kind() != reflect.Ptr || rval.IsNil() {
		return errors.New("rlp: decode target must be a non-nil pointer")
	}

	return s.decode(rval.Elem())
}

// decode decodifica según el tipo de val
func (s *Stream) decode(val reflect.Value) error {
	kind, size, err := s.Kind()
	if err != nil {
		return err
	}

	// Resolver interfaces y pointers
	for val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				val.Set(reflect.New(val.Type().Elem()))
			}
			val = val.Elem()
		} else {
			return errors.New("rlp: cannot decode into interface")
		}
	}

	// Manejar tipos especiales ANTES del switch
	// big.Int debe manejarse antes porque es un struct
	if val.Type() == reflect.TypeOf(big.Int{}) {
		return s.decodeBigInt(val.Addr().Interface().(*big.Int))
	}

	// Decodificar según tipo
	switch val.Kind() {
	case reflect.Bool:
		return s.decodeBool(val)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return s.decodeUint(val)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return s.decodeInt(val)

	case reflect.String:
		return s.decodeString(val)

	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			// []byte
			return s.decodeBytes(val)
		}
		return s.decodeSlice(val)

	case reflect.Array:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			// [N]byte
			return s.decodeByteArray(val)
		}
		return s.decodeArray(val)

	case reflect.Struct:
		return s.decodeStruct(val)

	default:
		return fmt.Errorf("rlp: unsupported type %v (kind=%v, size=%d)", val.Type(), kind, size)
	}
}

func (s *Stream) decodeBool(val reflect.Value) error {
	kind, size, _ := s.Kind()
	if kind == Byte {
		val.SetBool(s.byteval != 0)
		s.kind = 0
		return nil
	}
	if kind == String && size == 0 {
		val.SetBool(false)
		s.kind = 0
		return nil
	}
	if kind == String && size == 1 {
		b, err := s.readByte()
		if err != nil {
			return err
		}
		val.SetBool(b != 0)
		s.kind = 0
		return nil
	}
	return errors.New("rlp: invalid encoding for bool")
}

func (s *Stream) decodeUint(val reflect.Value) error {
	kind, size, _ := s.Kind()
	if kind == Byte {
		val.SetUint(uint64(s.byteval))
		s.kind = 0
		return nil
	}
	if kind != String {
		return errors.New("rlp: expected string for uint")
	}

	if size == 0 {
		val.SetUint(0)
		s.kind = 0
		return nil
	}

	// Leer bytes
	buf := make([]byte, size)
	if err := s.readFull(buf); err != nil {
		return err
	}

	// Verificar no-canonical (sin ceros al inicio excepto si es solo 0x00)
	if len(buf) > 1 && buf[0] == 0 {
		return ErrNonCanonical
	}

	// Convertir a uint64
	if len(buf) > 8 {
		return errors.New("rlp: uint overflow")
	}

	var n uint64
	for _, b := range buf {
		n = n<<8 | uint64(b)
	}

	val.SetUint(n)
	s.kind = 0
	return nil
}

func (s *Stream) decodeInt(val reflect.Value) error {
	var n uint64
	tempVal := reflect.ValueOf(&n).Elem()
	if err := s.decodeUint(tempVal); err != nil {
		return err
	}
	if n > uint64(1<<63-1) {
		return errors.New("rlp: int overflow")
	}
	val.SetInt(int64(n))
	return nil
}

func (s *Stream) decodeString(val reflect.Value) error {
	kind, size, _ := s.Kind()
	if kind == Byte {
		val.SetString(string([]byte{s.byteval}))
		s.kind = 0
		return nil
	}
	if kind != String {
		return errors.New("rlp: expected string")
	}

	buf := make([]byte, size)
	if err := s.readFull(buf); err != nil {
		return err
	}

	val.SetString(string(buf))
	s.kind = 0
	return nil
}

func (s *Stream) decodeBytes(val reflect.Value) error {
	kind, size, _ := s.Kind()
	if kind == Byte {
		val.SetBytes([]byte{s.byteval})
		s.kind = 0
		return nil
	}
	if kind != String {
		return errors.New("rlp: expected string for []byte")
	}

	buf := make([]byte, size)
	if err := s.readFull(buf); err != nil {
		return err
	}

	val.SetBytes(buf)
	s.kind = 0
	return nil
}

func (s *Stream) decodeByteArray(val reflect.Value) error {
	kind, size, _ := s.Kind()
	if kind != String {
		return errors.New("rlp: expected string for byte array")
	}

	if size != uint64(val.Len()) {
		return fmt.Errorf("rlp: array size mismatch: got %d, want %d", size, val.Len())
	}

	buf := make([]byte, size)
	if err := s.readFull(buf); err != nil {
		return err
	}

	reflect.Copy(val, reflect.ValueOf(buf))
	s.kind = 0
	return nil
}

func (s *Stream) decodeSlice(val reflect.Value) error {
	kind, _, _ := s.Kind()
	if kind != List {
		return errors.New("rlp: expected list for slice")
	}

	if err := s.List(); err != nil {
		return err
	}

	// Crear un nuevo slice vacío
	elemType := val.Type().Elem()
	slice := reflect.MakeSlice(val.Type(), 0, 0)

	// Decodificar elementos uno por uno
	for {
		// Crear nuevo elemento
		elem := reflect.New(elemType).Elem()

		// Intentar decodificar
		err := s.decode(elem)
		if err == io.EOF {
			// Fin de la lista
			break
		}
		if err != nil {
			return err
		}

		// Agregar elemento al slice
		slice = reflect.Append(slice, elem)
	}

	// Asignar slice completo al valor
	val.Set(slice)
	return s.ListEnd()
}

func (s *Stream) decodeArray(val reflect.Value) error {
	kind, _, _ := s.Kind()
	if kind != List {
		return errors.New("rlp: expected list for array")
	}

	if err := s.List(); err != nil {
		return err
	}

	for i := 0; i < val.Len(); i++ {
		if err := s.decode(val.Index(i)); err != nil {
			return err
		}
	}

	return s.ListEnd()
}

func (s *Stream) decodeStruct(val reflect.Value) error {
	kind, _, _ := s.Kind()
	if kind != List {
		return errors.New("rlp: expected list for struct")
	}

	if err := s.List(); err != nil {
		return err
	}

	for i := 0; i < val.NumField(); i++ {
		if !val.Type().Field(i).IsExported() {
			continue
		}
		if err := s.decode(val.Field(i)); err != nil {
			return err
		}
	}

	return s.ListEnd()
}

func (s *Stream) decodeBigInt(val *big.Int) error {
	kind, size, _ := s.Kind()
	if kind == Byte {
		val.SetUint64(uint64(s.byteval))
		s.kind = 0
		return nil
	}
	if kind != String {
		return errors.New("rlp: expected string for big.Int")
	}

	if size == 0 {
		val.SetUint64(0)
		s.kind = 0
		return nil
	}

	buf := make([]byte, size)
	if err := s.readFull(buf); err != nil {
		return err
	}

	val.SetBytes(buf)
	s.kind = 0
	return nil
}

// List inicia la decodificación de una lista
func (s *Stream) List() error {
	kind, size, _ := s.Kind()
	if kind != List {
		return errors.New("rlp: expected list")
	}

	s.stack = append(s.stack, listPos{0, size})
	s.kind = 0
	return nil
}

// ListEnd finaliza la decodificación de una lista
func (s *Stream) ListEnd() error {
	if len(s.stack) == 0 {
		return errors.New("rlp: not in list")
	}
	s.stack = s.stack[:len(s.stack)-1]
	return nil
}

// readByte lee un byte
func (s *Stream) readByte() (byte, error) {
	if len(s.buf) > 0 {
		b := s.buf[0]
		s.buf = s.buf[1:]
		return b, nil
	}

	var b [1]byte
	_, err := io.ReadFull(s.r, b[:])
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// readFull lee exactamente len(buf) bytes
func (s *Stream) readFull(buf []byte) error {
	_, err := io.ReadFull(s.r, buf)
	return err
}

// readUint lee un entero de n bytes
func (s *Stream) readUint(n int) (uint64, error) {
	buf := make([]byte, n)
	if err := s.readFull(buf); err != nil {
		return 0, err
	}

	// Verificar no-canonical
	if n > 1 && buf[0] == 0 {
		return 0, ErrNonCanonical
	}

	var val uint64
	for _, b := range buf {
		val = val<<8 | uint64(b)
	}
	return val, nil
}

// Helper para convertir bytes a uint64
func bytesToUint(b []byte) uint64 {
	var result uint64
	for i := 0; i < len(b) && i < 8; i++ {
		result = result<<8 | uint64(b[i])
	}
	return result
}
