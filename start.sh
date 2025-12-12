#!/bin/bash

clear

cat << "EOF"
╔════════════════════════════════════════════════════════════╗
║       🚀 INICIO RÁPIDO - MINICHAIN P2P 🚀                 ║
╚════════════════════════════════════════════════════════════╝
EOF

echo ""
echo "Este script iniciará 3 nodos conectados en red P2P"
echo ""
echo "Opciones:"
echo "  1) Iniciar 3 nodos locales (testing)"
echo "  2) Iniciar 1 nodo (para conectar manualmente)"
echo "  3) Ver guía de uso"
echo "  4) Salir"
echo ""
read -p "Selecciona una opción (1-4): " option

case $option in
    1)
        echo ""
        echo "🚀 Iniciando 3 nodos locales..."
        echo ""

        # Limpiar datos anteriores
        rm -rf ./node1 ./node2 ./node3 2>/dev/null

        # Iniciar nodo 1 (Bootstrap)
        echo "📍 Nodo 1 (Bootstrap) - P2P:3000 RPC:8545..."
        ./minichain-node --port 3000 --rpc 8545 --datadir ./node1 > node1.log 2>&1 &
        sleep 2

        # Iniciar nodo 2
        echo "📍 Nodo 2 - P2P:3001 RPC:8546..."
        ./minichain-node --port 3001 --rpc 8546 --datadir ./node2 --bootstrap localhost:3000 > node2.log 2>&1 &
        sleep 1

        # Iniciar nodo 3
        echo "📍 Nodo 3 - P2P:3002 RPC:8547..."
        ./minichain-node --port 3002 --rpc 8547 --datadir ./node3 --bootstrap localhost:3000 > node3.log 2>&1 &
        sleep 1

        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "✅ ¡3 NODOS EN MARCHA!"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "Los nodos están corriendo con los siguientes puertos:"
        echo "  • Nodo 1: P2P=3000  RPC=http://localhost:8545"
        echo "  • Nodo 2: P2P=3001  RPC=http://localhost:8546"
        echo "  • Nodo 3: P2P=3002  RPC=http://localhost:8547"
        echo ""
        echo "Para enviar transacciones:"
        echo "  ./minichain-sendtx --from Alice --to Bob --amount 10"
        echo "  ./minichain-sendtx --from Alice --to Bob --amount 10 --rpc http://localhost:8546"
        echo ""
        echo "Ver logs en tiempo real:"
        echo "  Nodo 1: tail -f node1.log"
        echo "  Nodo 2: tail -f node2.log"
        echo "  Nodo 3: tail -f node3.log"
        echo ""
        echo "Detener todos los nodos:"
        echo "  killall minichain-node"
        echo ""
        echo "O individualmente:"
        echo "  kill \$(lsof -ti:3000)  # Detener nodo 1"
        echo "  kill \$(lsof -ti:3001)  # Detener nodo 2"
        echo "  kill \$(lsof -ti:3002)  # Detener nodo 3"
        echo ""
        ;;

    2)
        echo ""
        read -p "Puerto (default 3000): " port
        port=${port:-3000}

        read -p "Directorio de datos (default ./chaindata): " datadir
        datadir=${datadir:-./chaindata}

        read -p "Nodo bootstrap (dejar vacío si eres el primero): " bootstrap

        echo ""
        echo "🚀 Iniciando nodo..."

        if [ -z "$bootstrap" ]; then
            echo "Modo: BOOTSTRAP (primer nodo)"
            ./minichain-node --port $port --datadir $datadir
        else
            echo "Modo: PEER (conectando a $bootstrap)"
            ./minichain-node --port $port --datadir $datadir --bootstrap $bootstrap
        fi
        ;;

    3)
        cat << "GUIDE"

╔════════════════════════════════════════════════════════════╗
║                    GUÍA DE USO                             ║
╚════════════════════════════════════════════════════════════╝

📖 INICIO MANUAL (3 Terminales)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Terminal 1 (Nodo Bootstrap):
  ./minichain-node --port 3000 --datadir ./node1

Terminal 2 (Nodo 2):
  ./minichain-node --port 3001 --datadir ./node2 --bootstrap localhost:3000

Terminal 3 (Nodo 3):
  ./minichain-node --port 3002 --datadir ./node3 --bootstrap localhost:3000


💻 MÚLTIPLES PCs
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

PC 1 (192.168.1.100) - Nodo Bootstrap:
  ./minichain-node --port 3000 --datadir ./chaindata

PC 2 (192.168.1.101):
  ./minichain-node --port 3001 --datadir ./chaindata --bootstrap 192.168.1.100:3000

PC 3 (192.168.1.102):
  ./minichain-node --port 3002 --datadir ./chaindata --bootstrap 192.168.1.100:3000


📋 PARÁMETROS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  --port        Puerto donde escuchar (default: 3000)
  --host        IP donde escuchar (default: 0.0.0.0)
  --datadir     Directorio de datos (default: ./chaindata)
  --difficulty  Dificultad de minado (default: 2)
  --bootstrap   Nodos bootstrap separados por comas


🔍 VERIFICAR QUE FUNCIONA
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Deberías ver en pantalla cada 30 segundos:

  🌐 Red P2P:
     • Peers conectados: 2
     • Lista de peers:
       1. localhost:3001 (altura: 0)
       2. localhost:3002 (altura: 0)


📚 MÁS INFO
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  README_P2P.md      - Guía rápida
  GUIA_RED_P2P.md    - Guía completa

GUIDE
        ;;

    4)
        echo "👋 ¡Hasta luego!"
        exit 0
        ;;

    *)
        echo "❌ Opción inválida"
        exit 1
        ;;
esac
