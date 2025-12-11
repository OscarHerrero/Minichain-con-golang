package trie

// Hex/Compact encoding para keys en Merkle Patricia Trie
// Basado en go-ethereum/trie/encoding.go
//
// Ethereum usa dos encodings para keys:
// 1. HEX: cada byte se expande a 2 nibbles (0-15) + terminator opcional
// 2. COMPACT: comprime nibbles en bytes

// keybytesToHex convierte bytes de key a hex encoding
// Cada byte se expande a 2 nibbles (4 bits cada uno)
// Agrega un terminador 16 al final para indicar "leaf"
func keybytesToHex(str []byte) []byte {
	l := len(str)*2 + 1
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16     // Nibble alto
		nibbles[i*2+1] = b % 16   // Nibble bajo
	}
	nibbles[l-1] = 16 // Terminator para indicar leaf
	return nibbles
}

// hexToKeybytes convierte hex encoding a bytes normales
// Remueve el terminator si existe
func hexToKeybytes(hex []byte) []byte {
	if hasTerm(hex) {
		hex = hex[:len(hex)-1]
	}
	if len(hex)&1 != 0 {
		panic("hex string has odd length")
	}
	key := make([]byte, len(hex)/2)
	for i := 0; i < len(key); i++ {
		key[i] = hex[i*2]<<4 | hex[i*2+1]
	}
	return key
}

// hasTerm verifica si el hex tiene terminator (16 al final)
func hasTerm(s []byte) bool {
	return len(s) > 0 && s[len(s)-1] == 16
}

// Compact Encoding
// ================
// Compact encoding reduce el espacio usado por nibbles
//
// Especificación:
// - Si la longitud es par (extension node):
//   - Agrega prefix byte con flag 0 en primer nibble
//   - Ejemplo: [0, 1, 2, 3] → [0x00, 0x01, 0x23]
//
// - Si la longitud es impar (extension node):
//   - Agrega prefix byte con flag 1 + primer nibble
//   - Ejemplo: [1, 2, 3] → [0x11, 0x23]
//
// - Si tiene terminator (leaf node):
//   - Agrega flag 2 (par) o flag 3 (impar)
//   - Ejemplo: [0, 1, 2, 3, 16] → [0x20, 0x01, 0x23]
//   - Ejemplo: [1, 2, 3, 16] → [0x31, 0x23]

// compactEncode convierte nibbles hex a compact encoding
func compactEncode(hex []byte) []byte {
	terminator := byte(0)
	if hasTerm(hex) {
		terminator = 1
		hex = hex[:len(hex)-1]
	}

	buf := make([]byte, len(hex)/2+1)
	buf[0] = terminator << 5 // Flag en los primeros 3 bits

	if len(hex)&1 == 1 {
		// Longitud impar: poner flag odd + primer nibble
		buf[0] |= 1 << 4 // Set odd flag
		buf[0] |= hex[0] // Primer nibble
		hex = hex[1:]
	}

	// Empaquetar nibbles restantes
	for i := 0; i < len(hex); i += 2 {
		buf[i/2+1] = hex[i]<<4 | hex[i+1]
	}

	return buf
}

// compactDecode decodifica compact encoding a nibbles hex
func compactDecode(compact []byte) []byte {
	if len(compact) == 0 {
		return compact
	}

	base := keybytesToHex(compact)
	// Eliminar los primeros dos nibbles si es par
	// o el primer nibble si es impar
	if base[0] < 2 {
		base = base[2:]
	} else {
		base = base[1:]
	}

	// Agregar terminator si tenía flag de leaf
	if compact[0]>>5 == 1 {
		base = append(base, 16)
	}

	return base
}

// prefixLen retorna la longitud del prefijo común entre dos slices
func prefixLen(a, b []byte) int {
	var i, length = 0, len(a)
	if len(b) < length {
		length = len(b)
	}
	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}

// commonPrefix retorna el prefijo común entre a y b
func commonPrefix(a, b []byte) []byte {
	length := prefixLen(a, b)
	return a[:length]
}
