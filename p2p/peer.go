package p2p

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// Peer representa una conexión con otro nodo
type Peer struct {
	conn       net.Conn      // Conexión TCP
	address    string        // Dirección del peer (IP:Puerto)
	nodeID     string        // ID único del nodo remoto
	version    string        // Versión del protocolo que usa
	lastSeen   time.Time     // Última vez que recibimos algo
	bestHeight int           // Altura de su blockchain
	incoming   bool          // true si es conexión entrante, false si saliente
	quit       chan struct{} // Canal para cerrar el peer
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

// NewPeer crea un nuevo peer
func NewPeer(conn net.Conn, incoming bool) *Peer {
	return &Peer{
		conn:     conn,
		address:  conn.RemoteAddr().String(),
		incoming: incoming,
		lastSeen: time.Now(),
		quit:     make(chan struct{}),
	}
}

// String retorna una representación en string del peer
func (p *Peer) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	direction := "outgoing"
	if p.incoming {
		direction = "incoming"
	}

	return fmt.Sprintf("Peer{addr=%s, nodeID=%s, height=%d, %s}",
		p.address, p.nodeID[:8], p.bestHeight, direction)
}

// SendMessage envía un mensaje al peer
func (p *Peer) SendMessage(msg *Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Serializar mensaje
	data, err := msg.Encode()
	if err != nil {
		return fmt.Errorf("error codificando mensaje: %v", err)
	}

	// Enviar por la conexión TCP
	if _, err := p.conn.Write(data); err != nil {
		return fmt.Errorf("error enviando mensaje: %v", err)
	}

	return nil
}

// ReadMessage lee un mensaje del peer
func (p *Peer) ReadMessage() (*Message, error) {
	msg, err := DecodeMessage(p.conn)
	if err != nil {
		return nil, err
	}

	// Actualizar last seen
	p.mu.Lock()
	p.lastSeen = time.Now()
	p.mu.Unlock()

	return msg, nil
}

// SendHandshake envía handshake al peer
func (p *Peer) SendHandshake(data *HandshakeData) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error serializando handshake: %v", err)
	}

	msg := NewMessage(MsgHandshake, payload)
	return p.SendMessage(msg)
}

// SendPing envía un ping al peer
func (p *Peer) SendPing() error {
	msg := NewMessage(MsgPing, nil)
	return p.SendMessage(msg)
}

// SendPong envía un pong al peer
func (p *Peer) SendPong() error {
	msg := NewMessage(MsgPong, nil)
	return p.SendMessage(msg)
}

// SendBlockchainInfo envía información de blockchain al peer
func (p *Peer) SendBlockchainInfo(info *BlockchainInfo) error {
	payload, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("error serializando blockchain info: %v", err)
	}

	msg := NewMessage(MsgBlockchain, payload)
	return p.SendMessage(msg)
}

// UpdateInfo actualiza la información del peer
func (p *Peer) UpdateInfo(nodeID, version string, bestHeight int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.nodeID = nodeID
	p.version = version
	p.bestHeight = bestHeight
}

// GetBestHeight retorna la altura de blockchain del peer
func (p *Peer) GetBestHeight() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.bestHeight
}

// GetAddress retorna la dirección del peer
func (p *Peer) GetAddress() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.address
}

// GetNodeID retorna el ID del nodo
func (p *Peer) GetNodeID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.nodeID
}

// IsAlive verifica si el peer está vivo
func (p *Peer) IsAlive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Si no hemos recibido nada en 2 minutos, considerarlo muerto
	return time.Since(p.lastSeen) < 2*time.Minute
}

// Close cierra la conexión con el peer
func (p *Peer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.quit:
		// Ya cerrado
		return nil
	default:
		close(p.quit)
		return p.conn.Close()
	}
}

// IsClosed verifica si el peer está cerrado
func (p *Peer) IsClosed() bool {
	select {
	case <-p.quit:
		return true
	default:
		return false
	}
}
