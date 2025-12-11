# CLAUDE.md - Minichain Development Guide

## Project Overview

**Minichain** is a blockchain implementation written in Go that includes:
- A proof-of-work blockchain with configurable difficulty
- Account-based transaction model with signatures
- Smart contract support via an Ethereum-compatible EVM
- Assembly compiler for writing smart contracts
- Wallet management with cryptographic key pairs
- Interactive CLI for blockchain operations

**Module Name**: `minichain`
**Go Version**: 1.25.5
**Total Lines of Code**: ~2,400 lines
**Language**: Go with Spanish comments and English code

---

## Repository Structure

```
Minichain-con-golang/
├── blockchain/          # Core blockchain logic
│   ├── account.go      # Account state management
│   ├── block.go        # Block structure and mining
│   ├── blockchain.go   # Blockchain implementation
│   └── transacction.go # Transaction handling (note: typo in filename)
├── compiler/           # Assembly to bytecode compiler
│   └── assembler.go    # EVM assembly compiler
├── crypto/             # Cryptographic primitives
│   ├── keypair.go      # ECDSA key pair generation
│   └── wallet.go       # Multi-account wallet
├── evm/                # Ethereum Virtual Machine
│   ├── contract.go     # Smart contract structure
│   ├── interpreter.go  # EVM interpreter
│   ├── memory.go       # EVM memory
│   ├── opcodes.go      # EVM opcode definitions
│   ├── stack.go        # EVM stack
│   └── storage.go      # Persistent contract storage
├── utils/              # Utility functions
│   └── crypto.go       # Hash functions
├── minichain.go        # Main CLI application
├── go.mod              # Go module definition
└── README.md           # Basic project description
```

---

## Package Architecture

### 1. `blockchain` Package

**Purpose**: Core blockchain functionality including blocks, transactions, and account state.

#### Key Components:

**Block** (`block.go`):
- Contains index, timestamp, transactions, previous hash, hash, and nonce
- Implements Proof-of-Work mining via `MineBlock(difficulty int)`
- Hash calculation includes ALL transaction data (from, to, amount, nonce, data, signature)
- Validation checks hash integrity and difficulty compliance

**Blockchain** (`blockchain.go`):
- Manages chain of blocks with configurable difficulty
- Maintains `AccountState` for all accounts
- Handles `PendingTxs` mempool for unconfirmed transactions
- Stores deployed smart contracts in `Contracts` map
- **Mining Process**:
  1. Creates block with pending transactions
  2. Mines block (Proof-of-Work)
  3. Executes all transactions in block
  4. Updates account state and contract storage
  5. Clears pending transactions

**Transaction** (`transacction.go`):
- Supports three types:
  1. **Simple Transfer**: From → To with amount
  2. **Contract Deployment**: To = "", Data = bytecode
  3. **Contract Call**: To = contract address, Data = calldata
- Includes nonce for replay protection
- Signed with ECDSA (secp256k1-like)
- **Execution Model** (Ethereum-style):
  1. Validates signature and nonce
  2. Creates snapshot of state
  3. Reserves maximum gas cost
  4. Increments nonce (not reverted even on failure)
  5. Executes transaction
  6. On success: refunds unused gas
  7. On failure: reverts state changes, keeps nonce increment, no gas refund

**Account** (`account.go`):
- Stores address, balance (float64), and nonce
- `AccountState` manages all accounts globally
- Supports snapshots for transaction revert capability
- Auto-creates accounts on first access with 0 balance

#### Important Patterns:

```go
// Transaction validation happens in AddTransaction
bc.AddTransaction(tx) // Validates before adding to mempool

// Mining executes all pending transactions
bc.MineBlock() // Mines block + executes all transactions

// Get account state
balance := bc.GetBalance(address)
nonce := bc.GetNonce(address)
```

---

### 2. `evm` Package

**Purpose**: Ethereum Virtual Machine implementation for smart contract execution.

#### Key Components:

**Contract** (`contract.go`):
- Address generated from hash(owner + bytecode)
- Contains bytecode, storage, owner, and balance
- `Execute(gas uint64)` runs bytecode with gas metering
- Storage is persistent key-value store (big.Int → big.Int)

**Interpreter** (`interpreter.go`):
- Singleton pattern: `GlobalInterpreter`
- `Run(ctx *ExecutionContext)` executes bytecode
- Tracks: PC (program counter), gas, stack, memory, storage
- Verbose mode prints execution trace
- Gas metering on every opcode

**OpCodes** (`opcodes.go`):
- Ethereum-compatible opcodes (same hex values)
- Categories:
  - **Arithmetic**: ADD, MUL, SUB, DIV, MOD
  - **Comparison**: LT, GT, EQ
  - **Stack**: POP, PUSH1-PUSH32, DUP1-DUP2, SWAP1-SWAP2
  - **Memory**: MLOAD, MSTORE
  - **Storage**: SLOAD (200 gas), SSTORE (20000 gas)
  - **Control**: STOP, JUMP, JUMPI
- Gas costs defined in `gasCosts` map

**Stack** (`stack.go`):
- Uses `[]*big.Int` for 256-bit integers
- Operations: Push, Pop, Peek, Len
- No size limit (differs from real Ethereum's 1024 limit)

**Memory** (`memory.go`):
- Byte-addressed expandable memory
- `Load(offset, size)` and `Store(offset, data)`
- Auto-expands on access

**Storage** (`storage.go`):
- Persistent key-value store: `map[string]*big.Int`
- Keys are hex-encoded big.Int
- Supports snapshots for transaction reverts
- SSTORE is most expensive operation (20000 gas)

---

### 3. `compiler` Package

**Purpose**: Converts human-readable assembly to EVM bytecode.

**Assembler** (`assembler.go`):

```go
assembler := compiler.NewAssembler()
bytecode, err := assembler.Assemble(assemblyCode)
```

**Assembly Syntax**:
```assembly
PUSH1 100       # Push decimal 100
PUSH1 0x64      # Push hex 0x64 (same as 100)
PUSH1 0         # Push 0
SSTORE          # Store 100 at storage[0]
STOP            # End execution
```

**Features**:
- Supports decimal and hex values (0x prefix)
- Auto-handles PUSH sizing (PUSH1 for values ≤255)
- Line-by-line parsing with error reporting
- `Disassemble()` for reverse conversion (bytecode → assembly)

---

### 4. `crypto` Package

**Purpose**: Cryptographic key management and signing.

**KeyPair** (`keypair.go`):
- ECDSA key generation (secp256k1-compatible)
- `GenerateKeyPair()` creates new random keypair
- `GetAddress()` derives address from public key (hex-encoded)
- `SignData(data []byte)` creates signature
- `VerifySignature()` validates signatures

**Wallet** (`wallet.go`):
- Manages multiple keypairs: `map[string]*KeyPair`
- `CreateAccount()` generates new account
- `GetKeyPair(address)` retrieves keypair for signing
- `ListAccounts()` shows all managed accounts

---

### 5. `utils` Package

**Purpose**: Common utilities.

**crypto.go**:
- `CalculateHash(data string) string` - SHA256 hash
- `MeetsTarget(hash string, difficulty int)` - Checks if hash starts with N zeros

---

### 6. Main Application (`minichain.go`)

**Interactive CLI Menu**:
1. View wallet accounts
2. Create new account
3. View account balances/nonces
4. Create transaction (with signing)
5. View pending transactions
6. Mine block
7. View entire blockchain
8. Verify blockchain integrity
9. Exit

**Smart Contract Menu**:
10. Deploy contract directly (no transaction)
11. List deployed contracts
12. Execute contract directly (no transaction)
13. View contract state
14. **TX: Deploy contract** (via transaction)
15. **TX: Call contract** (via transaction)

**Workflow**:
```
1. Create accounts → Get initial balances
2. Create transaction → Signs with keypair → Adds to mempool
3. Mine block → Executes transactions → Updates state
4. Deploy contract → Write assembly → Compile → Deploy/Mine
5. Call contract → Execute via transaction → Mine to finalize
```

---

## Key Development Conventions

### 1. Language Mixing
- **Code**: English (variables, functions, types)
- **Comments**: Spanish (implementation notes, explanations)
- **User-facing messages**: Spanish (CLI output)

Example:
```go
// Crear el bloque génesis (bloque #0)
func NewGenesisBlock() *Block {
    return &Block{
        Index:        0,
        Timestamp:    time.Now(),
        // ...
    }
}
```

### 2. Naming Conventions
- **Structs**: PascalCase (`Block`, `Transaction`, `AccountState`)
- **Functions**: camelCase for private, PascalCase for exported
- **Variables**: camelCase
- **Constants**: UPPER_SNAKE_CASE for opcodes

### 3. Error Handling
- Always return descriptive errors with context
- Use `fmt.Errorf()` for formatted errors
- Validate inputs before processing
- Example: `return fmt.Errorf("saldo insuficiente: tiene %.2f, necesita %.2f", balance, amount)`

### 4. Printing Patterns
- Use Unicode box-drawing characters for visual structure
- Emojis for semantic meaning (🔗, 💰, ✅, ❌, ⛏️, etc.)
- Truncate addresses/hashes for display: `address[:16]+"..."`

Example:
```go
fmt.Println("╔════════════════════════════════════════╗")
fmt.Println("║         ESTADO DE CUENTAS              ║")
fmt.Println("╚════════════════════════════════════════╝")
```

### 5. Float64 for Balances
- Uses `float64` for account balances (not wei/gwei like Ethereum)
- Currency unit: "MTC" (MiniCoin)
- Gas price: `0.000001` MTC per gas unit

---

## Important Implementation Details

### Transaction Lifecycle

```
┌─────────────────┐
│ Create TX       │ → NewTransaction(from, to, amount, nonce)
└────────┬────────┘
         ↓
┌─────────────────┐
│ Sign TX         │ → tx.Sign(keyPair)
└────────┬────────┘
         ↓
┌─────────────────┐
│ Validate        │ → tx.Validate(accountState, blockchain)
└────────┬────────┘
         ↓
┌─────────────────┐
│ Add to Mempool  │ → blockchain.AddTransaction(tx)
└────────┬────────┘
         ↓
┌─────────────────┐
│ Mine Block      │ → blockchain.MineBlock()
└────────┬────────┘
         ↓
┌─────────────────┐
│ Execute TX      │ → tx.Execute(accountState, blockchain)
└─────────────────┘
```

### Contract Deployment Flow

1. **Write Assembly**:
```assembly
PUSH1 100
PUSH1 0
SSTORE
STOP
```

2. **Compile to Bytecode**:
```go
assembler := compiler.NewAssembler()
bytecode, err := assembler.Assemble(code)
// bytecode = [0x60, 0x64, 0x60, 0x00, 0x55, 0x00]
```

3. **Create Deployment Transaction**:
```go
tx := blockchain.NewContractDeploymentTx(fromAddress, bytecode, nonce)
tx.Sign(keyPair)
blockchain.AddTransaction(tx)
```

4. **Mine Block**:
```go
blockchain.MineBlock()
// Executes deployment, creates contract, assigns address
```

5. **Contract Address Calculation**:
```go
address := hash(owner + bytecode)[:40]
```

### Gas Model

**Gas Costs** (key operations):
- Simple transfer: 21,000 gas
- SSTORE (write storage): 20,000 gas
- SLOAD (read storage): 200 gas
- Arithmetic ops: 3-5 gas
- Contract deployment: 32,000 + (bytecode_size × 200)

**Gas Price**: `0.000001 MTC per gas unit`

**Transaction Execution**:
1. Reserve max gas: `balance -= (gasLimit × gasPrice)`
2. Execute transaction
3. Calculate used gas
4. Refund unused: `balance += ((gasLimit - gasUsed) × gasPrice)`
5. On failure: No refund, consume all gas

---

## Common Pitfalls & Solutions

### 1. Typo in Filename
- `transacction.go` has double 'c' - this is intentional, don't rename without updating imports

### 2. Address Truncation
- Always check string length before truncating: `address[:16]`
- Use safe truncation:
```go
if len(address) >= 16 {
    fmt.Printf("%s...", address[:16])
} else {
    fmt.Printf("%s", address)
}
```

### 3. Nonce Management
- Nonce increments even on failed transactions (Ethereum behavior)
- Always use `bc.GetNonce(address)` for current nonce
- Never reuse nonces (replay protection)

### 4. Contract Execution Context
- Contracts need `Storage` reference to persist state
- Use `GlobalInterpreter` singleton for execution
- Set `Verbose: true` for debugging, `false` for production

### 5. Scanner Issues in CLI
- Create new `bufio.NewScanner(os.Stdin)` for nested input loops
- Use `strings.TrimSpace()` on all user input
- Handle "FIN" terminator for multi-line input

---

## Development Workflow

### Building & Running

```bash
# Build
go build -o minichain minichain.go

# Run
./minichain

# Or run directly
go run minichain.go
```

### Testing Smart Contracts

**Example: Counter Contract**
```assembly
PUSH1 0         # Push storage slot 0
SLOAD           # Load current value
PUSH1 1         # Push 1
ADD             # Increment
PUSH1 0         # Push storage slot 0
SSTORE          # Store new value
STOP            # End
```

**Testing Steps**:
1. Create account with balance
2. Write assembly code
3. Choose "14. TX: Desplegar contrato"
4. Write assembly, terminate with "FIN"
5. Sign and add to mempool
6. Mine block (option 6)
7. View contract state (option 13)
8. Call contract (option 15)
9. Mine again to execute call
10. View state to see counter increment

### Debugging

**Enable Verbose Mode**:
- Edit contract execution to set `Verbose: true` in `ExecutionContext`
- Shows: PC, opcode, stack operations, gas consumption

**Check Blockchain Integrity**:
- Option 8 in menu validates entire chain
- Checks: hash validity, difficulty compliance, chain linkage

**View Pending Transactions**:
- Option 5 shows all transactions before mining
- Verify signatures, amounts, nonces

---

## AI Assistant Guidelines

### When Adding Features

1. **Understand Context First**:
   - Read related files completely
   - Check how existing code handles similar cases
   - Look for patterns in error handling and validation

2. **Maintain Consistency**:
   - Use Spanish comments for implementation notes
   - Follow existing printing patterns with Unicode boxes
   - Use emoji conventions consistently
   - Keep English code/variable names

3. **Error Handling**:
   - Always validate inputs
   - Return descriptive errors with context
   - Handle edge cases (empty strings, zero values, nil pointers)

4. **Testing Approach**:
   - Test via CLI menu options
   - Create multiple accounts for testing
   - Verify state changes after mining
   - Check gas calculations

### When Fixing Bugs

1. **Reproduce First**:
   - Use CLI to recreate issue
   - Check state before and after
   - Review transaction execution logs

2. **Check Related Systems**:
   - Account state management
   - Nonce handling
   - Gas calculations
   - Storage snapshots for reverts

3. **Verify Fix**:
   - Test happy path
   - Test error cases
   - Verify blockchain integrity (option 8)
   - Check no regressions in related features

### When Refactoring

1. **Don't Break Existing Behavior**:
   - Maintain Spanish comments and CLI messages
   - Keep emoji patterns
   - Preserve gas costs unless intentional
   - Don't rename `transacction.go` (breaks imports)

2. **Improve Incrementally**:
   - Refactor one package at a time
   - Run full tests after each change
   - Update comments to match code

### Code Review Checklist

- [ ] Spanish comments for implementation details
- [ ] English code and variable names
- [ ] Proper error handling with context
- [ ] Address truncation safety checks
- [ ] Nonce management correctness
- [ ] Gas calculation accuracy
- [ ] State snapshot/revert logic
- [ ] Unicode box formatting consistency
- [ ] Emoji semantic meaning matches convention

---

## Architecture Decisions

### Why Float64 for Balances?
- Simpler for educational purposes
- Avoids wei/gwei complexity
- Sufficient precision for demo blockchain
- **Tradeoff**: Not suitable for production (precision issues)

### Why Singleton Interpreter?
- Gas table shared across all executions
- Stateless execution (state in Context)
- Simpler than passing interpreter everywhere
- **Pattern**: `GlobalInterpreter.Run(ctx)`

### Why Snapshots for Reverts?
- Ethereum-style transaction atomicity
- Failed transactions don't affect state (except nonce)
- Enables complex contract interactions
- **Cost**: Memory overhead for large states

### Why Address from Hash?
- Deterministic contract addresses
- Prevents address collisions
- Links contract to creator and code
- **Formula**: `hash(owner + bytecode)[:40]`

---

## Future Considerations

### Not Yet Implemented

1. **Networking**: No P2P protocol (single node only)
2. **Persistence**: No database (state lost on exit)
3. **Advanced EVM Features**:
   - CALL/DELEGATECALL opcodes
   - Contract-to-contract calls
   - Events/logs
   - Precompiled contracts
4. **Consensus**: Only PoW, no PoS or other algorithms
5. **MEV Protection**: No transaction privacy
6. **Fork Handling**: No chain reorganization logic

### Extensibility Points

1. **Add New Opcodes**:
   - Define in `evm/opcodes.go`
   - Implement in `interpreter.go`
   - Add gas cost to `gasCosts` map
   - Update `assembler.go` opcodeMap

2. **Change Consensus**:
   - Modify `block.go` mining logic
   - Update `MineBlock()` validation
   - Adjust difficulty algorithm

3. **Add Transaction Types**:
   - Extend `Transaction` struct
   - Add validation in `Validate()`
   - Handle in `Execute()`
   - Update CLI menu

4. **Storage Backend**:
   - Replace `map[string]*Account` with DB
   - Implement state trie (Merkle Patricia)
   - Add persistence layer

---

## Quick Reference

### Key File Locations

| Component | File | Key Functions |
|-----------|------|---------------|
| Blockchain Core | `blockchain/blockchain.go` | `NewBlockchain()`, `MineBlock()`, `AddTransaction()` |
| Mining | `blockchain/block.go` | `MineBlock()`, `CalculateBlockHash()` |
| Transactions | `blockchain/transacction.go` | `Sign()`, `Validate()`, `Execute()` |
| Accounts | `blockchain/account.go` | `GetAccount()`, `AddBalance()`, `SubtractBalance()` |
| Smart Contracts | `evm/contract.go` | `NewContract()`, `Execute()` |
| EVM Execution | `evm/interpreter.go` | `Run()`, `ExecuteOpcode()` |
| Assembly Compiler | `compiler/assembler.go` | `Assemble()`, `Disassemble()` |
| Wallet | `crypto/wallet.go` | `CreateAccount()`, `GetKeyPair()` |
| Cryptography | `crypto/keypair.go` | `GenerateKeyPair()`, `SignData()` |
| Main App | `minichain.go` | `main()` (CLI menu) |

### Common Operations Code Snippets

**Create & Mine Transaction**:
```go
// Create account
wallet := crypto.NewWallet()
address, _ := wallet.CreateAccount()

// Fund account
bc.AccountState.AddBalance(address, 100.0)

// Create transaction
nonce := bc.GetNonce(address)
tx := blockchain.NewTransaction(from, to, amount, nonce)

// Sign & submit
keyPair, _ := wallet.GetKeyPair(from)
tx.Sign(keyPair)
bc.AddTransaction(tx)

// Mine
bc.MineBlock()
```

**Deploy & Execute Contract**:
```go
// Compile assembly
assembler := compiler.NewAssembler()
bytecode, _ := assembler.Assemble(`
PUSH1 100
PUSH1 0
SSTORE
STOP
`)

// Deploy via transaction
tx := blockchain.NewContractDeploymentTx(owner, bytecode, nonce)
tx.Sign(keyPair)
bc.AddTransaction(tx)
bc.MineBlock()

// Call contract
contractAddr := tx.ContractAddress
callTx := blockchain.NewContractCallTx(caller, contractAddr, []byte{}, nonce)
callTx.Sign(keyPair)
bc.AddTransaction(callTx)
bc.MineBlock()
```

---

## Conclusion

Minichain is an educational blockchain implementation demonstrating:
- **Core Concepts**: PoW mining, transaction signing, account state
- **Smart Contracts**: EVM-compatible execution with assembly language
- **Gas Metering**: Ethereum-style gas costs and refunds
- **Transaction Atomicity**: State snapshots and reverts

**Best Use Cases**:
- Learning blockchain internals
- Experimenting with smart contracts
- Understanding EVM execution
- Prototyping consensus algorithms

**Not Suitable For**:
- Production use
- Real financial transactions
- High-performance requirements
- Multi-node networks

---

*This document is intended for AI assistants and human developers working on the Minichain codebase. Keep it updated as the codebase evolves.*
