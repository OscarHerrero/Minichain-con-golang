@echo off
chcp 65001 >nul
cls

echo ╔════════════════════════════════════════════════════════════╗
echo ║       🚀 INICIO RÁPIDO - MINICHAIN P2P 🚀                 ║
echo ╚════════════════════════════════════════════════════════════╝
echo.
echo Este script iniciará 3 nodos conectados en red P2P
echo.
echo Opciones:
echo   1) Iniciar 3 nodos locales (testing)
echo   2) Iniciar 1 nodo (para conectar manualmente)
echo   3) Ver guía de uso
echo   4) Salir
echo.
set /p option="Selecciona una opción (1-4): "

if "%option%"=="1" goto option1
if "%option%"=="2" goto option2
if "%option%"=="3" goto option3
if "%option%"=="4" goto option4
echo ❌ Opción inválida
pause
exit /b 1

:option1
echo.
echo 🚀 Iniciando 3 nodos locales...
echo.

REM Limpiar datos anteriores
if exist node1 rmdir /s /q node1
if exist node2 rmdir /s /q node2
if exist node3 rmdir /s /q node3

REM Iniciar nodo 1 (Bootstrap)
echo 📍 Nodo 1 (Bootstrap) iniciando en puerto 3000...
start "Minichain Node 1" cmd /k "minichain-node.exe --port 3000 --datadir ./node1"
timeout /t 3 /nobreak >nul

REM Iniciar nodo 2
echo 📍 Nodo 2 iniciando en puerto 3001...
start "Minichain Node 2" cmd /k "minichain-node.exe --port 3001 --datadir ./node2 --bootstrap localhost:3000"
timeout /t 2 /nobreak >nul

REM Iniciar nodo 3
echo 📍 Nodo 3 iniciando en puerto 3002...
start "Minichain Node 3" cmd /k "minichain-node.exe --port 3002 --datadir ./node3 --bootstrap localhost:3000"
timeout /t 1 /nobreak >nul

echo.
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo ✅ ¡3 NODOS EN MARCHA EN VENTANAS SEPARADAS!
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo.
echo Los nodos están corriendo en ventanas separadas.
echo.
echo Para detener:
echo   - Cierra cada ventana (Ctrl+C en cada una)
echo   - O ejecuta: taskkill /F /IM minichain-node.exe
echo.
pause
exit /b 0

:option2
echo.
set /p port="Puerto (default 3000): "
if "%port%"=="" set port=3000

set /p datadir="Directorio de datos (default ./chaindata): "
if "%datadir%"=="" set datadir=./chaindata

set /p bootstrap="Nodo bootstrap (dejar vacío si eres el primero): "

echo.
echo 🚀 Iniciando nodo...

if "%bootstrap%"=="" (
    echo Modo: BOOTSTRAP (primer nodo)
    minichain-node.exe --port %port% --datadir %datadir%
) else (
    echo Modo: PEER (conectando a %bootstrap%)
    minichain-node.exe --port %port% --datadir %datadir% --bootstrap %bootstrap%
)
exit /b 0

:option3
cls
echo.
echo ╔════════════════════════════════════════════════════════════╗
echo ║                    GUÍA DE USO                             ║
echo ╚════════════════════════════════════════════════════════════╝
echo.
echo 📖 INICIO MANUAL (3 Terminales CMD)
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo.
echo Terminal 1 (Nodo Bootstrap):
echo   minichain-node.exe --port 3000 --datadir ./node1
echo.
echo Terminal 2 (Nodo 2):
echo   minichain-node.exe --port 3001 --datadir ./node2 --bootstrap localhost:3000
echo.
echo Terminal 3 (Nodo 3):
echo   minichain-node.exe --port 3002 --datadir ./node3 --bootstrap localhost:3000
echo.
echo.
echo 💻 MÚLTIPLES PCs
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo.
echo PC 1 (192.168.1.100) - Nodo Bootstrap:
echo   minichain-node.exe --port 3000 --datadir ./chaindata
echo.
echo PC 2 (192.168.1.101):
echo   minichain-node.exe --port 3001 --datadir ./chaindata --bootstrap 192.168.1.100:3000
echo.
echo PC 3 (192.168.1.102):
echo   minichain-node.exe --port 3002 --datadir ./chaindata --bootstrap 192.168.1.100:3000
echo.
echo.
echo 📋 PARÁMETROS
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo.
echo   --port        Puerto donde escuchar (default: 3000)
echo   --host        IP donde escuchar (default: 0.0.0.0)
echo   --datadir     Directorio de datos (default: ./chaindata)
echo   --difficulty  Dificultad de minado (default: 2)
echo   --bootstrap   Nodos bootstrap separados por comas
echo.
echo.
echo 🔍 VERIFICAR QUE FUNCIONA
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo.
echo Deberías ver en pantalla cada 30 segundos:
echo.
echo   🌐 Red P2P:
echo      • Peers conectados: 2
echo      • Lista de peers:
echo        1. localhost:3001 (altura: 0)
echo        2. localhost:3002 (altura: 0)
echo.
echo.
echo 🔧 ABRIR FIREWALL (Si tienes problemas)
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo.
echo PowerShell como Administrador:
echo   New-NetFirewallRule -DisplayName "Minichain P2P" -Direction Inbound -LocalPort 3000 -Protocol TCP -Action Allow
echo.
echo.
echo 📚 MÁS INFO
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo.
echo   README_P2P.md      - Guía rápida
echo   GUIA_RED_P2P.md    - Guía completa
echo.
pause
exit /b 0

:option4
echo 👋 ¡Hasta luego!
exit /b 0
