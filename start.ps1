# PowerShell Script para iniciar Minichain P2P
# Codificación: UTF-8

Clear-Host

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║       🚀 INICIO RÁPIDO - MINICHAIN P2P 🚀                 ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""
Write-Host "Este script iniciará 3 nodos conectados en red P2P"
Write-Host ""
Write-Host "Opciones:"
Write-Host "  1) Iniciar 3 nodos locales (testing)"
Write-Host "  2) Iniciar 1 nodo (para conectar manualmente)"
Write-Host "  3) Ver guía de uso"
Write-Host "  4) Salir"
Write-Host ""
$option = Read-Host "Selecciona una opción (1-4)"

switch ($option) {
    "1" {
        Write-Host ""
        Write-Host "🚀 Iniciando 3 nodos locales..." -ForegroundColor Green
        Write-Host ""

        # Limpiar datos anteriores
        if (Test-Path "node1") { Remove-Item -Recurse -Force node1 }
        if (Test-Path "node2") { Remove-Item -Recurse -Force node2 }
        if (Test-Path "node3") { Remove-Item -Recurse -Force node3 }

        # Iniciar nodo 1 (Bootstrap)
        Write-Host "📍 Nodo 1 (Bootstrap) - P2P:3000 RPC:8545..." -ForegroundColor Yellow
        Start-Process powershell -ArgumentList "-NoExit", "-Command", ".\minichain-node.exe --port 3000 --rpc 8545 --datadir ./node1" -WindowStyle Normal
        Start-Sleep -Seconds 3

        # Iniciar nodo 2
        Write-Host "📍 Nodo 2 - P2P:3001 RPC:8546..." -ForegroundColor Yellow
        Start-Process powershell -ArgumentList "-NoExit", "-Command", ".\minichain-node.exe --port 3001 --rpc 8546 --datadir ./node2 --bootstrap localhost:3000" -WindowStyle Normal
        Start-Sleep -Seconds 2

        # Iniciar nodo 3
        Write-Host "📍 Nodo 3 - P2P:3002 RPC:8547..." -ForegroundColor Yellow
        Start-Process powershell -ArgumentList "-NoExit", "-Command", ".\minichain-node.exe --port 3002 --rpc 8547 --datadir ./node3 --bootstrap localhost:3000" -WindowStyle Normal
        Start-Sleep -Seconds 1

        Write-Host ""
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Green
        Write-Host "✅ ¡3 NODOS EN MARCHA EN VENTANAS SEPARADAS!" -ForegroundColor Green
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Green
        Write-Host ""
        Write-Host "Los nodos están corriendo en ventanas PowerShell separadas:"
        Write-Host "  • Nodo 1: P2P=3000  RPC=http://localhost:8545"
        Write-Host "  • Nodo 2: P2P=3001  RPC=http://localhost:8546"
        Write-Host "  • Nodo 3: P2P=3002  RPC=http://localhost:8547"
        Write-Host ""
        Write-Host "Para enviar transacciones:"
        Write-Host "  .\minichain-sendtx.exe --from Alice --to Bob --amount 10"
        Write-Host "  .\minichain-sendtx.exe --from Alice --to Bob --amount 10 --rpc http://localhost:8546"
        Write-Host ""
        Write-Host "Para detener:"
        Write-Host "  - Cierra cada ventana (Ctrl+C en cada una)"
        Write-Host "  - O ejecuta: Get-Process minichain-node | Stop-Process"
        Write-Host ""
        Read-Host "Presiona Enter para continuar"
    }

    "2" {
        Write-Host ""
        $port = Read-Host "Puerto (default 3000)"
        if ([string]::IsNullOrWhiteSpace($port)) { $port = "3000" }

        $datadir = Read-Host "Directorio de datos (default ./chaindata)"
        if ([string]::IsNullOrWhiteSpace($datadir)) { $datadir = "./chaindata" }

        $bootstrap = Read-Host "Nodo bootstrap (dejar vacío si eres el primero)"

        Write-Host ""
        Write-Host "🚀 Iniciando nodo..." -ForegroundColor Green

        if ([string]::IsNullOrWhiteSpace($bootstrap)) {
            Write-Host "Modo: BOOTSTRAP (primer nodo)" -ForegroundColor Yellow
            & .\minichain-node.exe --port $port --datadir $datadir
        } else {
            Write-Host "Modo: PEER (conectando a $bootstrap)" -ForegroundColor Yellow
            & .\minichain-node.exe --port $port --datadir $datadir --bootstrap $bootstrap
        }
    }

    "3" {
        Clear-Host
        Write-Host ""
        Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
        Write-Host "║                    GUÍA DE USO                             ║" -ForegroundColor Cyan
        Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "📖 INICIO MANUAL (3 Terminales PowerShell)" -ForegroundColor Yellow
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        Write-Host ""
        Write-Host "Terminal 1 (Nodo Bootstrap):"
        Write-Host "  .\minichain-node.exe --port 3000 --datadir ./node1" -ForegroundColor White
        Write-Host ""
        Write-Host "Terminal 2 (Nodo 2):"
        Write-Host "  .\minichain-node.exe --port 3001 --datadir ./node2 --bootstrap localhost:3000" -ForegroundColor White
        Write-Host ""
        Write-Host "Terminal 3 (Nodo 3):"
        Write-Host "  .\minichain-node.exe --port 3002 --datadir ./node3 --bootstrap localhost:3000" -ForegroundColor White
        Write-Host ""
        Write-Host ""
        Write-Host "💻 MÚLTIPLES PCs" -ForegroundColor Yellow
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        Write-Host ""
        Write-Host "PC 1 (192.168.1.100) - Nodo Bootstrap:"
        Write-Host "  .\minichain-node.exe --port 3000 --datadir ./chaindata" -ForegroundColor White
        Write-Host ""
        Write-Host "PC 2 (192.168.1.101):"
        Write-Host "  .\minichain-node.exe --port 3001 --datadir ./chaindata --bootstrap 192.168.1.100:3000" -ForegroundColor White
        Write-Host ""
        Write-Host ""
        Write-Host "📋 PARÁMETROS" -ForegroundColor Yellow
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        Write-Host ""
        Write-Host "  --port        Puerto donde escuchar (default: 3000)"
        Write-Host "  --host        IP donde escuchar (default: 0.0.0.0)"
        Write-Host "  --datadir     Directorio de datos (default: ./chaindata)"
        Write-Host "  --difficulty  Dificultad de minado (default: 2)"
        Write-Host "  --bootstrap   Nodos bootstrap separados por comas"
        Write-Host ""
        Write-Host ""
        Write-Host "🔧 ABRIR FIREWALL (PowerShell como Administrador)" -ForegroundColor Yellow
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        Write-Host ""
        Write-Host "New-NetFirewallRule -DisplayName 'Minichain P2P' -Direction Inbound -LocalPort 3000 -Protocol TCP -Action Allow" -ForegroundColor White
        Write-Host ""
        Write-Host ""
        Write-Host "📚 MÁS INFO" -ForegroundColor Yellow
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        Write-Host ""
        Write-Host "  README_P2P.md      - Guía rápida"
        Write-Host "  GUIA_RED_P2P.md    - Guía completa"
        Write-Host ""
        Read-Host "Presiona Enter para continuar"
    }

    "4" {
        Write-Host "👋 ¡Hasta luego!" -ForegroundColor Cyan
        exit
    }

    default {
        Write-Host "❌ Opción inválida" -ForegroundColor Red
        Read-Host "Presiona Enter para continuar"
    }
}
