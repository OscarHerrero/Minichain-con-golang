package p2p

import (
	"encoding/json"
	"fmt"
	"log"
	"minichain/blockchain"
	"net/http"
)

// RPCServer es un servidor HTTP simple para RPC
type RPCServer struct {
	port       int
	blockchain *blockchain.Blockchain
	server     *Server
}

// NewRPCServer crea un nuevo servidor RPC
func NewRPCServer(port int, bc *blockchain.Blockchain, p2pServer *Server) *RPCServer {
	return &RPCServer{
		port:       port,
		blockchain: bc,
		server:     p2pServer,
	}
}

// Start inicia el servidor RPC
func (rpc *RPCServer) Start() error {
	// Endpoint para enviar transacciones
	http.HandleFunc("/tx", rpc.handleTransaction)

	// Endpoint para obtener estado de la blockchain
	http.HandleFunc("/status", rpc.handleStatus)

	// Endpoint de health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%d", rpc.port)
	log.Printf("🌐 Servidor RPC iniciado en http://localhost%s", addr)
	log.Println("   Endpoints disponibles:")
	log.Println("   - POST /tx       (Enviar transacción)")
	log.Println("   - GET  /status   (Estado de la blockchain)")
	log.Println("   - GET  /health   (Health check)")

	return http.ListenAndServe(addr, nil)
}

// TxRequest es la estructura de una transacción recibida por RPC
type TxRequest struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Data   string  `json:"data"`
}

// handleTransaction maneja el endpoint POST /tx
func (rpc *RPCServer) handleTransaction(w http.ResponseWriter, r *http.Request) {
	// Solo aceptar POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido. Usa POST", http.StatusMethodNotAllowed)
		return
	}

	// Parsear JSON
	var txReq TxRequest
	if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
		http.Error(w, fmt.Sprintf("Error parseando JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Validar campos básicos
	if txReq.From == "" {
		http.Error(w, "Campo 'from' es requerido", http.StatusBadRequest)
		return
	}

	// Crear transacción
	tx := &blockchain.Transaction{
		From:   txReq.From,
		To:     txReq.To,
		Amount: txReq.Amount,
		Nonce:  0, // TODO: Calcular nonce correcto
		Data:   []byte{},
	}

	// Parsear data si existe
	if txReq.Data != "" {
		// TODO: Convertir hex string a bytes
		tx.Data = []byte(txReq.Data)
	}

	// Agregar al mempool (sin validación para simplificar por ahora)
	rpc.blockchain.PendingTxs = append(rpc.blockchain.PendingTxs, tx)

	log.Printf("📥 Transacción recibida por RPC: %s → %s (%.2f MTC)",
		txReq.From, txReq.To, txReq.Amount)

	// Propagar la transacción a todos los peers
	rpc.server.BroadcastTransaction(tx)

	// Responder con éxito
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":  "success",
		"message": "Transacción agregada al mempool",
		"txCount": len(rpc.blockchain.PendingTxs),
	}

	json.NewEncoder(w).Encode(response)
}

// StatusResponse es la respuesta del endpoint /status
type StatusResponse struct {
	Blocks        int    `json:"blocks"`
	LastBlockHash string `json:"lastBlockHash"`
	PendingTxs    int    `json:"pendingTxs"`
	Peers         int    `json:"peers"`
	Mining        bool   `json:"mining"`
}

// handleStatus maneja el endpoint GET /status
func (rpc *RPCServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Solo aceptar GET
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido. Usa GET", http.StatusMethodNotAllowed)
		return
	}

	lastBlock := rpc.blockchain.Blocks[len(rpc.blockchain.Blocks)-1]

	status := StatusResponse{
		Blocks:        len(rpc.blockchain.Blocks),
		LastBlockHash: lastBlock.Hash,
		PendingTxs:    len(rpc.blockchain.PendingTxs),
		Peers:         rpc.server.PeerCount(),
		Mining:        rpc.server.IsMining(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
