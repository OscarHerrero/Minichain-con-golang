package p2p

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"minichain/blockchain"
	"minichain/core/rawdb"
	"net"
	"strings"
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

	// Control de minado
	mining      bool       // Si este nodo está minando
	miningMu    sync.Mutex // Mutex para controlar minado
	stopMining  chan struct{}
	newBlockCh  chan *blockchain.Block // Canal para notificar bloques nuevos

	// Cache de transacciones vistas (para evitar loops de propagación)
	seenTxs   map[string]bool // Hash de transacción -> visto
	seenTxsMu sync.RWMutex    // Mutex para seenTxs
}

// truncateAddr trunca una dirección de forma segura para logging
func truncateAddr(addr string, maxLen int) string {
	if len(addr) <= maxLen {
		return addr
	}
	return addr[:maxLen] + "..."
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
		stopMining: make(chan struct{}),
		newBlockCh: make(chan *blockchain.Block, 10),
		seenTxs:    make(map[string]bool),
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

	log.Printf("🌐 Servidor P2P iniciado en %s (NodeID: %s)", addr, truncateAddr(s.nodeID, 16))

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

	// Sincronización automática: Si el peer tiene una cadena más larga, sincronizar
	ourHeight := len(s.blockchain.Blocks) - 1
	theirHeight := theirHandshake.BestBlockIndex

	if theirHeight > ourHeight {
		log.Printf("🔄 Peer %s tiene cadena más larga (%d vs %d) - iniciando sincronización...",
			truncateAddr(peer.GetAddress(), 20), theirHeight, ourHeight)
		s.requestBlockchainFrom(peer, ourHeight+1)
	} else if ourHeight > theirHeight {
		log.Printf("📤 Nuestra cadena es más larga (%d vs %d) - peer se sincronizará con nosotros",
			ourHeight, theirHeight)
	} else {
		log.Printf("✅ Blockchains sincronizadas (altura: %d)", ourHeight)
	}

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
		// Enviar toda nuestra blockchain
		log.Printf("📤 Enviando blockchain completa a %s (%d bloques)...",
			truncateAddr(peer.GetAddress(), 20), len(s.blockchain.Blocks))

		// Serializar todos los bloques
		blocksData, err := json.Marshal(s.blockchain.Blocks)
		if err != nil {
			return fmt.Errorf("error serializando blockchain: %v", err)
		}

		// Enviar con MsgBlocks
		msg := NewMessage(MsgBlocks, blocksData)
		return peer.SendMessage(msg)

	case MsgBlockchain:
		// Recibido info de blockchain del peer
		var info BlockchainInfo
		if err := json.Unmarshal(msg.Payload, &info); err != nil {
			return fmt.Errorf("error decodificando blockchain info: %v", err)
		}

		log.Printf("📊 Peer %s tiene blockchain con altura %d", truncateAddr(peer.GetAddress(), 20), info.Height)

		// Actualizar altura del peer
		peer.mu.Lock()
		peer.bestHeight = info.BestBlockIndex
		peer.mu.Unlock()

		// TODO: Si su blockchain es más larga, sincronizar
		return nil

	case MsgNewBlock:
		// Recibido nuevo bloque
		var newBlock blockchain.Block
		if err := json.Unmarshal(msg.Payload, &newBlock); err != nil {
			return fmt.Errorf("error decodificando bloque: %v", err)
		}

		log.Printf("📦 Nuevo bloque recibido de %s: Bloque #%d", peer.GetAddress(), newBlock.Index)

		// Procesar el bloque
		return s.handleNewBlock(&newBlock, peer)

	case MsgNewTransaction:
		// Recibida nueva transacción
		var tx blockchain.Transaction
		if err := json.Unmarshal(msg.Payload, &tx); err != nil {
			return fmt.Errorf("error decodificando transacción: %v", err)
		}

		log.Printf("💸 Nueva transacción recibida de %s: %s → %s (%.2f MTC)",
			peer.GetAddress(), tx.From, tx.To, tx.Amount)

		// Calcular hash para verificar si ya la vimos
		txHash := calculateTxHash(&tx)

		s.seenTxsMu.Lock()
		alreadySeen := s.seenTxs[txHash]
		if !alreadySeen {
			s.seenTxs[txHash] = true
		}
		s.seenTxsMu.Unlock()

		if alreadySeen {
			// Ya vimos esta transacción, no hacer nada
			return nil
		}

		// Agregar al mempool
		s.blockchain.PendingTxs = append(s.blockchain.PendingTxs, &tx)

		log.Printf("   ✅ Transacción agregada al mempool (total: %d pendientes)", len(s.blockchain.PendingTxs))

		// Propagar a otros peers (excepto el que nos la envió)
		s.BroadcastTransactionExcept(&tx, peer)

		return nil

	case MsgBlocks:
		// Recibida blockchain completa
		var blocks []*blockchain.Block
		if err := json.Unmarshal(msg.Payload, &blocks); err != nil {
			return fmt.Errorf("error decodificando bloques: %v", err)
		}

		log.Printf("📥 Blockchain recibida de %s (%d bloques)",
			truncateAddr(peer.GetAddress(), 20), len(blocks))

		// Intentar reemplazar nuestra cadena con la recibida
		s.replaceChain(blocks)

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

// StartMining inicia el minado continuo estilo Ethereum
func (s *Server) StartMining() {
	s.miningMu.Lock()
	if s.mining {
		s.miningMu.Unlock()
		log.Println("⚠️  El minado ya está activo")
		return
	}
	s.mining = true
	s.miningMu.Unlock()

	log.Println("⛏️  Minado continuo iniciado")

	// Iniciar goroutine de minado
	s.wg.Add(1)
	go s.miningLoop()
}

// StopMining detiene el minado continuo
func (s *Server) StopMining() {
	s.miningMu.Lock()
	if !s.mining {
		s.miningMu.Unlock()
		return
	}
	s.mining = false
	s.miningMu.Unlock()

	// Enviar señal de stop
	select {
	case s.stopMining <- struct{}{}:
	default:
	}

	log.Println("🛑 Minado continuo detenido")
}

// miningLoop es el bucle principal de minado continuo
// Mina un bloque cada segundo (con o sin transacciones)
func (s *Server) miningLoop() {
	defer s.wg.Done()

	// Ticker para minar cada segundo
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Verificar si debemos seguir minando
			s.miningMu.Lock()
			shouldMine := s.mining
			s.miningMu.Unlock()

			if !shouldMine {
				return
			}

			// Contar transacciones pendientes
			txCount := len(s.blockchain.PendingTxs)

			log.Printf("⛏️  Iniciando minado de bloque %d (%d transacciones)...\n",
				len(s.blockchain.Blocks), txCount)

			// Intentar minar el bloque con posibilidad de interrupción
			block := s.mineBlockWithCancellation()

			if block != nil {
				// ¡Bloque minado exitosamente!
				log.Printf("✅ Bloque %d minado exitosamente! Hash: %s (txs: %d)\n",
					block.Index, truncateAddr(block.Hash, 16), len(block.Transactions))

				// Propagar el bloque a todos los peers
				s.BroadcastBlock(block)

				// Notificar callback si existe
				if s.onNewBlock != nil {
					s.onNewBlock(block)
				}
			}

		case <-s.quit:
			return
		}
	}
}

// mineBlockWithCancellation mina un bloque con la posibilidad de cancelación
func (s *Server) mineBlockWithCancellation() *blockchain.Block {
	// Preparar el bloque
	prevBlock := s.blockchain.Blocks[len(s.blockchain.Blocks)-1]

	// Copiar transacciones pendientes para este bloque
	// (puede ser un slice vacío si no hay transacciones)
	txs := make([]*blockchain.Transaction, len(s.blockchain.PendingTxs))
	copy(txs, s.blockchain.PendingTxs)

	newBlock := &blockchain.Block{
		Index:        len(s.blockchain.Blocks),
		Timestamp:    time.Now(),
		Transactions: txs,
		PreviousHash: prevBlock.Hash,
		Nonce:        0,
	}

	// Ejecutar transacciones (sin StateDB completo por ahora)
	// TODO: Ejecutar transacciones y calcular state roots

	// Inicializar roots
	newBlock.StateRoot = make([]byte, 32)
	newBlock.TxRoot = make([]byte, 32)
	newBlock.ReceiptRoot = make([]byte, 32)

	// Minar con posibilidad de cancelación
	success := s.mineWithCancellation(newBlock, s.blockchain.Difficulty)

	if !success {
		// Minado cancelado (nuevo bloque recibido)
		log.Println("⚠️  Minado cancelado - nuevo bloque recibido")
		return nil
	}

	// Agregar bloque a la cadena
	s.blockchain.Blocks = append(s.blockchain.Blocks, newBlock)

	// Limpiar transacciones pendientes
	s.blockchain.PendingTxs = []*blockchain.Transaction{}

	// Persistir bloque en base de datos
	if s.blockchain != nil && s.blockchain.GetDB() != nil {
		db := s.blockchain.GetDB()

		// Convertir el bloque al formato de persistencia
		header := s.blockchain.ConvertBlockToHeader(newBlock)
		body := s.blockchain.ConvertBlockToBody(newBlock)

		// Guardar en la base de datos
		if err := rawdb.WriteBlock(db, header, body); err != nil {
			log.Printf("⚠️  Error persistiendo bloque: %v\n", err)
		} else {
			// Convertir hash hex a bytes
			hashBytes, err := hex.DecodeString(newBlock.Hash)
			if err == nil {
				// Escribir hash canónico (altura -> hash)
				rawdb.WriteCanonicalHash(db, hashBytes, uint64(newBlock.Index))
				// Actualizar head block
				rawdb.WriteHeadBlockHash(db, hashBytes)
				log.Printf("💾 Bloque %d persistido en disco\n", newBlock.Index)
			}
		}
	}

	return newBlock
}

// mineWithCancellation realiza el minado con cancelación
func (s *Server) mineWithCancellation(block *blockchain.Block, difficulty int) bool {
	target := strings.Repeat("0", difficulty)

	for {
		// Verificar si hay señal de cancelación
		select {
		case <-s.stopMining:
			return false
		case <-s.newBlockCh:
			// Nuevo bloque recibido, cancelar minado
			return false
		default:
			// Continuar minando
		}

		// Calcular hash
		block.Hash = block.CalculateBlockHash()

		// ¿Cumple con la dificultad?
		if strings.HasPrefix(block.Hash, target) {
			// ¡Encontrado!
			return true
		}

		// Incrementar nonce
		block.Nonce++

		// Pequeña pausa cada 10000 intentos para permitir cancelación
		if block.Nonce%10000 == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// BroadcastBlock propaga un bloque a todos los peers
func (s *Server) BroadcastBlock(block *blockchain.Block) {
	// Serializar bloque a JSON
	blockData, err := json.Marshal(block)
	if err != nil {
		log.Printf("❌ Error serializando bloque: %v", err)
		return
	}

	msg := NewMessage(MsgNewBlock, blockData)

	s.peersMu.RLock()
	defer s.peersMu.RUnlock()

	log.Printf("📡 Propagando bloque %d a %d peers...", block.Index, len(s.peers))

	for _, peer := range s.peers {
		if err := peer.SendMessage(msg); err != nil {
			log.Printf("⚠️  Error enviando bloque a %s: %v", peer.GetAddress(), err)
		}
	}
}

// IsMining retorna si el nodo está minando actualmente
func (s *Server) IsMining() bool {
	s.miningMu.Lock()
	defer s.miningMu.Unlock()
	return s.mining
}

// handleNewBlock procesa un bloque recibido de un peer
func (s *Server) handleNewBlock(newBlock *blockchain.Block, peer *Peer) error {
	// 1. Verificar que el bloque es válido
	if !newBlock.IsValid(s.blockchain.Difficulty) {
		log.Printf("❌ Bloque #%d inválido - rechazado", newBlock.Index)
		return fmt.Errorf("bloque inválido")
	}

	// 2. Obtener altura actual de nuestra cadena
	currentHeight := len(s.blockchain.Blocks) - 1

	// 3. Verificar qué tipo de bloque es
	if newBlock.Index == currentHeight+1 {
		// ✅ Es el siguiente bloque en nuestra cadena
		lastBlock := s.blockchain.Blocks[currentHeight]

		// Verificar que el PreviousHash coincide
		if newBlock.PreviousHash != lastBlock.Hash {
			log.Printf("❌ Bloque #%d rechazado - PreviousHash no coincide", newBlock.Index)
			return fmt.Errorf("previousHash no coincide")
		}

		log.Printf("✅ Bloque #%d válido - agregando a la cadena", newBlock.Index)

		// Cancelar minado actual
		select {
		case s.newBlockCh <- newBlock:
		default:
		}

		// Ejecutar transacciones del bloque
		log.Printf("💼 Ejecutando %d transacciones del bloque...", len(newBlock.Transactions))
		for i, tx := range newBlock.Transactions {
			if err := tx.Execute(s.blockchain.AccountState, s.blockchain); err != nil {
				log.Printf("   ⚠️  Error ejecutando tx %d: %v", i, err)
			}
		}

		// Agregar bloque a nuestra cadena
		s.blockchain.Blocks = append(s.blockchain.Blocks, newBlock)

		// Propagar a otros peers (evitando el que nos lo envió)
		s.BroadcastBlockExcept(newBlock, peer)

		log.Printf("📊 Blockchain actualizada - altura: %d", len(s.blockchain.Blocks)-1)

		return nil

	} else if newBlock.Index <= currentHeight {
		// Bloque antiguo o duplicado - ignorar
		log.Printf("⚠️  Bloque #%d ignorado - ya lo tenemos", newBlock.Index)
		return nil

	} else {
		// newBlock.Index > currentHeight+1
		// El peer tiene una cadena más larga - necesitamos sincronizar
		log.Printf("🔄 Peer %s tiene cadena más larga (altura: %d, nosotros: %d)",
			truncateAddr(peer.GetAddress(), 20), newBlock.Index, currentHeight)

		// Solicitar toda la blockchain del peer
		log.Printf("   📥 Solicitando blockchain completa desde altura %d...", currentHeight+1)
		s.requestBlockchainFrom(peer, currentHeight+1)

		return nil
	}
}

// BroadcastBlockExcept propaga un bloque a todos los peers excepto uno
func (s *Server) BroadcastBlockExcept(block *blockchain.Block, except *Peer) {
	// Serializar bloque a JSON
	blockData, err := json.Marshal(block)
	if err != nil {
		log.Printf("❌ Error serializando bloque: %v", err)
		return
	}

	msg := NewMessage(MsgNewBlock, blockData)

	s.peersMu.RLock()
	defer s.peersMu.RUnlock()

	propagatedCount := 0
	for _, peer := range s.peers {
		// Saltar el peer que nos envió el bloque
		if except != nil && peer.GetAddress() == except.GetAddress() {
			continue
		}

		if err := peer.SendMessage(msg); err != nil {
			log.Printf("⚠️  Error enviando bloque a %s: %v", peer.GetAddress(), err)
		} else {
			propagatedCount++
		}
	}

	if propagatedCount > 0 {
		log.Printf("📡 Bloque #%d propagado a %d peers adicionales", block.Index, propagatedCount)
	}
}

// requestBlockchainFrom solicita la blockchain completa desde una altura específica
func (s *Server) requestBlockchainFrom(peer *Peer, fromHeight int) {
	// Enviar mensaje MsgGetBlockchain
	msg := NewMessage(MsgGetBlockchain, nil)

	if err := peer.SendMessage(msg); err != nil {
		log.Printf("❌ Error solicitando blockchain: %v", err)
	}
}

// replaceChain reemplaza nuestra blockchain si la nueva es más larga y válida
func (s *Server) replaceChain(newBlocks []*blockchain.Block) bool {
	// 1. Verificar que la nueva cadena es más larga
	if len(newBlocks) <= len(s.blockchain.Blocks) {
		log.Printf("⚠️  Nueva cadena no es más larga - rechazada")
		return false
	}

	// 2. Validar toda la cadena
	log.Printf("🔍 Validando cadena recibida (%d bloques)...", len(newBlocks))

	// Verificar bloque génesis
	if newBlocks[0].Index != 0 {
		log.Printf("❌ Cadena inválida - primer bloque no es génesis")
		return false
	}

	// Verificar cada bloque y sus enlaces
	for i := 0; i < len(newBlocks); i++ {
		block := newBlocks[i]

		// Verificar que el bloque es válido
		if !block.IsValid(s.blockchain.Difficulty) {
			log.Printf("❌ Bloque #%d es inválido", i)
			return false
		}

		// Verificar enlaces (excepto el génesis)
		if i > 0 {
			prevBlock := newBlocks[i-1]
			if block.PreviousHash != prevBlock.Hash {
				log.Printf("❌ Cadena rota en bloque #%d", i)
				return false
			}
		}
	}

	log.Printf("✅ Cadena válida - reemplazando (longitud: %d → %d bloques)",
		len(s.blockchain.Blocks), len(newBlocks))

	// 3. Cancelar minado actual
	s.StopMining()

	// 4. Reemplazar la blockchain
	s.blockchain.Blocks = newBlocks

	// 5. Re-ejecutar todas las transacciones para reconstruir el estado
	log.Printf("💼 Re-ejecutando transacciones para reconstruir estado...")
	s.blockchain.AccountState = blockchain.NewAccountState()

	totalTxs := 0
	for i, block := range newBlocks {
		if i == 0 {
			continue // Saltar génesis
		}
		for _, tx := range block.Transactions {
			if err := tx.Execute(s.blockchain.AccountState, s.blockchain); err != nil {
				log.Printf("   ⚠️  Error re-ejecutando tx en bloque #%d: %v", i, err)
			}
			totalTxs++
		}
	}
	log.Printf("✅ Estado reconstruido (%d transacciones procesadas)", totalTxs)

	// 6. Limpiar transacciones pendientes que ya están en bloques
	// TODO: Implementar lógica más sofisticada para mantener TXs no minadas
	s.blockchain.PendingTxs = []*blockchain.Transaction{}

	// 7. Reiniciar minado
	s.StartMining()

	log.Printf("🎉 Blockchain reemplazada exitosamente - nueva altura: %d", len(s.blockchain.Blocks)-1)

	return true
}

// calculateTxHash calcula un hash simple de una transacción
func calculateTxHash(tx *blockchain.Transaction) string {
	data := fmt.Sprintf("%s:%s:%.2f:%d", tx.From, tx.To, tx.Amount, tx.Nonce)
	return fmt.Sprintf("%x", []byte(data))
}

// BroadcastTransaction propaga una transacción a todos los peers
func (s *Server) BroadcastTransaction(tx *blockchain.Transaction) {
	// Calcular hash de la transacción
	txHash := calculateTxHash(tx)

	// Verificar si ya vimos esta transacción
	s.seenTxsMu.Lock()
	if s.seenTxs[txHash] {
		s.seenTxsMu.Unlock()
		return // Ya la vimos, no propagar
	}
	// Marcar como vista
	s.seenTxs[txHash] = true
	s.seenTxsMu.Unlock()

	// Serializar transacción a JSON
	txData, err := json.Marshal(tx)
	if err != nil {
		log.Printf("❌ Error serializando transacción: %v", err)
		return
	}

	msg := NewMessage(MsgNewTransaction, txData)

	s.peersMu.RLock()
	defer s.peersMu.RUnlock()

	propagatedCount := 0
	for _, peer := range s.peers {
		if err := peer.SendMessage(msg); err != nil {
			log.Printf("⚠️  Error enviando transacción a %s: %v", peer.GetAddress(), err)
		} else {
			propagatedCount++
		}
	}

	if propagatedCount > 0 {
		log.Printf("📡 Transacción propagada a %d peers", propagatedCount)
	}
}

// BroadcastTransactionExcept propaga una transacción a todos los peers excepto uno
func (s *Server) BroadcastTransactionExcept(tx *blockchain.Transaction, except *Peer) {
	// Calcular hash de la transacción
	txHash := calculateTxHash(tx)

	// Verificar si ya vimos esta transacción
	s.seenTxsMu.Lock()
	if s.seenTxs[txHash] {
		s.seenTxsMu.Unlock()
		return // Ya la vimos, no propagar
	}
	// Marcar como vista
	s.seenTxs[txHash] = true
	s.seenTxsMu.Unlock()

	// Serializar transacción a JSON
	txData, err := json.Marshal(tx)
	if err != nil {
		log.Printf("❌ Error serializando transacción: %v", err)
		return
	}

	msg := NewMessage(MsgNewTransaction, txData)

	s.peersMu.RLock()
	defer s.peersMu.RUnlock()

	propagatedCount := 0
	for _, peer := range s.peers {
		// Saltar el peer que nos envió la transacción
		if except != nil && peer.GetAddress() == except.GetAddress() {
			continue
		}

		if err := peer.SendMessage(msg); err != nil {
			log.Printf("⚠️  Error enviando transacción a %s: %v", peer.GetAddress(), err)
		} else {
			propagatedCount++
		}
	}

	if propagatedCount > 0 {
		log.Printf("📡 Transacción propagada a %d peers adicionales", propagatedCount)
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
