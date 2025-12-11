package main

import (
	"flag"
	"fmt"
	"log"
	"minichain/blockchain"
	"minichain/p2p"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Parsear argumentos de línea de comandos
	port := flag.Int("port", 3000, "Puerto donde escuchar conexiones P2P")
	host := flag.String("host", "0.0.0.0", "IP donde escuchar (0.0.0.0 = todas)")
	datadir := flag.String("datadir", "./chaindata", "Directorio para datos de blockchain")
	difficulty := flag.Int("difficulty", 2, "Dificultad de minado")
	bootstrap := flag.String("bootstrap", "", "Nodos bootstrap separados por comas (ej: 192.168.1.10:3000,192.168.1.11:3000)")

	flag.Parse()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              🚀 MINICHAIN - NODO COMPLETO 🚀              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Crear o cargar blockchain con persistencia
	fmt.Printf("📂 Iniciando blockchain desde: %s\n", *datadir)
	bc, err := blockchain.NewBlockchainWithDB(*difficulty, *datadir)
	if err != nil {
		log.Fatalf("❌ Error iniciando blockchain: %v", err)
	}
	defer bc.Close()

	fmt.Printf("✅ Blockchain cargada con %d bloques\n", len(bc.Blocks))
	fmt.Println()

	// Crear servidor P2P
	server := p2p.NewServer(*host, *port, bc)

	// Iniciar servidor
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Error iniciando servidor P2P: %v", err)
	}

	fmt.Println()
	fmt.Println("┌────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ 🌐 Nodo escuchando en: %s:%d                       \n", *host, *port)
	fmt.Printf("│ 📊 Dificultad: %d                                          \n", *difficulty)
	fmt.Printf("│ 💾 Datos en: %s                                    \n", *datadir)
	fmt.Println("└────────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Conectar a nodos bootstrap si se especificaron
	if *bootstrap != "" {
		fmt.Println("🔗 Conectando a nodos bootstrap...")
		nodes := strings.Split(*bootstrap, ",")
		for _, node := range nodes {
			node = strings.TrimSpace(node)
			if node != "" {
				fmt.Printf("   → Conectando a %s...\n", node)

				// Intentar conectar (con timeout)
				go func(addr string) {
					time.Sleep(2 * time.Second) // Esperar un poco antes de conectar
					if err := server.ConnectToPeer(addr); err != nil {
						log.Printf("⚠️  Error conectando a %s: %v", addr, err)
					}
				}(node)
			}
		}
		fmt.Println()
	}

	// Iniciar goroutine para solicitar info de peers periódicamente
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if server.PeerCount() > 0 {
				server.BroadcastBlockchainInfo()
			}
		}
	}()

	// Mostrar estado periódicamente
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			printStatus(server, bc)
		}
	}()

	// Esperar señal de interrupción
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("✅ Nodo iniciado correctamente")
	fmt.Println("   Presiona Ctrl+C para detener")
	fmt.Println()

	// Mostrar estado inicial
	printStatus(server, bc)

	// Esperar señal
	<-sigChan

	fmt.Println("\n\n🛑 Señal de interrupción recibida, cerrando nodo...")

	// Cerrar servidor P2P
	server.Stop()

	// Cerrar blockchain
	bc.Close()

	fmt.Println("👋 Nodo detenido correctamente")
}

// printStatus imprime el estado actual del nodo
func printStatus(server *p2p.Server, bc *blockchain.Blockchain) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("⏰ %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Printf("📊 Blockchain:\n")
	fmt.Printf("   • Bloques: %d\n", len(bc.Blocks))
	fmt.Printf("   • Último hash: %s...\n", bc.Blocks[len(bc.Blocks)-1].Hash[:16])
	fmt.Printf("   • Transacciones pendientes: %d\n", len(bc.PendingTxs))
	fmt.Println()
	fmt.Printf("🌐 Red P2P:\n")
	fmt.Printf("   • Peers conectados: %d\n", server.PeerCount())

	peers := server.GetPeers()
	if len(peers) > 0 {
		fmt.Println("   • Lista de peers:")
		for i, peer := range peers {
			fmt.Printf("     %d. %s (altura: %d)\n", i+1, peer.GetAddress(), peer.GetBestHeight())
		}
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}
