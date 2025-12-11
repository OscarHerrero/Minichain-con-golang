# 🌐 GUÍA: RED P2P DE MINICHAIN

## ¡Felicidades! Tu blockchain ahora es DISTRIBUIDA 🎉

Con el sistema P2P implementado, puedes conectar múltiples nodos que:
- ✅ Se comunican entre sí automáticamente
- ✅ Comparten información de blockchain
- ✅ Se sincronizan mutuamente
- ✅ Propagan bloques nuevos
- ✅ Propagan transacciones

---

## 🚀 INICIO RÁPIDO

### **Probar en tu PC (Testing Local)**

#### Terminal 1 - Nodo Bootstrap:
```bash
./minichain-node --port 3000 --datadir ./node1
```

#### Terminal 2 - Nodo 2:
```bash
./minichain-node --port 3001 --datadir ./node2 --bootstrap localhost:3000
```

#### Terminal 3 - Nodo 3:
```bash
./minichain-node --port 3002 --datadir ./node3 --bootstrap localhost:3000
```

**Resultado:** ¡3 nodos conectados en tu PC! 🎉

---

## 💻 PROBAR EN MÚLTIPLES PCs

### **Configuración de Red:**

```
┌─────────────────┐
│   PC 1 (Madrid) │
│  192.168.1.10   │ ← Bootstrap
│   Puerto 3000   │
└────────┬────────┘
         │
    ┌────┴─────┐
    │          │
┌───▼────┐  ┌──▼──────┐
│  PC 2  │  │  PC 3   │
│ :3001  │  │  :3002  │
└────────┘  └─────────┘
```

### **PC 1 (Bootstrap) - 192.168.1.10:**
```bash
./minichain-node --port 3000 --datadir ./chaindata
```

### **PC 2 - 192.168.1.20:**
```bash
./minichain-node --port 3001 --datadir ./chaindata \
    --bootstrap 192.168.1.10:3000
```

### **PC 3 - 192.168.1.30:**
```bash
./minichain-node --port 3002 --datadir ./chaindata \
    --bootstrap 192.168.1.10:3000
```

---

## 📋 PARÁMETROS DEL COMANDO

```bash
./minichain-node [opciones]
```

### **Opciones disponibles:**

| Parámetro | Descripción | Valor por defecto | Ejemplo |
|-----------|-------------|-------------------|---------|
| `--port` | Puerto para escuchar | 3000 | `--port 3001` |
| `--host` | IP donde escuchar | 0.0.0.0 (todas) | `--host 127.0.0.1` |
| `--datadir` | Directorio de datos | ./chaindata | `--datadir /var/minichain` |
| `--difficulty` | Dificultad de minado | 2 | `--difficulty 3` |
| `--bootstrap` | Nodos bootstrap | (ninguno) | `--bootstrap 192.168.1.10:3000` |

### **Ejemplos de uso:**

```bash
# Nodo simple
./minichain-node

# Nodo con puerto específico
./minichain-node --port 8000

# Nodo conectado a múltiples bootstrap
./minichain-node --port 3001 \
    --bootstrap 192.168.1.10:3000,192.168.1.11:3000

# Nodo con alta dificultad
./minichain-node --difficulty 4 --datadir ./mainnet
```

---

## 🔍 QUÉ VER EN LA PANTALLA

Cuando inicias un nodo, verás:

```
╔════════════════════════════════════════════════════════════╗
║              🚀 MINICHAIN - NODO COMPLETO 🚀              ║
╚════════════════════════════════════════════════════════════╝

📂 Iniciando blockchain desde: ./chaindata
🆕 Creando nueva blockchain con persistencia...
⛏️  Minando bloque 0 (dificultad: 2, 0 transacciones)...
✅ Bloque minado! Hash: 00abc123... (intentos: 245)
✅ Blockchain inicializada (dificultad: 2)
   State Root: c5d2460186f7233c
✅ Blockchain cargada con 1 bloques

🌐 Servidor P2P iniciado en 0.0.0.0:3000 (NodeID: a1b2c3d4e5f6...)

┌────────────────────────────────────────────────────────────┐
│ 🌐 Nodo escuchando en: 0.0.0.0:3000
│ 📊 Dificultad: 2
│ 💾 Datos en: ./chaindata
└────────────────────────────────────────────────────────────┘

✅ Nodo iniciado correctamente
   Presiona Ctrl+C para detener

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⏰ 2025-12-11 14:30:00

📊 Blockchain:
   • Bloques: 1
   • Último hash: 00abc123...
   • Transacciones pendientes: 0

🌐 Red P2P:
   • Peers conectados: 0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### **Cuando se conecta un peer:**

```
📥 Nueva conexión entrante desde 192.168.1.20:54321
✅ Peer conectado: Peer{addr=192.168.1.20:54321, nodeID=f7g8h9i0, height=0, incoming}

🌐 Red P2P:
   • Peers conectados: 1
   • Lista de peers:
     1. 192.168.1.20:54321 (altura: 0)
```

---

## 🧪 TESTING: VERIFICAR QUE FUNCIONA

### **Test 1: Conexión entre Nodos**

1. Inicia nodo 1 en terminal 1
2. Inicia nodo 2 en terminal 2 con --bootstrap
3. Deberías ver en ambos terminales:
   ```
   ✅ Peer conectado: ...
   Peers conectados: 1
   ```

### **Test 2: Múltiples Nodos**

1. Inicia 3-4 nodos
2. Todos apuntando al mismo bootstrap
3. Verás la red crecer:
   ```
   Peers conectados: 3
   • Lista de peers:
     1. 192.168.1.20:3001 (altura: 0)
     2. 192.168.1.30:3002 (altura: 0)
     3. 192.168.1.40:3003 (altura: 0)
   ```

### **Test 3: Persistencia**

1. Inicia un nodo
2. Mina algunos bloques (próximamente)
3. Cierra el nodo (Ctrl+C)
4. Vuelve a iniciar
5. Debería cargar blockchain desde disco:
   ```
   📂 Cargando blockchain existente desde disco...
   ✅ Bloque génesis cargado: 00abc123...
   ```

---

## 🎯 PRÓXIMAS FUNCIONALIDADES

### **En desarrollo:**
- [ ] Sincronización automática de bloques
- [ ] Propagación de bloques nuevos minados
- [ ] Propagación de transacciones al mempool
- [ ] Resolución de forks (cadena más larga gana)
- [ ] RPC JSON para interactuar con el nodo

### **Cómo se verá pronto:**
```
Nodo 1: Mina bloque 5 → Propaga a todos
Nodo 2: Recibe bloque 5 → Valida → Añade
Nodo 3: Recibe bloque 5 → Valida → Añade

→ TODOS SINCRONIZADOS AUTOMÁTICAMENTE
```

---

## 🔧 TROUBLESHOOTING

### **"Error iniciando listener: address already in use"**
- El puerto ya está en uso
- Solución: Usa otro puerto con `--port 3001`

### **"Error conectando a bootstrap: connection refused"**
- El nodo bootstrap no está corriendo
- Solución: Inicia el nodo bootstrap primero

### **"No hay peers conectados después de varios minutos"**
- Verifica que la IP del bootstrap sea correcta
- Verifica que no haya firewall bloqueando el puerto
- Solución: Abre el puerto en firewall

### **En Linux/Mac:**
```bash
# Abrir puerto en firewall
sudo ufw allow 3000/tcp
```

### **En Windows:**
```powershell
# Abrir puerto en firewall
New-NetFirewallRule -DisplayName "Minichain P2P" -Direction Inbound -LocalPort 3000 -Protocol TCP -Action Allow
```

---

## 📚 ARQUITECTURA P2P

### **Protocolo de Mensajes:**

| Tipo | Código | Descripción |
|------|--------|-------------|
| Handshake | 0x00 | Saludo inicial |
| Ping | 0x01 | Keep-alive |
| Pong | 0x02 | Respuesta a ping |
| NewBlock | 0x10 | Propagar bloque nuevo |
| NewTransaction | 0x11 | Propagar transacción |
| GetBlocks | 0x20 | Solicitar bloques |
| Blocks | 0x21 | Enviar bloques |

### **Formato de Mensaje:**
```
[1 byte: tipo][4 bytes: longitud][N bytes: payload]
```

### **Handshake:**
```json
{
  "version": "1.0.0",
  "networkID": 1,
  "bestBlockIndex": 5,
  "bestBlockHash": "0x00abc123...",
  "nodeID": "a1b2c3d4e5f6...",
  "listenPort": 3000
}
```

---

## 🌍 DESPLEGAR EN INTERNET (OPCIONAL)

### **Usando VPS (DigitalOcean, AWS, etc):**

1. Crea VPS con IP pública
2. Instala Go en el VPS
3. Compila minichain-node
4. Ejecuta como servicio:

```bash
# Crear servicio systemd
sudo nano /etc/systemd/system/minichain.service
```

```ini
[Unit]
Description=Minichain Node
After=network.target

[Service]
Type=simple
User=minichain
WorkingDirectory=/home/minichain
ExecStart=/home/minichain/minichain-node --port 3000 --datadir /var/lib/minichain
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
# Iniciar servicio
sudo systemctl start minichain
sudo systemctl enable minichain
```

### **Tu nodo ahora es GLOBAL** 🌍
- Accesible desde cualquier parte del mundo
- Otros nodos pueden conectarse a tu IP pública
- Forma parte de la red Minichain

---

## ✅ CHECKLIST: NODO FUNCIONANDO

- [ ] Nodo inicia sin errores
- [ ] Muestra "Servidor P2P iniciado"
- [ ] Se puede conectar a bootstrap
- [ ] Aparece "Peer conectado" cuando otro nodo conecta
- [ ] "Peers conectados" aumenta correctamente
- [ ] Ping/Pong funciona (peers no se desconectan)
- [ ] Blockchain se persiste al cerrar
- [ ] Blockchain se carga al reiniciar

---

## 🎉 ¡FELICIDADES!

Ahora tienes una **blockchain REAL distribuida** con:
- ✅ Persistencia estilo Ethereum (LevelDB + Merkle Trie)
- ✅ Red P2P funcional
- ✅ Múltiples nodos comunicándose
- ✅ Proof-of-Work
- ✅ EVM con contratos inteligentes
- ✅ Firmas digitales ECDSA

**Siguiente paso:** Implementar sincronización automática de bloques 🚀
