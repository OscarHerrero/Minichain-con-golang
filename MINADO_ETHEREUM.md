# ⛏️ MINADO CONTINUO ESTILO ETHEREUM

## 🎯 ¿QUÉ SE HA IMPLEMENTADO?

Tu blockchain ahora funciona **exactamente igual que Ethereum** en cuanto al minado y propagación de bloques:

### ✅ Funcionalidades Implementadas

1. **Minado Continuo Automático**
   - Los nodos minan bloques constantemente cuando hay transacciones pendientes
   - No necesitas ejecutar comandos manualmente

2. **Propagación de Bloques**
   - Cuando un nodo mina un bloque, lo propaga automáticamente a todos sus peers
   - Los peers reciben, validan y agregan el bloque a su cadena

3. **Cancelación de Minado**
   - Si un nodo está minando y recibe un bloque nuevo de otro peer, **cancela inmediatamente** su minado actual
   - Esto evita trabajo duplicado (igual que Ethereum)

4. **Validación de Bloques**
   - Los bloques recibidos se validan antes de agregarlos
   - Se verifica el hash anterior, la dificultad y la integridad

5. **Evitar Propagación Duplicada**
   - Cuando un nodo recibe un bloque, NO lo reenvía al peer que se lo envió
   - Solo propaga a los demás peers

---

## 🚀 CÓMO USAR

### **Opción 1: Modo Normal (Minado Automático)**

```bash
# Nodo 1 (Bootstrap)
./minichain-node --port 3000 --datadir ./node1

# Nodo 2
./minichain-node --port 3001 --datadir ./node2 --bootstrap localhost:3000

# Nodo 3
./minichain-node --port 3002 --datadir ./node3 --bootstrap localhost:3000
```

**El minado está habilitado por defecto** (--mine=true)

### **Opción 2: Con Auto-Transacciones para Testing**

Para ver el minado en acción sin tener que crear transacciones manualmente:

```bash
# Nodo 1 con auto-transacciones
./minichain-node --port 3000 --datadir ./node1 --autotx

# Nodo 2
./minichain-node --port 3001 --datadir ./node2 --bootstrap localhost:3000

# Nodo 3
./minichain-node --port 3002 --datadir ./node3 --bootstrap localhost:3000
```

Con `--autotx`, el nodo crea automáticamente una transacción cada 20 segundos.

### **Opción 3: Deshabilitar Minado**

Si NO quieres que un nodo mine (solo que actúe como relay):

```bash
./minichain-node --port 3000 --datadir ./node1 --mine=false
```

---

## 📊 QUÉ VER EN PANTALLA

Cuando los nodos estén funcionando, verás algo como esto:

### **Cuando se crea una transacción:**

```
🤖 Transacción automática creada (#1) - Total pendientes: 1
```

### **Cuando se inicia el minado:**

```
⛏️  Iniciando minado de bloque 1 con 1 transacciones...
```

### **Cuando se mina un bloque:**

```
✅ Bloque 1 minado exitosamente! Hash: 00a3f5b8c2d1e4f7...
📡 Propagando bloque 1 a 2 peers...
```

### **Cuando se recibe un bloque:**

```
📦 Nuevo bloque recibido de 127.0.0.1:3001: Bloque #1
✅ Bloque #1 válido - agregando a la cadena
📊 Blockchain actualizada - altura: 1
📡 Bloque #1 propagado a 1 peers adicionales
```

### **Cuando se cancela el minado:**

```
⚠️  Minado cancelado - nuevo bloque recibido
```

### **Estado periódico cada 30 segundos:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⏰ 2025-12-12 10:30:00

📊 Blockchain:
   • Bloques: 5                    ← ¡Creciendo!
   • Último hash: 00abc123...
   • Transacciones pendientes: 0

⛏️  Minado:
   • Estado: ✅ ACTIVO             ← Minando continuamente

🌐 Red P2P:
   • Peers conectados: 2
   • Lista de peers:
     1. 127.0.0.1:3001 (altura: 5)
     2. 127.0.0.1:3002 (altura: 5)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## 🧪 PROBAR EL SISTEMA

### **Test 1: Ver que todos los nodos tienen la misma blockchain**

1. Inicia 3 nodos con `--autotx` en el nodo 1
2. Espera 1-2 minutos
3. Verás que todos los nodos muestran el mismo número de bloques
4. Los hashes coinciden entre todos

### **Test 2: Ver la cancelación de minado**

1. Inicia 3 nodos con `--autotx` en todos
2. Observa los logs
3. Verás mensajes de "Minado cancelado - nuevo bloque recibido"
4. Esto significa que un nodo estaba minando, pero otro nodo terminó primero

### **Test 3: Probar con dificultad más alta**

```bash
# Dificultad 4 (más difícil, tarda más en minar)
./minichain-node --port 3000 --datadir ./node1 --difficulty 4 --autotx
./minichain-node --port 3001 --datadir ./node2 --difficulty 4 --bootstrap localhost:3000
./minichain-node --port 3002 --datadir ./node3 --difficulty 4 --bootstrap localhost:3000
```

Con dificultad 4, el minado tarda más y verás mejor la competencia entre nodos.

---

## 🎬 SCRIPTS ACTUALIZADOS

Los scripts de inicio funcionan exactamente igual:

### **Windows:**

```cmd
start.bat
```

O con PowerShell:

```powershell
.\start.ps1
```

### **Linux/Mac:**

```bash
./start.sh
```

Para testing con auto-transacciones, edita los scripts y agrega `--autotx`:

```bash
# En start.sh, línea del nodo 1:
./minichain-node --port 3000 --datadir ./node1 --autotx
```

---

## 🔧 PARÁMETROS DISPONIBLES

```bash
--port          Puerto P2P (default: 3000)
--host          IP donde escuchar (default: 0.0.0.0)
--datadir       Directorio de datos (default: ./chaindata)
--difficulty    Dificultad de minado (default: 2)
--bootstrap     Nodos bootstrap (ej: 192.168.1.10:3000,192.168.1.11:3000)
--mine          Habilitar minado (default: true)
--autotx        Crear transacciones automáticas para testing (default: false)
```

---

## 📋 ARQUITECTURA DEL MINADO

### **Flujo Completo:**

```
┌─────────────────────────────────────────────────────────────┐
│  1. Transacción creada → Se agrega al mempool               │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  2. Minador detecta transacciones pendientes                 │
│     → Crea bloque candidato                                  │
│     → Comienza PoW (buscar nonce válido)                     │
└───────────────────────┬─────────────────────────────────────┘
                        │
            ┌───────────┴───────────┐
            │                       │
            ▼                       ▼
┌─────────────────────┐   ┌─────────────────────┐
│ CASO A:             │   │ CASO B:             │
│ Este nodo mina      │   │ Otro nodo mina      │
│ el bloque primero   │   │ primero             │
└──────┬──────────────┘   └──────┬──────────────┘
       │                         │
       │                         ▼
       │              ┌─────────────────────┐
       │              │ Bloque recibido     │
       │              │ por red P2P         │
       │              └──────┬──────────────┘
       │                     │
       │                     ▼
       │              ┌─────────────────────┐
       │              │ Cancelar minado     │
       │              │ actual              │
       │              └──────┬──────────────┘
       │                     │
       ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│  3. Bloque válido agregado a la blockchain                   │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  4. Propagar bloque a todos los peers (excepto origen)       │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  5. Limpiar mempool y comenzar a minar siguiente bloque      │
└─────────────────────────────────────────────────────────────┘
```

---

## ⚡ DIFERENCIAS CON IMPLEMENTACIÓN ANTERIOR

### **ANTES:**

- ❌ No había minado automático
- ❌ Los bloques no se propagaban
- ❌ Había que ejecutar comandos manualmente para minar
- ❌ No había cancelación de minado

### **AHORA:**

- ✅ Minado continuo automático
- ✅ Bloques se propagan a la red instantáneamente
- ✅ Si otro nodo mina primero, se cancela el minado actual
- ✅ Funciona exactamente como Ethereum

---

## 🛠️ ARCHIVOS MODIFICADOS

### **Nuevas funciones en `p2p/server.go`:**

- `StartMining()` - Inicia minado continuo
- `StopMining()` - Detiene minado
- `miningLoop()` - Bucle principal de minado
- `mineBlockWithCancellation()` - Mina un bloque con cancelación
- `mineWithCancellation()` - PoW interruptible
- `BroadcastBlock()` - Propaga bloque a todos los peers
- `handleNewBlock()` - Procesa bloques recibidos
- `BroadcastBlockExcept()` - Propaga evitando duplicados

### **Modificaciones en `cmd/node/main.go`:**

- Agregada flag `--mine` (default: true)
- Agregada flag `--autotx` (default: false)
- Función `autoCreateTransactions()` para testing
- Mostrar estado de minado en output periódico

---

## 🚧 PENDIENTE (MEJORAS FUTURAS)

1. **Resolución de forks completa**
   - Por ahora se ignoran bloques que no son el siguiente
   - TODO: Implementar "cadena más larga gana"

2. **Sincronización de cadena**
   - Si un peer tiene cadena más larga, descargar bloques faltantes

3. **Propagación de transacciones**
   - Las transacciones deberían propagarse por la red
   - Por ahora solo se propagan los bloques minados

4. **Ejecución de transacciones en bloques recibidos**
   - Los bloques recibidos se agregan pero sus transacciones no se ejecutan
   - TODO: Ejecutar y actualizar state

---

## ✅ RESUMEN

Tu blockchain ahora tiene:

1. ✅ Minado continuo automático
2. ✅ Propagación de bloques en tiempo real
3. ✅ Cancelación inteligente de minado
4. ✅ Validación de bloques recibidos
5. ✅ Evitar propagación duplicada

**¡Es una blockchain funcional estilo Ethereum!** 🎉

---

## 📚 MÁS INFORMACIÓN

- [README_P2P.md](./README_P2P.md) - Guía rápida de P2P
- [GUIA_RED_P2P.md](./GUIA_RED_P2P.md) - Arquitectura completa P2P
- [WINDOWS.md](./WINDOWS.md) - Guía específica para Windows
