package p2p

// Versión del protocolo P2P de Minichain
const (
	ProtocolVersion = "1.0.0"
	ProtocolName    = "minichain"
	MaxMessageSize  = 10 * 1024 * 1024 // 10 MB
)

// Tipos de mensajes en el protocolo P2P
type MessageType uint8

const (
	// Mensajes de control
	MsgHandshake MessageType = 0x00 // Saludo inicial entre peers
	MsgPing      MessageType = 0x01 // Keep-alive
	MsgPong      MessageType = 0x02 // Respuesta a ping

	// Mensajes de blockchain
	MsgNewBlock       MessageType = 0x10 // Propagar nuevo bloque minado
	MsgNewTransaction MessageType = 0x11 // Propagar nueva transacción

	// Mensajes de sincronización
	MsgGetBlocks     MessageType = 0x20 // Solicitar bloques
	MsgBlocks        MessageType = 0x21 // Enviar bloques
	MsgGetBlockchain MessageType = 0x22 // Solicitar info de blockchain
	MsgBlockchain    MessageType = 0x23 // Enviar info de blockchain

	// Mensajes de peers
	MsgGetPeers MessageType = 0x30 // Solicitar lista de peers
	MsgPeers    MessageType = 0x31 // Enviar lista de peers
)

// String retorna el nombre del tipo de mensaje
func (m MessageType) String() string {
	switch m {
	case MsgHandshake:
		return "Handshake"
	case MsgPing:
		return "Ping"
	case MsgPong:
		return "Pong"
	case MsgNewBlock:
		return "NewBlock"
	case MsgNewTransaction:
		return "NewTransaction"
	case MsgGetBlocks:
		return "GetBlocks"
	case MsgBlocks:
		return "Blocks"
	case MsgGetBlockchain:
		return "GetBlockchain"
	case MsgBlockchain:
		return "Blockchain"
	case MsgGetPeers:
		return "GetPeers"
	case MsgPeers:
		return "Peers"
	default:
		return "Unknown"
	}
}

// Códigos de error
type ErrorCode uint8

const (
	ErrInvalidMessage ErrorCode = 0x01
	ErrInvalidVersion ErrorCode = 0x02
	ErrTooLarge       ErrorCode = 0x03
	ErrTimeout        ErrorCode = 0x04
	ErrInvalidBlock   ErrorCode = 0x05
)
