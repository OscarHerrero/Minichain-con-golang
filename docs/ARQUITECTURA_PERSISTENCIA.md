# 🗄️ ARQUITECTURA DE PERSISTENCIA - Estilo Ethereum/Geth

## 📋 Resumen

Implementación de persistencia siguiendo exactamente la arquitectura de **Ethereum Go (Geth)**, usando:
- **LevelDB** como base de datos clave-valor
- **RLP encoding** para serialización
- **Merkle Patricia Trie** para el estado
- **Separación de ChainDB y StateDB**

---

## 🏗️ ARQUITECTURA DE GETH

### Componentes Principales

```
┌─────────────────────────────────────────────────────────┐
│                    MINICHAIN                             │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────┐         ┌──────────────────┐     │
│  │    ChainDB       │         │    StateDB       │     │
│  │  (LevelDB)       │         │  (LevelDB)       │     │
│  ├──────────────────┤         ├──────────────────┤     │
│  │ • Blocks         │         │ • Accounts       │     │
│  │ • Headers        │         │ • Contracts      │     │
│  │ • Transactions   │         │ • Storage        │     │
│  │ • Receipts       │         │ • Code           │     │
│  │ • Hashes → Number│         │                  │     │
│  └──────────────────┘         └──────────────────┘     │
│           │                            │                │
│           └────────────┬───────────────┘                │
│                        ↓                                │
│              ┌──────────────────┐                       │
│              │  Merkle Patricia │                       │
│              │      Trie        │                       │
│              └──────────────────┘                       │
└─────────────────────────────────────────────────────────┘
```

### Diferencia con Geth v1.9.0+

Geth separa datos en dos partes:
1. **Recent blocks** (últimos ~3 épocas): LevelDB en SSD
2. **Ancient data** (bloques antiguos): Freezer database (archivos flat)

Nosotros empezaremos solo con LevelDB para todo.

---

## 📊 ESTRUCTURA DE DATOS

### 1. ChainDB (Base de Datos de Cadena)

Almacena la estructura de la blockchain:

```
Key Scheme (prefijos):
┌──────────────────────────────────────────────────────┐
│ Prefix │ Key                    │ Value              │
├────────┼────────────────────────┼────────────────────┤
│ 'h'    │ num (uint64) + hash    │ header (RLP)       │
│ 'b'    │ num (uint64) + hash    │ body (RLP)         │
│ 't'    │ tx hash                │ transaction (RLP)  │
│ 'r'    │ num (uint64) + hash    │ receipts (RLP)     │
│ 'H'    │ num (uint64)           │ hash               │
│ 'l'    │ hash                   │ block number       │
└────────┴────────────────────────┴────────────────────┘
```

**Ejemplo de keys:**
```go
// Header del bloque #5
key: "h" + [5 en 8 bytes] + [hash del bloque]
value: RLP(header)

// Body del bloque #5
key: "b" + [5 en 8 bytes] + [hash del bloque]
value: RLP([transactions])

// Hash canónico del bloque #5
key: "H" + [5 en 8 bytes]
value: hash del bloque
```

### 2. StateDB (Base de Datos de Estado)

Almacena el estado de todas las cuentas usando Merkle Patricia Trie:

```
State Trie Structure:
┌──────────────────────────────────────────────┐
│         STATE ROOT (en header)               │
│                   │                          │
│         ┌─────────┴─────────┐               │
│         ↓                   ↓               │
│   [Address 1]         [Address 2]           │
│       │                     │               │
│   Account Object      Account Object        │
│   ├─ Nonce            ├─ Nonce              │
│   ├─ Balance          ├─ Balance            │
│   ├─ StorageRoot ──┐  ├─ StorageRoot        │
│   └─ CodeHash      │  └─ CodeHash           │
│                    │                         │
│                    └─→ Storage Trie         │
│                        ├─ Slot 1: Value     │
│                        ├─ Slot 2: Value     │
│                        └─ Slot N: Value     │
└──────────────────────────────────────────────┘
```

**Keys en StateDB:**
```go
// Nodo del Trie (por hash)
key: hash del nodo (32 bytes)
value: RLP(nodo del trie)

// Code de contrato
key: "c" + codeHash
value: bytecode del contrato

// Preimage (dirección → hash)
key: "secure-key-" + hash
value: dirección original
```

### 3. Merkle Patricia Trie

#### Tipos de Nodos:

```go
1. EmptyNode: null

2. LeafNode: [path, value]
   - path: nibbles restantes
   - value: datos RLP encoded

3. ExtensionNode: [path, key]
   - path: nibbles compartidos
   - key: hash del siguiente nodo

4. BranchNode: [v0, v1, ..., v15, value]
   - v0-v15: hashes de 16 hijos (hex)
   - value: valor si termina aquí
```

**Ejemplo visual:**

```
Insert: "dog" → "perro", "doge" → "moneda"

                    Root
                     │
            ExtensionNode ("do")
                     │
              BranchNode
              ├─ 'g' → LeafNode("" → "perro")
              └─ 'g' → ExtensionNode("e")
                           │
                      LeafNode("" → "moneda")
```

---

## 🔐 RLP ENCODING

**Recursive Length Prefix** - Serialización usada en Ethereum

### Reglas:

```
1. String (0-55 bytes):
   [0x80 + len] + data

2. String (56+ bytes):
   [0xb7 + len(len)] + len + data

3. List (0-55 bytes total):
   [0xc0 + len] + items

4. List (56+ bytes total):
   [0xf7 + len(len)] + len + items
```

### Ejemplos:

```go
// String "dog"
RLP: [0x83, 'd', 'o', 'g']
//    0x83 = 0x80 + 3

// Number 15
RLP: [0x0f]

// List ["cat", "dog"]
RLP: [0xc8, 0x83, 'c','a','t', 0x83, 'd','o','g']
//    0xc8 = 0xc0 + 8 (total bytes)

// Empty string ""
RLP: [0x80]

// Empty list []
RLP: [0xc0]
```

---

## 📁 ESTRUCTURA DE DIRECTORIOS

```
minichain/
├── database/
│   ├── database.go          # Interfaz base Database
│   ├── leveldb/
│   │   └── leveldb.go       # Implementación LevelDB
│   ├── memorydb/
│   │   └── memorydb.go      # DB en memoria (para tests)
│   └── batch.go             # Batch writes
├── ethdb/
│   └── database.go          # Alias para compatibilidad
├── rlp/
│   ├── encode.go            # RLP encoder
│   ├── decode.go            # RLP decoder
│   └── rlp_test.go
├── trie/
│   ├── trie.go              # Merkle Patricia Trie
│   ├── node.go              # Tipos de nodos
│   ├── encoding.go          # Compact/hex encoding
│   ├── hasher.go            # Hashing de nodos
│   ├── database.go          # Trie database
│   ├── secure_trie.go       # Secure trie (hash keys)
│   └── trie_test.go
├── core/
│   ├── state/
│   │   ├── statedb.go       # StateDB principal
│   │   ├── state_object.go  # Objeto de cuenta
│   │   └── database.go      # State database wrapper
│   └── rawdb/
│       ├── accessors_chain.go  # Lectura/escritura bloques
│       ├── accessors_state.go  # Lectura/escritura estado
│       ├── schema.go           # Key schemes
│       └── database.go         # Helpers
└── blockchain/
    ├── block.go             # Block (modificado con stateRoot)
    ├── blockchain.go        # Blockchain (con persistencia)
    └── account.go           # Account (serializable)
```

---

## 🔑 KEYS Y PREFIJOS (como Geth)

### ChainDB Keys:

```go
// Header
headerPrefix = 'h'  // header key prefix
headerKey = headerPrefix + num (uint64 big endian) + hash

// Body
bodyPrefix = 'b'
bodyKey = bodyPrefix + num (uint64 big endian) + hash

// Block number → hash (canonical)
headerHashPrefix = 'H'
headerHashKey = headerHashPrefix + num (uint64 big endian)

// Block hash → number
headerNumberPrefix = 'l'  // ('l' = lookup)
headerNumberKey = headerNumberPrefix + hash

// Transaction
txPrefix = 't'
txKey = txPrefix + txHash

// Receipt
receiptPrefix = 'r'
receiptKey = receiptPrefix + num (uint64 big endian) + hash

// Metadata
lastHeaderKey = "LastHeader"
lastBlockKey = "LastBlock"
```

### StateDB Keys:

```go
// Trie node
securePrefix = "secure-key-"
secureKey = Keccak256(address)

// Code
codePrefix = 'c'
codeKey = codePrefix + Keccak256(code)

// Preimage (para debugging)
preimagePrefix = "secure-key-"
preimageKey = preimagePrefix + hash
```

---

## 🔄 FLUJO DE OPERACIONES

### 1. Minar Bloque

```go
1. Recoger transacciones del mempool
2. Crear nuevo bloque
   ├─ PreviousHash = último bloque
   └─ StateRoot = ?  (calcularlo)

3. Ejecutar transacciones
   ├─ Actualizar StateDB
   │  ├─ Modificar balances
   │  ├─ Actualizar storage de contratos
   │  └─ Incrementar nonces
   └─ Generar receipts

4. Calcular Merkle Roots
   ├─ StateRoot = State Trie Root
   ├─ TxRoot = Transaction Trie Root
   └─ ReceiptRoot = Receipt Trie Root

5. Minar bloque (PoW)
   └─ Calcular hash incluyendo stateRoot

6. Persistir en ChainDB
   ├─ Guardar Header
   ├─ Guardar Body
   ├─ Guardar Transactions
   ├─ Guardar Receipts
   └─ Actualizar canonical hash

7. Commit StateDB
   └─ Guardar todos los cambios del trie
```

### 2. Sincronizar desde Otro Nodo

```go
1. Recibir bloque de peer
2. Validar bloque
   ├─ Verificar PoW
   ├─ Verificar PreviousHash
   └─ Verificar firmas de transacciones

3. Ejecutar transacciones localmente
4. Verificar StateRoot
   ├─ StateRoot calculado == StateRoot del bloque?
   └─ Si no coincide → RECHAZAR bloque

5. Persistir bloque si válido
6. Actualizar estado local
```

### 3. Consultar Estado de Cuenta

```go
1. Obtener StateRoot del último bloque
2. Abrir State Trie en ese root
3. Buscar address en el trie
   └─ Trie.Get(Keccak256(address))
4. Decodificar Account object (RLP)
5. Retornar balance, nonce, etc.
```

---

## 💾 EJEMPLO PRÁCTICO

### Guardar un bloque:

```go
// 1. Crear block header
header := &Header{
    ParentHash: parent.Hash(),
    Number:     big.NewInt(100),
    StateRoot:  stateRoot,    // ← Del State Trie
    TxRoot:     txRoot,       // ← Del Transaction Trie
    ReceiptRoot: receiptRoot, // ← Del Receipt Trie
    Difficulty: difficulty,
    Timestamp:  timestamp,
    Nonce:      nonce,
}

// 2. Calcular hash del bloque
hash := header.Hash()

// 3. Guardar header
key := headerKey(100, hash)
value := rlp.Encode(header)
chainDB.Put(key, value)

// 4. Guardar body (transacciones)
bodyKey := bodyKey(100, hash)
bodyValue := rlp.Encode(transactions)
chainDB.Put(bodyKey, bodyValue)

// 5. Guardar hash canónico
canonicalKey := headerHashKey(100)
chainDB.Put(canonicalKey, hash)
```

### Actualizar estado de cuenta:

```go
// 1. Abrir State Trie
stateTrie := trie.New(currentStateRoot, trieDB)

// 2. Obtener cuenta actual
key := crypto.Keccak256(address)
accountRLP := stateTrie.Get(key)
account := rlp.Decode(accountRLP)

// 3. Modificar cuenta
account.Balance += 100
account.Nonce++

// 4. Guardar cuenta modificada
newAccountRLP := rlp.Encode(account)
stateTrie.Update(key, newAccountRLP)

// 5. Calcular nuevo StateRoot
newStateRoot := stateTrie.Hash()

// 6. Commit changes
stateTrie.Commit()
```

---

## 🧪 DEPENDENCIAS GO

```go
// go.mod
module minichain

go 1.21

require (
    github.com/syndtr/goleveldb v1.0.0       // LevelDB
    golang.org/x/crypto v0.17.0               // Keccak256
)
```

---

## 🔍 DIFERENCIAS CON NUESTRA IMPLEMENTACIÓN ACTUAL

| Aspecto | Actual | Con Persistencia Geth |
|---------|--------|----------------------|
| **Almacenamiento** | RAM (se pierde) | Disco (permanente) |
| **Bloques** | Array en memoria | LevelDB con keys |
| **Estado** | Map simple | Merkle Patricia Trie |
| **StateRoot** | ❌ No existe | ✅ En cada bloque |
| **Serialización** | JSON informal | RLP standard |
| **Contratos** | Map en memoria | Trie + code storage |
| **Validación** | Solo PoW | PoW + StateRoot |
| **Sincronización** | ❌ Imposible | ✅ Verificable |

---

## 📈 VENTAJAS DE ESTA ARQUITECTURA

### 1. **Persistencia Real**
- Datos sobreviven al cerrar el programa
- Recuperación automática al iniciar

### 2. **Verificabilidad**
- StateRoot en cada bloque permite verificar estado
- Imposible alterar estado sin invalidar bloque

### 3. **Sincronización**
- Nodos pueden validar bloques recibidos
- StateRoot prueba que ejecutaron transacciones correctamente

### 4. **Eficiencia**
- Merkle proofs para verificar datos sin descargar todo
- Batch writes para operaciones múltiples

### 5. **Compatibilidad**
- Mismo formato que Ethereum
- Herramientas existentes pueden leer la DB

---

## 🚀 PLAN DE IMPLEMENTACIÓN

### Fase 1: Base (1-2 días)
1. ✅ Integrar LevelDB
2. ✅ Implementar interfaz Database
3. ✅ Implementar RLP encoding básico

### Fase 2: Trie (2-3 días)
4. ✅ Implementar nodos del Trie
5. ✅ Implementar Merkle Patricia Trie
6. ✅ Implementar hashing y encoding

### Fase 3: State (1-2 días)
7. ✅ Implementar StateDB
8. ✅ Implementar state objects
9. ✅ Integrar con Trie

### Fase 4: Chain (1-2 días)
10. ✅ Modificar Block con StateRoot
11. ✅ Implementar ChainDB accessors
12. ✅ Integrar con Blockchain

### Fase 5: Testing (1 día)
13. ✅ Tests unitarios
14. ✅ Tests de integración
15. ✅ Verificar persistencia funciona

---

## 📚 Referencias

- [Geth Database Documentation](https://geth.ethereum.org/docs/fundamentals/databases)
- [Ethereum LevelDB Structure](https://github.com/tpmccallum/ethereum_database_research_and_testing/blob/master/leveldb/leveldb.md)
- [Geth LevelDB Implementation](https://github.com/ethereum/go-ethereum/blob/master/ethdb/leveldb/leveldb.go)
- [Ethereum Trie Package](https://pkg.go.dev/github.com/ethereum/go-ethereum/trie)
- [Merkle Patricia Trie Explained](https://medium.com/coinmonks/data-structure-in-ethereum-episode-4-diving-by-examples-f6a4cbd8c329)

---

*Última actualización: 2025-12-11*
*Minichain - Implementación de persistencia estilo Ethereum/Geth* 🔐
