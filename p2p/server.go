package p2p

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"minichain/blockchain"
	"net"
	"sync"
	"time"
)

// Server es el servidor P2P que gestiona todas las conexiones
type Server struct {
	host       string                  // IP donde escuchar
	port       int                     // Puerto donde escuchar
	listener   net.Listener            // Listener TCP
	blockchain *blockchain.Blockchain  // Referencia a la blockchain
	peers      map[string]*Peer        // Peers conectados (key: address)
	peersMu    sync.RWMutex            // Mutex para peers
	nodeID     string                  // ID único de este nodo
	networkID  uint64                  // ID de la red
	quit       chan struct{}           // Canal para cerrar el servidor
	wg         sync.WaitGroup          // WaitGroup para goroutines
	maxPeers   int                     // Número máximo de peers
	onNewBlock func(*blockchain.Block) // Callback cuando hay nuevo bloque
}

// NewServer crea un nuevo servidor P2P
func NewServer(host string, port int, bc *blockchain.Blockchain) *Server {
	// Generar ID único para este nodo
	nodeID := generateNodeID()

	return &Server{
		host:       host,
		port:       port,
		blockchain: bc,
		peers:      make(map[string]*Peer),
		nodeID:     nodeID,
		networkID:  1, // Red principal
		quit:       make(chan struct{}),
		maxPeers:   25, // Máximo 25 peers
	}
}

// generateNodeID genera un ID único para el nodo
func generateNodeID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Start inicia el servidor P2P
func (s *Server) Start() error {
	// Crear listener TCP
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("error iniciando listener: %v", err)
	}

	s.listener = listener

	log.Printf("🌐 Servidor P2P iniciado en %s (NodeID: %s)", addr, s.nodeID[:16])

	// Iniciar goroutine para aceptar conexiones
	s.wg.Add(1)
	go s.acceptLoop()

	// Iniciar goroutine para mantener peers vivos
	s.wg.Add(1)
	go s.keepAliveLoop()

	return nil
}

// acceptLoop acepta conexiones entrantes
func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.quit:
			return
		default:
		}

		// Aceptar nueva conexión
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Printf("⚠️  Error aceptando conexión: %v", err)
				continue
			}
		}

		// Verificar límite de peers
		if s.PeerCount() >= s.maxPeers {
			log.Printf("⚠️  Límite de peers alcanzado, rechazando %s", conn.RemoteAddr())
			conn.Close()
			continue
		}

		// Crear nuevo peer
		peer := NewPeer(conn, true)

		log.Printf("📥 Nueva conexión entrante desde %s", peer.GetAddress())

		// Manejar el peer en una nueva goroutine
		s.wg.Add(1)
		go s.handlePeer(peer)
	}
}

// ConnectToPeer se conecta a un peer remoto
func (s *Server) ConnectToPeer(address string) error {
	// Verificar si ya estamos conectados
	if s.isPeerConnected(address) {
		return fmt.Errorf("ya conectado a %s", address)
	}

	// Conectar
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error conectando a %s: %v", address, err)
	}

	// Crear peer
	peer := NewPeer(conn, false)

	log.Printf("📤 Conectado a peer %s", address)

	// Manejar el peer
	s.wg.Add(1)
	go s.handlePeer(peer)

	return nil
}

// handlePeer maneja la comunicación con un peer
func (s *Server) handlePeer(peer *Peer) {
	defer s.wg.Done()
	defer peer.Close()

	// Realizar handshake
	if err := s.performHandshake(peer); err != nil {
		log.Printf("❌ Error en handshake con %s: %v", peer.GetAddress(), err)
		return
	}

	// Agregar peer a la lista
	s.addPeer(peer)
	defer s.removePeer(peer)

	log.Printf("✅ Peer conectado: %s", peer)

	// Loop principal de mensajes
	for {
		select {
		case <-s.quit:
			return
		case <-peer.quit:
			return
		default:
		}

		// Leer mensaje
		msg, err := peer.ReadMessage()
		if err != nil {
			if !peer.IsClosed() {
				log.Printf("⚠️  Error leyendo de %s: %v", peer.GetAddress(), err)
			}
			return
		}

		// Procesar mensaje
		if err := s.handleMessage(peer, msg); err != nil {
			log.Printf("⚠️  Error procesando mensaje de %s: %v", peer.GetAddress(), err)
		}
	}
}

// performHandshake realiza el handshake con un peer
func (s *Server) performHandshake(peer *Peer) error {
	// Enviar nuestro handshake
	myHandshake := &HandshakeData{
		Version:        ProtocolVersion,
		NetworkID:      s.networkID,
		BestBlockIndex: len(s.blockchain.Blocks) - 1,
		BestBlockHash:  s.blockchain.Blocks[len(s.blockchain.Blocks)-1].Hash,
		NodeID:         s.nodeID,
		ListenPort:     s.port,
	}

	if err := peer.SendHandshake(myHandshake); err != nil {
		return fmt.Errorf("error enviando handshake: %v", err)
	}

	// Esperar handshake del peer (timeout 10 segundos)
	peer.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	msg, err := peer.ReadMessage()
	peer.conn.SetReadDeadline(time.Time{}) // Quitar deadline

	if err != nil {
		return fmt.Errorf("error recibiendo handshake: %v", err)
	}

	if msg.Type != MsgHandshake {
		return fmt.Errorf("esperaba handshake, recibí %s", msg.Type)
	}

	// Decodificar handshake
	var theirHandshake HandshakeData
	if err := json.Unmarshal(msg.Payload, &theirHandshake); err != nil {
		return fmt.Errorf("error decodificando handshake: %v", err)
	}

	// Verificar versión y network ID
	if theirHandshake.Version != ProtocolVersion {
		return fmt.Errorf("versión incompatible: %s (esperada: %s)",
			theirHandshake.Version, ProtocolVersion)
	}

	if theirHandshake.NetworkID != s.networkID {
		return fmt.Errorf("network ID diferente: %d (esperada: %d)",
			theirHandshake.NetworkID, s.networkID)
	}

	// Actualizar info del peer
	peer.UpdateInfo(theirHandshake.NodeID, theirHandshake.Version, theirHandshake.BestBlockIndex)

	return nil
}

// handleMessage procesa un mensaje recibido
func (s *Server) handleMessage(peer *Peer, msg *Message) error {
	switch msg.Type {
	case MsgPing:
		// Responder con pong
		return peer.SendPong()

	case MsgPong:
		// Pong recibido, peer está vivo
		return nil

	case MsgGetBlockchain:
		// Enviar info de nuestra blockchain
		info := &BlockchainInfo{
			Height:         len(s.blockchain.Blocks),
			BestBlockHash:  s.blockchain.Blocks[len(s.blockchain.Blocks)-1].Hash,
			BestBlockIndex: len(s.blockchain.Blocks) - 1,
			Difficulty:     s.blockchain.Difficulty,
		}
		return peer.SendBlockchainInfo(info)

	case MsgBlockchain:
		// Recibido info de blockchain del peer
		var info BlockchainInfo
		if err := json.Unmarshal(msg.Payload, &info); err != nil {
			return fmt.Errorf("error decodificando blockchain info: %v", err)
		}

		log.Printf("📊 Peer %s tiene blockchain con altura %d", peer.GetAddress()[:15], info.Height)

		// Actualizar altura del peer
		peer.mu.Lock()
		peer.bestHeight = info.BestBlockIndex
		peer.mu.Unlock()

		// TODO: Si su blockchain es más larga, sincronizar
		return nil

	case MsgNewBlock:
		// Recibido nuevo bloque
		// TODO: Deserializar y validar bloque
		log.Printf("📦 Nuevo bloque recibido de %s", peer.GetAddress())
		return nil

	case MsgNewTransaction:
		// Recibida nueva transacción
		// TODO: Deserializar y agregar al mempool
		log.Printf("💸 Nueva transacción recibida de %s", peer.GetAddress())
		return nil

	default:
		log.Printf("⚠️  Mensaje desconocido: %s", msg.Type)
		return nil
	}
}

// keepAliveLoop envía pings periódicos a los peers
func (s *Server) keepAliveLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.quit:
			return
		case <-ticker.C:
			s.peersMu.RLock()
			peers := make([]*Peer, 0, len(s.peers))
			for _, peer := range s.peers {
				peers = append(peers, peer)
			}
			s.peersMu.RUnlock()

			// Enviar ping a cada peer
			for _, peer := range peers {
				if !peer.IsAlive() {
					log.Printf("⚠️  Peer %s no responde, desconectando...", peer.GetAddress())
					peer.Close()
				} else {
					peer.SendPing()
				}
			}
		}
	}
}

// BroadcastBlockchainInfo solicita info de blockchain a todos los peers
func (s *Server) BroadcastBlockchainInfo() {
	msg := NewMessage(MsgGetBlockchain, nil)

	s.peersMu.RLock()
	defer s.peersMu.RUnlock()

	for _, peer := range s.peers {
		if err := peer.SendMessage(msg); err != nil {
			log.Printf("⚠️  Error enviando mensaje a %s: %v", peer.GetAddress(), err)
		}
	}
}

// addPeer agrega un peer a la lista
func (s *Server) addPeer(peer *Peer) {
	s.peersMu.Lock()
	defer s.peersMu.Unlock()
	s.peers[peer.GetAddress()] = peer
}

// removePeer elimina un peer de la lista
func (s *Server) removePeer(peer *Peer) {
	s.peersMu.Lock()
	defer s.peersMu.Unlock()
	delete(s.peers, peer.GetAddress())
	log.Printf("👋 Peer desconectado: %s", peer.GetAddress())
}

// isPeerConnected verifica si ya estamos conectados a un peer
func (s *Server) isPeerConnected(address string) bool {
	s.peersMu.RLock()
	defer s.peersMu.RUnlock()
	_, ok := s.peers[address]
	return ok
}

// PeerCount retorna el número de peers conectados
func (s *Server) PeerCount() int {
	s.peersMu.RLock()
	defer s.peersMu.RUnlock()
	return len(s.peers)
}

// GetPeers retorna una lista de peers conectados
func (s *Server) GetPeers() []*Peer {
	s.peersMu.RLock()
	defer s.peersMu.RUnlock()

	peers := make([]*Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	return peers
}

// Stop detiene el servidor P2P
func (s *Server) Stop() error {
	log.Println("🛑 Deteniendo servidor P2P...")

	// Cerrar canal quit
	close(s.quit)

	// Cerrar listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Cerrar todos los peers
	s.peersMu.Lock()
	for _, peer := range s.peers {
		peer.Close()
	}
	s.peersMu.Unlock()

	// Esperar a que terminen todas las goroutines
	s.wg.Wait()

	log.Println("✅ Servidor P2P detenido")

	return nil
}
