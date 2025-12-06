package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// KeyPair representa un par de claves pública/privada
type KeyPair struct {
	PrivateKey *ecdsa.PrivateKey // Clave privada (NUNCA compartir)
	PublicKey  *ecdsa.PublicKey  // Clave pública (tu "dirección")
}

// GenerateKeyPair genera un nuevo par de claves usando curva elíptica
// Usa el mismo algoritmo que Bitcoin (secp256k1 simulado con P256)
func GenerateKeyPair() (*KeyPair, error) {
	// Generar clave privada usando curva elíptica P256
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("error generando clave privada: %v", err)
	}

	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// GetAddress convierte la clave pública en una dirección legible
// Similar a cómo Bitcoin/Ethereum generan direcciones desde la clave pública
func (kp *KeyPair) GetAddress() string {
	// Concatenar las coordenadas X e Y de la clave pública
	pubKeyBytes := append(kp.PublicKey.X.Bytes(), kp.PublicKey.Y.Bytes()...)

	// Hash SHA-256 de la clave pública
	hash := sha256.Sum256(pubKeyBytes)

	// Convertir a hexadecimal y tomar los primeros 40 caracteres
	// (Ethereum usa 40 caracteres, Bitcoin usa formato diferente)
	address := hex.EncodeToString(hash[:])[:40]

	return address
}

// SignData firma datos con la clave privada
// Esto demuestra que TÚ autorizaste la transacción
func (kp *KeyPair) SignData(data []byte) (string, error) {
	// Hash de los datos
	hash := sha256.Sum256(data)

	// Firmar el hash con la clave privada
	r, s, err := ecdsa.Sign(rand.Reader, kp.PrivateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("error firmando: %v", err)
	}

	// Combinar r y s en una sola firma
	signature := append(r.Bytes(), s.Bytes()...)

	return hex.EncodeToString(signature), nil
}

// VerifySignature verifica que una firma sea válida
// Cualquiera puede verificar que TÚ firmaste, pero solo TÚ puedes firmar
func VerifySignature(publicKeyX, publicKeyY *big.Int, data []byte, signatureHex string) bool {
	// Reconstruir la clave pública
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     publicKeyX,
		Y:     publicKeyY,
	}

	// Decodificar la firma
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}

	// Separar r y s
	if len(signatureBytes) < 64 {
		return false
	}
	r := new(big.Int).SetBytes(signatureBytes[:32])
	s := new(big.Int).SetBytes(signatureBytes[32:64])

	// Hash de los datos
	hash := sha256.Sum256(data)

	// Verificar la firma
	return ecdsa.Verify(publicKey, hash[:], r, s)
}

// Print muestra información del par de claves
func (kp *KeyPair) Print() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║            PAR DE CLAVES               ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("🔑 Dirección:      %s\n", kp.GetAddress())
	fmt.Printf("🔐 Clave pública:  X=%s...\n", kp.PublicKey.X.Text(16)[:16])
	fmt.Printf("                   Y=%s...\n", kp.PublicKey.Y.Text(16)[:16])
	fmt.Println("⚠️  Clave privada: [OCULTA - Nunca compartir]")
}
