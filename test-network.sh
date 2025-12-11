#!/bin/bash

# Script para probar red P2P con 3 nodos locales

echo "╔════════════════════════════════════════════════════════════╗"
echo "║         🧪 TESTING RED P2P - MINICHAIN 🧪                 ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Limpiar datos anteriores
echo "🧹 Limpiando datos de pruebas anteriores..."
rm -rf ./test-node1 ./test-node2 ./test-node3

# Compilar el nodo
echo "🔨 Compilando nodo..."
go build -o minichain-node ./cmd/node
if [ $? -ne 0 ]; then
    echo "❌ Error compilando"
    exit 1
fi

echo "✅ Compilación exitosa"
echo ""

# Crear directorios
mkdir -p ./test-node1 ./test-node2 ./test-node3

echo "🚀 Iniciando 3 nodos..."
echo ""

# Nodo 1 (Bootstrap)
echo "📍 Iniciando Nodo 1 (Bootstrap) en puerto 3000..."
./minichain-node --port 3000 --datadir ./test-node1 &
NODE1_PID=$!
echo "   PID: $NODE1_PID"

# Esperar a que el nodo 1 inicie
sleep 3

# Nodo 2
echo "📍 Iniciando Nodo 2 en puerto 3001..."
./minichain-node --port 3001 --datadir ./test-node2 --bootstrap localhost:3000 &
NODE2_PID=$!
echo "   PID: $NODE2_PID"

# Esperar un poco
sleep 2

# Nodo 3
echo "📍 Iniciando Nodo 3 en puerto 3002..."
./minichain-node --port 3002 --datadir ./test-node3 --bootstrap localhost:3000 &
NODE3_PID=$!
echo "   PID: $NODE3_PID"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 3 NODOS INICIADOS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Estado de la red:"
echo "   • Nodo 1 (Bootstrap): localhost:3000 [PID: $NODE1_PID]"
echo "   • Nodo 2:             localhost:3001 [PID: $NODE2_PID]"
echo "   • Nodo 3:             localhost:3002 [PID: $NODE3_PID]"
echo ""
echo "🔍 Los nodos deberían conectarse automáticamente"
echo "   Verifica los logs arriba para ver mensajes de conexión"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "⏸️  Presiona Ctrl+C para detener todos los nodos"
echo ""

# Función para matar todos los nodos al salir
cleanup() {
    echo ""
    echo "🛑 Deteniendo nodos..."
    kill $NODE1_PID $NODE2_PID $NODE3_PID 2>/dev/null
    wait $NODE1_PID $NODE2_PID $NODE3_PID 2>/dev/null
    echo "✅ Todos los nodos detenidos"
    exit 0
}

# Capturar Ctrl+C
trap cleanup INT TERM

# Esperar indefinidamente
wait
