package p2p

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"minichain/blockchain"
	"minichain/crypto"
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

	// Endpoint para obtener balance de una cuenta
	http.HandleFunc("/balance/", rpc.handleBalance)

	// Endpoints API para el dashboard
	http.HandleFunc("/api/blocks", rpc.handleAPIBlocks)
	http.HandleFunc("/api/block/", rpc.handleAPIBlock)
	http.HandleFunc("/api/accounts", rpc.handleAPIAccounts)

	// Endpoint del dashboard (HTML)
	http.HandleFunc("/", rpc.handleDashboard)

	// Endpoint de health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%d", rpc.port)
	log.Printf("🌐 Servidor RPC iniciado en http://localhost%s", addr)
	log.Println("   Endpoints disponibles:")
	log.Println("   - GET  /                (Dashboard web)")
	log.Println("   - POST /tx              (Enviar transacción)")
	log.Println("   - GET  /status          (Estado de la blockchain)")
	log.Println("   - GET  /balance/<addr>  (Obtener balance de una cuenta)")
	log.Println("   - GET  /api/blocks      (Lista de bloques)")
	log.Println("   - GET  /api/block/<n>   (Detalle de bloque)")
	log.Println("   - GET  /api/accounts    (Lista de cuentas)")
	log.Println("   - GET  /health          (Health check)")

	return http.ListenAndServe(addr, nil)
}

// TxRequest es la estructura de una transacción recibida por RPC
type TxRequest struct {
	From       string      `json:"from"`
	To         string      `json:"to"`
	Amount     float64     `json:"amount"`
	Nonce      int         `json:"nonce"`
	Data       string      `json:"data"`
	Signature  string      `json:"signature"`
	PublicKeyX interface{} `json:"publicKeyX"` // big.Int se serializa como string/number
	PublicKeyY interface{} `json:"publicKeyY"`
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

	// Parsear big.Int desde interface{}
	var pubKeyX, pubKeyY *big.Int

	if txReq.PublicKeyX != nil {
		pubKeyX = new(big.Int)
		switch v := txReq.PublicKeyX.(type) {
		case string:
			pubKeyX.SetString(v, 10)
		case float64:
			pubKeyX.SetInt64(int64(v))
		}
	}

	if txReq.PublicKeyY != nil {
		pubKeyY = new(big.Int)
		switch v := txReq.PublicKeyY.(type) {
		case string:
			pubKeyY.SetString(v, 10)
		case float64:
			pubKeyY.SetInt64(int64(v))
		}
	}

	// Crear transacción
	tx := &blockchain.Transaction{
		From:       txReq.From,
		To:         txReq.To,
		Amount:     txReq.Amount,
		Nonce:      txReq.Nonce,
		Data:       []byte{},
		Signature:  txReq.Signature,
		PublicKeyX: pubKeyX,
		PublicKeyY: pubKeyY,
	}

	// Parsear data si existe
	if txReq.Data != "" {
		tx.Data = []byte(txReq.Data)
	}

	// Verificar firma si está presente
	if tx.Signature != "" && tx.PublicKeyX != nil && tx.PublicKeyY != nil {
		// Reconstruir datos para verificar
		txData := fmt.Sprintf("%s%s%.2f%d%s", tx.From, tx.To, tx.Amount, tx.Nonce, string(tx.Data))

		// Verificar firma usando la función del paquete crypto
		if !crypto.VerifySignature(tx.PublicKeyX, tx.PublicKeyY, []byte(txData), tx.Signature) {
			http.Error(w, "❌ Firma inválida", http.StatusBadRequest)
			log.Printf("❌ Transacción rechazada - firma inválida: %s → %s", tx.From, tx.To)
			return
		}

		log.Printf("✅ Firma verificada correctamente")
	}

	// Agregar al mempool
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

// BalanceResponse es la respuesta del endpoint /balance
type BalanceResponse struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
	Nonce   int     `json:"nonce"`
}

// handleBalance maneja el endpoint GET /balance/<address>
func (rpc *RPCServer) handleBalance(w http.ResponseWriter, r *http.Request) {
	// Solo aceptar GET
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido. Usa GET", http.StatusMethodNotAllowed)
		return
	}

	// Extraer la dirección de la URL (después de /balance/)
	address := r.URL.Path[len("/balance/"):]
	if address == "" {
		http.Error(w, "Dirección requerida. Usa /balance/<address>", http.StatusBadRequest)
		return
	}

	// Obtener balance y nonce del AccountState
	balance := rpc.blockchain.GetBalance(address)
	nonce := rpc.blockchain.GetNonce(address)

	response := BalanceResponse{
		Address: address,
		Balance: balance,
		Nonce:   nonce,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("📊 Balance consultado: %s = %.2f MTC (nonce: %d)", address[:16]+"...", balance, nonce)
}

// handleAPIBlocks maneja el endpoint GET /api/blocks
func (rpc *RPCServer) handleAPIBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido. Usa GET", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rpc.blockchain.Blocks)
}

// handleAPIBlock maneja el endpoint GET /api/block/<index>
func (rpc *RPCServer) handleAPIBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido. Usa GET", http.StatusMethodNotAllowed)
		return
	}

	// Extraer índice de la URL
	indexStr := r.URL.Path[len("/api/block/"):]
	if indexStr == "" {
		http.Error(w, "Índice requerido. Usa /api/block/<index>", http.StatusBadRequest)
		return
	}

	var index int
	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		http.Error(w, "Índice inválido", http.StatusBadRequest)
		return
	}

	if index < 0 || index >= len(rpc.blockchain.Blocks) {
		http.Error(w, "Bloque no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rpc.blockchain.Blocks[index])
}

// AccountInfo representa información de una cuenta
type AccountInfo struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
	Nonce   int     `json:"nonce"`
}

// handleAPIAccounts maneja el endpoint GET /api/accounts
func (rpc *RPCServer) handleAPIAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido. Usa GET", http.StatusMethodNotAllowed)
		return
	}

	// Convertir AccountState a lista de AccountInfo
	accounts := []AccountInfo{}
	for addr, account := range rpc.blockchain.AccountState.Accounts {
		accounts = append(accounts, AccountInfo{
			Address: addr,
			Balance: account.Balance,
			Nonce:   account.Nonce,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accounts)
}

// handleDashboard sirve el HTML del dashboard
func (rpc *RPCServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dashboardHTML))
}

// dashboardHTML es el HTML del explorador de blockchain
const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Minichain Explorer</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
            color: #333;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        .header {
            text-align: center;
            color: white;
            margin-bottom: 30px;
        }

        .header h1 {
            font-size: 3em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }

        .header p {
            font-size: 1.2em;
            opacity: 0.9;
        }

        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .stat-card {
            background: white;
            padding: 25px;
            border-radius: 15px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            transition: transform 0.3s;
        }

        .stat-card:hover {
            transform: translateY(-5px);
        }

        .stat-label {
            font-size: 0.9em;
            color: #666;
            margin-bottom: 8px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }

        .stat-value {
            font-size: 2.5em;
            font-weight: bold;
            color: #667eea;
        }

        .section {
            background: white;
            border-radius: 15px;
            padding: 25px;
            margin-bottom: 20px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }

        .section-title {
            font-size: 1.8em;
            margin-bottom: 20px;
            color: #667eea;
            border-bottom: 3px solid #667eea;
            padding-bottom: 10px;
        }

        .block-list, .account-list {
            display: grid;
            gap: 15px;
        }

        .block-item, .account-item {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 10px;
            border-left: 4px solid #667eea;
            transition: all 0.3s;
        }

        .block-item:hover, .account-item:hover {
            background: #e9ecef;
            transform: translateX(5px);
        }

        .block-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }

        .block-index {
            font-size: 1.5em;
            font-weight: bold;
            color: #667eea;
        }

        .block-hash {
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
            color: #666;
            background: white;
            padding: 5px 10px;
            border-radius: 5px;
        }

        .block-info {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 10px;
            margin-top: 10px;
        }

        .info-item {
            font-size: 0.9em;
        }

        .info-label {
            color: #666;
            margin-right: 5px;
        }

        .info-value {
            font-weight: 600;
            color: #333;
        }

        .account-address {
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
            color: #667eea;
            margin-bottom: 10px;
            word-break: break-all;
        }

        .account-info {
            display: flex;
            gap: 30px;
        }

        .balance {
            font-size: 1.3em;
            font-weight: bold;
            color: #28a745;
        }

        .refresh-btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 25px;
            font-size: 1em;
            cursor: pointer;
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
            transition: all 0.3s;
            margin-bottom: 20px;
        }

        .refresh-btn:hover {
            background: #5568d3;
            transform: translateY(-2px);
            box-shadow: 0 8px 20px rgba(102, 126, 234, 0.6);
        }

        .loading {
            text-align: center;
            padding: 50px;
            font-size: 1.2em;
            color: #667eea;
        }

        .tx-count {
            background: #ffc107;
            color: white;
            padding: 5px 10px;
            border-radius: 15px;
            font-size: 0.9em;
            font-weight: bold;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .mining-indicator {
            display: inline-block;
            width: 12px;
            height: 12px;
            border-radius: 50%;
            margin-left: 10px;
        }

        .mining-indicator.active {
            background: #28a745;
            animation: pulse 2s infinite;
        }

        .mining-indicator.inactive {
            background: #dc3545;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⛓️ Minichain Explorer</h1>
            <p>Explorador de Blockchain en Tiempo Real</p>
        </div>

        <button class="refresh-btn" onclick="loadAll()">🔄 Actualizar Datos</button>

        <div class="stats-grid" id="stats">
            <div class="loading">Cargando estadísticas...</div>
        </div>

        <div class="section">
            <h2 class="section-title">📦 Bloques Recientes</h2>
            <div class="block-list" id="blocks">
                <div class="loading">Cargando bloques...</div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">💰 Cuentas</h2>
            <div class="account-list" id="accounts">
                <div class="loading">Cargando cuentas...</div>
            </div>
        </div>
    </div>

    <script>
        async function loadStatus() {
            const response = await fetch('/status');
            const data = await response.json();

            const miningIndicator = data.mining
                ? '<span class="mining-indicator active"></span>'
                : '<span class="mining-indicator inactive"></span>';

            document.getElementById('stats').innerHTML = '' +
                '<div class="stat-card">' +
                    '<div class="stat-label">Altura de Blockchain</div>' +
                    '<div class="stat-value">' + data.blocks + '</div>' +
                '</div>' +
                '<div class="stat-card">' +
                    '<div class="stat-label">Transacciones Pendientes</div>' +
                    '<div class="stat-value">' + data.pendingTxs + '</div>' +
                '</div>' +
                '<div class="stat-card">' +
                    '<div class="stat-label">Peers Conectados</div>' +
                    '<div class="stat-value">' + data.peers + '</div>' +
                '</div>' +
                '<div class="stat-card">' +
                    '<div class="stat-label">Estado de Minado</div>' +
                    '<div class="stat-value">' + (data.mining ? 'ON' : 'OFF') + miningIndicator + '</div>' +
                '</div>';
        }

        async function loadBlocks() {
            const response = await fetch('/api/blocks');
            const blocks = await response.json();

            const blocksHTML = blocks.slice().reverse().slice(0, 10).map(block => {
                const date = new Date(block.Timestamp);
                const hashShort = block.Hash.substring(0, 16) + '...';

                return '<div class="block-item">' +
                    '<div class="block-header">' +
                        '<div class="block-index">Bloque #' + block.Index + '</div>' +
                        '<div class="block-hash">' + hashShort + '</div>' +
                    '</div>' +
                    '<div class="block-info">' +
                        '<div class="info-item">' +
                            '<span class="info-label">Transacciones:</span>' +
                            '<span class="tx-count">' + block.Transactions.length + '</span>' +
                        '</div>' +
                        '<div class="info-item">' +
                            '<span class="info-label">Timestamp:</span>' +
                            '<span class="info-value">' + date.toLocaleString() + '</span>' +
                        '</div>' +
                        '<div class="info-item">' +
                            '<span class="info-label">Nonce:</span>' +
                            '<span class="info-value">' + block.Nonce + '</span>' +
                        '</div>' +
                    '</div>' +
                '</div>';
            }).join('');

            document.getElementById('blocks').innerHTML = blocksHTML || '<div class="loading">No hay bloques</div>';
        }

        async function loadAccounts() {
            const response = await fetch('/api/accounts');
            const accounts = await response.json();

            if (accounts.length === 0) {
                document.getElementById('accounts').innerHTML = '<div class="loading">No hay cuentas con balance</div>';
                return;
            }

            const accountsHTML = accounts.sort((a, b) => b.balance - a.balance).map(account => {
                const addressShort = account.address.substring(0, 20) + '...';

                return '<div class="account-item">' +
                    '<div class="account-address">' + account.address + '</div>' +
                    '<div class="account-info">' +
                        '<div>' +
                            '<span class="info-label">Balance:</span>' +
                            '<span class="balance">' + account.balance.toFixed(2) + ' MTC</span>' +
                        '</div>' +
                        '<div>' +
                            '<span class="info-label">Nonce:</span>' +
                            '<span class="info-value">' + account.nonce + '</span>' +
                        '</div>' +
                    '</div>' +
                '</div>';
            }).join('');

            document.getElementById('accounts').innerHTML = accountsHTML;
        }

        async function loadAll() {
            await Promise.all([
                loadStatus(),
                loadBlocks(),
                loadAccounts()
            ]);
        }

        // Cargar datos al inicio
        loadAll();

        // Auto-refresh cada 5 segundos
        setInterval(loadAll, 5000);
    </script>
</body>
</html>
`
