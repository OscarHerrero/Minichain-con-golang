package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Message representa un mensaje P2P
type Message struct {
	Type    MessageType // Tipo de mensaje
	Payload []byte      // Datos del mensaje
}

// NewMessage crea un nuevo mensaje
func NewMessage(msgType MessageType, payload []byte) *Message {
	return &Message{
		Type:    msgType,
		Payload: payload,
	}
}

// Encode serializa el mensaje para envío por red
// Formato: [1 byte tipo][4 bytes longitud][N bytes payload]
func (m *Message) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Escribir tipo de mensaje (1 byte)
	if err := binary.Write(buf, binary.BigEndian, m.Type); err != nil {
		return nil, fmt.Errorf("error escribiendo tipo: %v", err)
	}

	// Escribir longitud del payload (4 bytes)
	payloadLen := uint32(len(m.Payload))
	if payloadLen > MaxMessageSize {
		return nil, fmt.Errorf("mensaje demasiado grande: %d bytes (máximo: %d)", payloadLen, MaxMessageSize)
	}

	if err := binary.Write(buf, binary.BigEndian, payloadLen); err != nil {
		return nil, fmt.Errorf("error escribiendo longitud: %v", err)
	}

	// Escribir payload
	if _, err := buf.Write(m.Payload); err != nil {
		return nil, fmt.Errorf("error escribiendo payload: %v", err)
	}

	return buf.Bytes(), nil
}

// DecodeMessage lee un mensaje desde un reader
func DecodeMessage(r io.Reader) (*Message, error) {
	msg := &Message{}

	// Leer tipo de mensaje (1 byte)
	if err := binary.Read(r, binary.BigEndian, &msg.Type); err != nil {
		return nil, fmt.Errorf("error leyendo tipo: %v", err)
	}

	// Leer longitud del payload (4 bytes)
	var payloadLen uint32
	if err := binary.Read(r, binary.BigEndian, &payloadLen); err != nil {
		return nil, fmt.Errorf("error leyendo longitud: %v", err)
	}

	// Validar longitud
	if payloadLen > MaxMessageSize {
		return nil, fmt.Errorf("mensaje demasiado grande: %d bytes", payloadLen)
	}

	// Leer payload
	if payloadLen > 0 {
		msg.Payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(r, msg.Payload); err != nil {
			return nil, fmt.Errorf("error leyendo payload: %v", err)
		}
	}

	return msg, nil
}

// String retorna una representación en string del mensaje
func (m *Message) String() string {
	return fmt.Sprintf("Message{Type: %s, PayloadSize: %d bytes}", m.Type, len(m.Payload))
}

// HandshakeData contiene la información del handshake inicial
type HandshakeData struct {
	Version        string // Versión del protocolo
	NetworkID      uint64 // ID de la red (para distinguir redes)
	BestBlockIndex int    // Altura de la blockchain
	BestBlockHash  string // Hash del mejor bloque
	NodeID         string // ID único del nodo
	ListenPort     int    // Puerto donde escucha este nodo
}

// BlockchainInfo contiene información sobre el estado de la blockchain
type BlockchainInfo struct {
	Height         int    // Número de bloques
	BestBlockHash  string // Hash del último bloque
	BestBlockIndex int    // Índice del último bloque
	Difficulty     int    // Dificultad actual
}

// PeerInfo contiene información sobre un peer
type PeerInfo struct {
	Address    string // IP:Puerto
	NodeID     string // ID único del nodo
	Version    string // Versión del protocolo
	LastSeen   int64  // Timestamp de última comunicación
	BestHeight int    // Altura de su blockchain
}
