@echo off
chcp 65001 >nul

echo ╔════════════════════════════════════════════════════════════╗
echo ║            🚀 MINICHAIN P2P - NODO 2 🚀                   ║
echo ╚════════════════════════════════════════════════════════════╝
echo.
echo Iniciando Nodo 2:
echo   • Puerto P2P:  3001
echo   • Puerto RPC:  8546
echo   • Dashboard:   http://localhost:8546
echo.
echo Conectando a peers: localhost:3000, localhost:3002
echo.

go run cmd/node/main.go --port 3001 --rpc 8546 --datadir ./node2 --bootstrap localhost:3000,localhost:3002
