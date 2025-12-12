# 🚀 GUÍA RÁPIDA DE USO

## ⚡ INICIO RÁPIDO (30 segundos)

### 1. Iniciar un nodo

```bash
./minichain-node --port 3000 --datadir ./node1
```

El nodo automáticamente:
- ✅ Minará bloques cada segundo
- ✅ Escuchará RPC en http://localhost:8545
- ✅ Estará listo para recibir transacciones

### 2. Enviar una transacción

En otra terminal:

```bash
./minichain-sendtx --from Alice --to Bob --amount 10
```

¡Eso es todo! Verás en el nodo que:
1. La transacción se agrega al mempool
2. En el siguiente segundo se mina un bloque
3. El bloque incluye tu transacción

---

## 📋 COMANDOS PRINCIPALES

### **minichain-node** - Nodo completo

```bash
./minichain-node [opciones]
```

**Opciones:**

| Flag | Default | Descripción |
|------|---------|-------------|
| `--port` | 3000 | Puerto P2P |
| `--host` | 0.0.0.0 | IP donde escuchar |
| `--rpc` | 8545 | Puerto RPC/HTTP |
| `--datadir` | ./chaindata | Directorio de datos |
| `--difficulty` | 2 | Dificultad de minado |
| `--mine` | true | Habilitar minado |
| `--autotx` | false | Auto-transacciones (testing) |
| `--bootstrap` | - | Nodos bootstrap (ej: localhost:3000) |

**Ejemplos:**

```bash
# Nodo simple
./minichain-node

# Nodo en otro puerto
./minichain-node --port 3001 --datadir ./node2

# Conectar a otro nodo
./minichain-node --port 3001 --datadir ./node2 --bootstrap localhost:3000

# Nodo sin minado (solo relay)
./minichain-node --mine=false

# Cambiar puerto RPC
./minichain-node --rpc 9000
```

---

### **minichain-sendtx** - Enviar transacciones

```bash
./minichain-sendtx --from <dirección> --to <dirección> --amount <cantidad>
```

**Opciones:**

| Flag | Descripción |
|------|-------------|
| `--from` | Dirección origen (requerido) |
| `--to` | Dirección destino (requerido) |
| `--amount` | Cantidad a enviar (default: 0) |
| `--data` | Data en hex (opcional) |
| `--rpc` | URL del RPC (default: http://localhost:8545) |

**Ejemplos:**

```bash
# Transferencia simple
./minichain-sendtx --from Alice --to Bob --amount 10

# Transferencia a nodo específico
./minichain-sendtx --from Alice --to Bob --amount 5 --rpc http://localhost:8545

# Con data adicional
./minichain-sendtx --from Alice --to Bob --amount 1 --data "0x1234"
```

---

## 🎬 CASOS DE USO

### **Caso 1: Red local con 3 nodos**

```bash
# Terminal 1 - Nodo bootstrap
./minichain-node --port 3000 --datadir ./node1

# Terminal 2 - Nodo 2
./minichain-node --port 3001 --datadir ./node2 --bootstrap localhost:3000

# Terminal 3 - Nodo 3
./minichain-node --port 3002 --datadir ./node3 --bootstrap localhost:3000

# Terminal 4 - Enviar transacciones
./minichain-sendtx --from Alice --to Bob --amount 10
./minichain-sendtx --from Charlie --to Dave --amount 5
```

Los 3 nodos:
- Minarán bloques cada segundo
- Propagarán bloques entre sí
- Tendrán la misma blockchain

---

### **Caso 2: Testing con auto-transacciones**

```bash
# Nodo que genera transacciones automáticas cada 20 segundos
./minichain-node --port 3000 --datadir ./node1 --autotx

# Otro nodo que las recibe y mina
./minichain-node --port 3001 --datadir ./node2 --bootstrap localhost:3000
```

Verás bloques minándose automáticamente con las transacciones.

---

### **Caso 3: Múltiples PCs en red local**

**En PC 1 (192.168.1.10):**
```bash
./minichain-node --port 3000 --datadir ./node1
```

**En PC 2 (192.168.1.11):**
```bash
./minichain-node --port 3000 --datadir ./node1 --bootstrap 192.168.1.10:3000
```

**En PC 3 (192.168.1.12):**
```bash
./minichain-node --port 3000 --datadir ./node1 --bootstrap 192.168.1.10:3000
```

**Enviar transacción desde cualquier PC:**
```bash
./minichain-sendtx --from Alice --to Bob --amount 10 --rpc http://192.168.1.10:8545
```

---

## 🔍 VERIFICAR ESTADO

### **1. Via logs del nodo**

El nodo muestra estado cada 30 segundos:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⏰ 2025-12-12 10:30:00

📊 Blockchain:
   • Bloques: 42
   • Último hash: 00a3f5b8c2d1e4f7...
   • Transacciones pendientes: 0

⛏️  Minado:
   • Estado: ✅ ACTIVO

🌐 Red P2P:
   • Peers conectados: 2
   • Lista de peers:
     1. 127.0.0.1:3001 (altura: 42)
     2. 127.0.0.1:3002 (altura: 42)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### **2. Via API HTTP**

```bash
# Estado del nodo
curl http://localhost:8545/status

# Health check
curl http://localhost:8545/health
```

**Respuesta de /status:**
```json
{
  "blocks": 42,
  "lastBlockHash": "00a3f5b8c2d1e4f7...",
  "pendingTxs": 0,
  "peers": 2,
  "mining": true
}
```

---

## 🎯 FLUJO COMPLETO

```
1. Iniciar nodos
   ./minichain-node --port 3000
   ./minichain-node --port 3001 --bootstrap localhost:3000

2. Los nodos se conectan
   🔗 Conectando a localhost:3000...
   ✅ Peer conectado: localhost:3000

3. El minado empieza automáticamente
   ⛏️  Iniciando minado de bloque 1 (0 transacciones)...
   ✅ Bloque 1 minado exitosamente!

4. Enviar transacción
   ./minichain-sendtx --from Alice --to Bob --amount 10

5. La transacción se agrega al mempool
   📥 Transacción recibida por RPC: Alice → Bob (10.00 MTC)

6. En el siguiente segundo se mina con la transacción
   ⛏️  Iniciando minado de bloque 2 (1 transacciones)...
   ✅ Bloque 2 minado exitosamente! (txs: 1)

7. El bloque se propaga a todos los peers
   📡 Propagando bloque 2 a 1 peers...
   📦 Nuevo bloque recibido de 127.0.0.1:3000: Bloque #2

8. Todos los nodos tienen el mismo estado
   Nodo 1: 2 bloques
   Nodo 2: 2 bloques
```

---

## 🛠️ TROUBLESHOOTING

### **No se minan bloques**

Verifica que el minado esté habilitado:
```bash
./minichain-node --mine=true
```

### **No puedo enviar transacciones**

1. Verifica que el nodo esté corriendo
2. Verifica el puerto RPC: `curl http://localhost:8545/health`
3. Usa el puerto correcto: `--rpc http://localhost:8545`

### **Los nodos no se conectan**

1. Verifica el firewall (puerto 3000 debe estar abierto)
2. En Windows, ejecuta como administrador:
   ```powershell
   New-NetFirewallRule -DisplayName "Minichain" -Direction Inbound -Protocol TCP -LocalPort 3000 -Action Allow
   ```
3. Usa la IP correcta en `--bootstrap`

### **Error "address already in use"**

Otro proceso está usando el puerto. Opciones:
1. Cambia el puerto: `--port 3001`
2. Mata el proceso: `lsof -ti:3000 | xargs kill` (Linux/Mac)

---

## 📊 LOGS IMPORTANTES

### **Minado exitoso:**
```
✅ Bloque 5 minado exitosamente! Hash: 00a3f5b8... (txs: 2)
📡 Propagando bloque 5 a 2 peers...
```

### **Bloque recibido:**
```
📦 Nuevo bloque recibido de 127.0.0.1:3001: Bloque #5
✅ Bloque #5 válido - agregando a la cadena
📊 Blockchain actualizada - altura: 5
```

### **Minado cancelado (otro nodo ganó):**
```
⚠️  Minado cancelado - nuevo bloque recibido
```

### **Transacción recibida:**
```
📥 Transacción recibida por RPC: Alice → Bob (10.00 MTC)
```

---

## 🎓 CONCEPTOS IMPORTANTES

### **Minado cada segundo**

A diferencia de Bitcoin que ajusta la dificultad, esta blockchain mina un bloque cada segundo **con o sin transacciones**.

**Ventajas:**
- ⚡ Confirmaciones rápidas
- 🔄 Flujo constante de bloques
- 📦 Bloques vacíos válidos (como Ethereum)

### **Bloques vacíos**

Es normal ver bloques sin transacciones:
```
⛏️  Iniciando minado de bloque 10 (0 transacciones)...
✅ Bloque 10 minado exitosamente! (txs: 0)
```

Esto mantiene la cadena activa y permite timestamps regulares.

### **Propagación P2P**

Cuando un nodo mina un bloque:
1. Lo propaga a todos sus peers
2. Los peers lo validan y agregan
3. Los peers lo propagan a SUS peers
4. Toda la red se sincroniza

---

## 🔗 MÁS INFORMACIÓN

- [MINADO_ETHEREUM.md](./MINADO_ETHEREUM.md) - Detalles técnicos del minado
- [README_P2P.md](./README_P2P.md) - Arquitectura P2P
- [GUIA_RED_P2P.md](./GUIA_RED_P2P.md) - Guía completa de redes
- [WINDOWS.md](./WINDOWS.md) - Instrucciones para Windows

---

## ✅ CHECKLIST DE INICIO

- [ ] Compilar: `go build -o minichain-node ./cmd/node`
- [ ] Compilar sendtx: `go build -o minichain-sendtx ./cmd/sendtx`
- [ ] Iniciar nodo 1: `./minichain-node --port 3000`
- [ ] Iniciar nodo 2: `./minichain-node --port 3001 --bootstrap localhost:3000`
- [ ] Verificar conexión: Ver logs "Peer conectado"
- [ ] Verificar minado: Ver logs "Bloque minado"
- [ ] Enviar transacción: `./minichain-sendtx --from Alice --to Bob --amount 10`
- [ ] Verificar inclusión: Ver logs "Bloque minado (txs: 1)"

¡Listo! Tu blockchain está funcionando. 🎉
