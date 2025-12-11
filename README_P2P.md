# 🌐 MINICHAIN - RED P2P IMPLEMENTADA

## ✅ ¡Tu blockchain ahora es DISTRIBUIDA!

### 🚀 INICIO RÁPIDO (1 minuto)

```bash
# 1. Compilar
go build -o minichain-node ./cmd/node

# 2. Probar red con 3 nodos automáticamente
./test-network.sh

# O manualmente:

# Terminal 1 - Nodo Bootstrap
./minichain-node --port 3000 --datadir ./node1

# Terminal 2 - Nodo 2
./minichain-node --port 3001 --datadir ./node2 --bootstrap localhost:3000

# Terminal 3 - Nodo 3
./minichain-node --port 3002 --datadir ./node3 --bootstrap localhost:3000
```

### 📖 Documentación Completa
Ver **[GUIA_RED_P2P.md](./GUIA_RED_P2P.md)** para:
- Cómo conectar nodos en diferentes PCs
- Parámetros completos
- Troubleshooting
- Arquitectura del protocolo P2P

---

## 🎯 LO QUE YA FUNCIONA

### ✅ **Blockchain Core**
- Proof-of-Work completo
- Transacciones firmadas (ECDSA)
- EVM con 31 opcodes
- Contratos inteligentes
- Gas y snapshots

### ✅ **Persistencia Estilo Ethereum**
- LevelDB integrado
- Merkle Patricia Trie completo
- StateDB para cuentas/contratos
- ChainDB para bloques
- RLP encoding/decoding

### ✅ **Red P2P (NUEVO)**
- ✅ Servidor TCP en cada nodo
- ✅ Protocolo de mensajes binario
- ✅ Handshake entre peers
- ✅ Gestión de conexiones
- ✅ Keep-alive (ping/pong)
- ✅ Descubrimiento de peers
- ✅ Múltiples nodos comunicándose

---

## 🔄 PRÓXIMOS PASOS

### **En desarrollo:**
- [ ] Sincronización automática de bloques
- [ ] Propagación de bloques nuevos
- [ ] Propagación de transacciones
- [ ] Resolución de forks
- [ ] JSON-RPC API

---

## 📊 ESTADO DEL PROYECTO

```
COMPLETADO:
✅ Blockchain Core          100%
✅ EVM                       100%
✅ Persistencia              95%
✅ P2P Networking            70%
   ✅ Conexión entre nodos   100%
   ✅ Protocolo mensajes     100%
   ✅ Gestión de peers       100%
   ⚠️  Sincronización        30%
   ⚠️  Propagación           0%

PENDIENTE:
⏳ JSON-RPC API              0%
⏳ Sync completo             30%
```

---

## 🧪 TESTING

### **Verificar que P2P funciona:**

1. Ejecuta `./test-network.sh`
2. Deberías ver en los logs:
   ```
   ✅ Peer conectado: Peer{addr=...}
   Peers conectados: 2
   ```
3. Cada 30 segundos verás estadísticas:
   ```
   🌐 Red P2P:
      • Peers conectados: 2
      • Lista de peers:
        1. localhost:3001 (altura: 0)
        2. localhost:3002 (altura: 0)
   ```

### **Testing en red local (múltiples PCs):**

Ver **[GUIA_RED_P2P.md](./GUIA_RED_P2P.md)** sección "Probar en Múltiples PCs"

---

## 💡 ARQUITECTURA P2P

```
┌─────────────┐
│   Nodo 1    │
│  (Bootstrap)│◄─────┐
│   :3000     │      │
└──────┬──────┘      │
       │             │
       │ Handshake   │ Handshake
       │ Ping/Pong   │ Ping/Pong
       │             │
┌──────▼──────┐ ┌────┴──────┐
│   Nodo 2    │ │   Nodo 3  │
│   :3001     │ │   :3002   │
└─────────────┘ └───────────┘
```

### **Protocolo de Mensajes:**
- Formato binario eficiente
- Tipos: Handshake, Ping, Pong, NewBlock, etc.
- Keep-alive automático cada 30s
- Desconexión automática si no responde

---

## 📝 EJEMPLOS DE USO

### **Nodo simple:**
```bash
./minichain-node
```

### **Nodo en puerto específico:**
```bash
./minichain-node --port 8000
```

### **Nodo conectado a bootstrap:**
```bash
./minichain-node --bootstrap 192.168.1.10:3000
```

### **Nodo con múltiples bootstrap:**
```bash
./minichain-node --bootstrap 192.168.1.10:3000,192.168.1.11:3000
```

### **Nodo con alta dificultad:**
```bash
./minichain-node --difficulty 4
```

---

## 🎓 APRENDER MÁS

- **[ESTADO_Y_SIGUIENTE_PASO.md](./ESTADO_Y_SIGUIENTE_PASO.md)** - Estado completo del proyecto
- **[GUIA_RED_P2P.md](./GUIA_RED_P2P.md)** - Guía completa de P2P
- **[PLAN_DE_DESARROLLO.md](./PLAN_DE_DESARROLLO.md)** - Plan de desarrollo original

---

## 🏆 LOGROS

- ✅ Blockchain funcional con PoW
- ✅ EVM compatible con contratos
- ✅ Persistencia estilo Ethereum
- ✅ **Red P2P distribuida (NUEVO)**
- ✅ Múltiples nodos comunicándose
- ✅ Protocolo de red robusto

**¡Esto es una blockchain REAL!** 🎉

---

## 🤝 CONTRIBUIR

El proyecto está en desarrollo activo. Próximas características:
- Sincronización automática de blockchain
- Propagación de bloques y transacciones
- API JSON-RPC
- Cliente web3 compatible

---

## 📞 SOPORTE

Si tienes problemas:
1. Verifica que el puerto no esté en uso
2. Revisa que el firewall permita conexiones
3. Lee **[GUIA_RED_P2P.md](./GUIA_RED_P2P.md)** sección Troubleshooting
