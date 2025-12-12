@echo off
chcp 65001 >nul

echo ╔════════════════════════════════════════════════════════════╗
echo ║            🚀 MINICHAIN P2P - NODO 3 🚀                   ║
echo ╚════════════════════════════════════════════════════════════╝
echo.
echo Iniciando Nodo 3:
echo   • Puerto P2P:  3002
echo   • Puerto RPC:  8547
echo   • Dashboard:   http://localhost:8547
echo.
echo Conectando a peers: localhost:3000, localhost:3001
echo.

go run cmd/node/main.go --port 3002 --rpc 8547 --datadir ./node3 --bootstrap localhost:3000,localhost:3001
