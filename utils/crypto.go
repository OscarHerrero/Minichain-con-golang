package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// CalculateHash calcula el hash SHA-256 de un string
// Lo usaremos para crear el "fingerprint" único de cada bloque
func CalculateHash(data string) string {
	// sha256.Sum256 toma los bytes y devuelve un array de 32 bytes
	hash := sha256.Sum256([]byte(data))
	
	// Convertimos los bytes a hexadecimal para que sea legible
	// Ej: [255, 32, 18...] → "ff2012..."
	return hex.EncodeToString(hash[:])
}

// CalculateHashBytes es lo mismo pero acepta bytes directamente
func CalculateHashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// MeetsTarget verifica si un hash cumple con la dificultad del minado
// difficulty = cantidad de ceros al inicio que debe tener el hash
// Ej: difficulty=3 → hash debe empezar con "000..."
func MeetsTarget(hash string, difficulty int) bool {
	// Creamos un string con N ceros
	target := ""
	for i := 0; i < difficulty; i++ {
		target += "0"
	}
	
	// Verificamos si el hash empieza con esos ceros
	// Ej: "000a4f2..." cumple con difficulty=3
	return len(hash) >= difficulty && hash[:difficulty] == target
}

// PrintHash imprime un hash de forma bonita
func PrintHash(label string, hash string) {
	fmt.Printf("%s: %s\n", label, hash)
}