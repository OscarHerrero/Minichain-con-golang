# 🪟 GUÍA PARA WINDOWS - MINICHAIN P2P

## 🚀 INICIO RÁPIDO

### **Método 1: Script BAT (Más fácil)**

Doble clic en `start.bat` y selecciona opción **1**.

O desde CMD:
```cmd
start.bat
```

### **Método 2: Script PowerShell**

Clic derecho en `start.ps1` → "Ejecutar con PowerShell"

O desde PowerShell:
```powershell
.\start.ps1
```

---

## ⚠️ PERMISOS DE EJECUCIÓN (PowerShell)

Si PowerShell no permite ejecutar scripts:

1. Abre PowerShell **como Administrador**
2. Ejecuta:
```powershell
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```
3. Confirma con "S" o "Y"

---

## 🖥️ INICIO MANUAL (3 Ventanas CMD)

### **Ventana 1 - Nodo Bootstrap:**
```cmd
minichain-node.exe --port 3000 --datadir ./node1
```

### **Ventana 2 - Nodo 2:**
```cmd
minichain-node.exe --port 3001 --datadir ./node2 --bootstrap localhost:3000
```

### **Ventana 3 - Nodo 3:**
```cmd
minichain-node.exe --port 3002 --datadir ./node3 --bootstrap localhost:3000
```

---

## 🔨 COMPILAR EN WINDOWS

Si no tienes el ejecutable `minichain-node.exe`:

```cmd
go build -o minichain-node.exe ./cmd/node
```

---

## 🔧 ABRIR FIREWALL

### **Opción 1: PowerShell (Recomendado)**

PowerShell **como Administrador**:
```powershell
New-NetFirewallRule -DisplayName "Minichain P2P" -Direction Inbound -LocalPort 3000 -Protocol TCP -Action Allow
```

### **Opción 2: Interfaz Gráfica**

1. Abre **Windows Defender Firewall**
2. Clic en "Configuración avanzada"
3. Clic en "Reglas de entrada" → "Nueva regla"
4. Tipo: **Puerto**
5. Protocolo: **TCP**, Puerto: **3000**
6. Acción: **Permitir la conexión**
7. Nombre: **Minichain P2P**

---

## 💻 PROBAR EN MÚLTIPLES PCs WINDOWS

### **Paso 1: Averiguar tu IP**

En CMD:
```cmd
ipconfig
```

Busca "Adaptador de red" → "Dirección IPv4" (ej: 192.168.1.100)

### **Paso 2: PC 1 (Bootstrap) - 192.168.1.100**
```cmd
minichain-node.exe --port 3000 --datadir ./chaindata
```

### **Paso 3: PC 2 - 192.168.1.101**
```cmd
minichain-node.exe --port 3001 --datadir ./chaindata --bootstrap 192.168.1.100:3000
```

### **Paso 4: PC 3 - 192.168.1.102**
```cmd
minichain-node.exe --port 3002 --datadir ./chaindata --bootstrap 192.168.1.100:3000
```

---

## 🛑 DETENER NODOS

### **Método 1: Cerrar Ventanas**
Simplemente cierra cada ventana CMD/PowerShell (o presiona Ctrl+C)

### **Método 2: CMD**
```cmd
taskkill /F /IM minichain-node.exe
```

### **Método 3: PowerShell**
```powershell
Get-Process minichain-node | Stop-Process
```

---

## ✅ VERIFICAR QUE FUNCIONA

Deberías ver en cada ventana cada 30 segundos:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⏰ 2025-12-11 15:30:00

📊 Blockchain:
   • Bloques: 1
   • Último hash: 00abc123...
   • Transacciones pendientes: 0

🌐 Red P2P:
   • Peers conectados: 2        ← ¡FUNCIONA!
   • Lista de peers:
     1. localhost:3001 (altura: 0)
     2. localhost:3002 (altura: 0)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## 🔍 TROUBLESHOOTING WINDOWS

### **"No se reconoce como comando"**
- Asegúrate de estar en el directorio correcto:
```cmd
cd C:\ruta\a\Minichain-con-golang
```

### **"Puerto ya en uso"**
Ver qué proceso está usando el puerto:
```cmd
netstat -ano | findstr :3000
taskkill /F /PID <número_del_proceso>
```

### **"Error de conexión" entre PCs**
1. Verifica que ambos PCs están en la misma red
2. Desactiva temporalmente el firewall para probar
3. Usa `ping 192.168.1.100` desde PC 2 para verificar conectividad

### **Script PowerShell no ejecuta**
```powershell
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

---

## 📊 COMPARACIÓN DE MÉTODOS

| Método | Ventajas | Desventajas |
|--------|----------|-------------|
| `start.bat` | ✅ Doble clic<br>✅ No requiere permisos | ⚠️ Ventanas CMD básicas |
| `start.ps1` | ✅ Colores<br>✅ Más profesional | ⚠️ Requiere permisos |
| Manual CMD | ✅ Control total | ⚠️ Más trabajo |

**Recomendación:** `start.bat` para principiantes, `start.ps1` para avanzados.

---

## 🎯 RESUMEN RÁPIDO

```
1. Doble clic en start.bat
2. Selecciona opción 1
3. ¡3 nodos corriendo!
4. Verifica "Peers conectados: 2"
```

---

## 📚 MÁS INFORMACIÓN

- [README_P2P.md](./README_P2P.md) - Guía rápida multiplataforma
- [GUIA_RED_P2P.md](./GUIA_RED_P2P.md) - Guía completa con arquitectura

---

## 💡 TIPS PARA WINDOWS

### **Ejecutar como Servicio (Avanzado)**

1. Instala NSSM (Non-Sucking Service Manager):
```cmd
choco install nssm
```

2. Crea servicio:
```cmd
nssm install Minichain "C:\ruta\minichain-node.exe" "--port 3000 --datadir C:\chaindata"
```

3. Inicia servicio:
```cmd
nssm start Minichain
```

### **Ver Logs**

Los scripts automáticos NO crean logs en archivos (se ven en pantalla).

Para guardar logs en archivo:
```cmd
minichain-node.exe --port 3000 --datadir ./node1 > node1.log 2>&1
```

---

## ✅ CHECKLIST WINDOWS

- [ ] Go instalado (`go version`)
- [ ] Compilado `minichain-node.exe`
- [ ] Puerto 3000 abierto en firewall
- [ ] Scripts `.bat` y `.ps1` descargados
- [ ] Probado con 3 nodos locales
- [ ] Verificado "Peers conectados"

---

¡Tu blockchain ahora funciona en Windows! 🎉
