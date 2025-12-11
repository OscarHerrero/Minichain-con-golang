package blockchain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"minichain/crypto"
)

// Transaction representa una transacción en la blockchain
type Transaction struct {
	From       string
	To         string // Si es "", es despliegue de contrato
	Amount     float64
	Nonce      int
	Data       []byte // Bytecode (para deploy) o calldata (para call)
	Signature  string
	PublicKeyX *big.Int
	PublicKeyY *big.Int

	// Metadata de ejecución
	ContractAddress string // Si despliega contrato, guarda la dirección aquí
	GasUsed         uint64 // Gas consumido en la ejecución
}

// IsContractDeployment verifica si es una transacción de despliegue
func (tx *Transaction) IsContractDeployment() bool {
	return tx.To == "" && len(tx.Data) > 0
}

// IsContractCall verifica si es una llamada a contrato
func (tx *Transaction) IsContractCall(bc *Blockchain) bool {
	if tx.To == "" {
		return false
	}

	// Verificar si el destinatario es un contrato
	_, err := bc.GetContract(tx.To)
	return err == nil
}

// NewTransaction crea una nueva transacción (sin firmar)
func NewTransaction(from, to string, amount float64, nonce int) *Transaction {
	return &Transaction{
		From:   from,
		To:     to,
		Amount: amount,
		Nonce:  nonce,
	}
}

// Sign firma la transacción con un par de claves
func (tx *Transaction) Sign(keyPair *crypto.KeyPair) error {
	// Verificar que la dirección coincide con el par de claves
	if tx.From != keyPair.GetAddress() {
		return fmt.Errorf("la dirección From no coincide con el par de claves")
	}

	// Guardar la clave pública (necesaria para verificar la firma)
	tx.PublicKeyX = keyPair.PublicKey.X
	tx.PublicKeyY = keyPair.PublicKey.Y

	// Crear los datos a firmar (sin la firma misma)
	dataToSign := tx.getDataToSign()

	// Firmar los datos
	signature, err := keyPair.SignData(dataToSign)
	if err != nil {
		return fmt.Errorf("error firmando transacción: %v", err)
	}

	tx.Signature = signature

	return nil
}

// getDataToSign obtiene los datos que se firman
// No incluye la firma misma (obvio, no puedes firmar la firma)
func (tx *Transaction) getDataToSign() []byte {
	data := fmt.Sprintf("%s:%s:%.2f:%d", tx.From, tx.To, tx.Amount, tx.Nonce)
	return []byte(data)
}

// VerifySignature verifica que la firma sea válida
func (tx *Transaction) VerifySignature() bool {
	if tx.Signature == "" {
		return false
	}

	if tx.PublicKeyX == nil || tx.PublicKeyY == nil {
		return false
	}

	// Obtener los datos que fueron firmados
	dataToSign := tx.getDataToSign()

	// Verificar la firma
	return crypto.VerifySignature(tx.PublicKeyX, tx.PublicKeyY, dataToSign, tx.Signature)
}

// Validate valida la transacción antes de añadirla al mempool
func (tx *Transaction) Validate(state *AccountState, bc *Blockchain) error {
	// Verificar que esté firmada
	if tx.Signature == "" {
		return fmt.Errorf("transacción no firmada")
	}

	// Verificar la firma
	if !tx.VerifySignature() {
		return fmt.Errorf("firma inválida")
	}

	// Verificar que el monto no sea negativo
	if tx.Amount < 0 {
		return fmt.Errorf("monto no puede ser negativo: %.2f", tx.Amount)
	}

	// Determinar tipo de transacción y validar
	isContractDeployment := tx.IsContractDeployment()
	isContractCall := tx.IsContractCall(bc)

	// Validar que la transacción tenga propósito
	if !isContractDeployment && !isContractCall && tx.Amount == 0 {
		return fmt.Errorf("transacción sin propósito: sin monto, sin deploy, sin llamada")
	}

	// Verificar que el nonce sea correcto
	account := state.GetAccount(tx.From)
	expectedNonce := account.Nonce

	if tx.Nonce != expectedNonce {
		return fmt.Errorf("nonce incorrecto: esperado %d, recibido %d", expectedNonce, tx.Nonce)
	}

	// Verificar saldo suficiente (solo si hay transferencia de fondos)
	if tx.Amount > 0 {
		if account.Balance < tx.Amount {
			return fmt.Errorf("saldo insuficiente: %.2f < %.2f", account.Balance, tx.Amount)
		}
	}

	return nil
}

// Execute ejecuta la transacción con lógica de revert (como Ethereum)
func (tx *Transaction) Execute(state *AccountState, bc *Blockchain) error {
	gasPrice := 0.000001 // 1 gas = 0.000001 MTC

	// ====================================
	// FASE 1: VALIDACIONES PREVIAS
	// ====================================

	account := state.GetAccount(tx.From)

	// Calcular gas máximo necesario
	var gasLimit uint64
	if tx.IsContractDeployment() {
		baseGas := uint64(32000)
		bytecodeGas := uint64(len(tx.Data)) * 200
		gasLimit = baseGas + bytecodeGas
	} else if len(tx.Data) > 0 || tx.IsContractCall(bc) {
		gasLimit = 1000000 // Gas límite para ejecución
	} else {
		gasLimit = 21000 // Gas base para transferencia simple
	}

	maxGasCost := float64(gasLimit) * gasPrice

	// Verificar saldo para: monto + gas máximo
	totalNeeded := tx.Amount + maxGasCost
	if account.Balance < totalNeeded {
		return fmt.Errorf("saldo insuficiente: tiene %.6f MTC, necesita %.6f MTC (monto: %.2f + gas máximo: %.6f)",
			account.Balance, totalNeeded, tx.Amount, maxGasCost)
	}

	// ====================================
	// FASE 2: CREAR SNAPSHOTS
	// ====================================

	accountSnapshot := state.CreateSnapshot()

	var storageSnapshots map[string]map[string]*big.Int
	if tx.IsContractCall(bc) {
		storageSnapshots = make(map[string]map[string]*big.Int)
		contract, _ := bc.GetContract(tx.To)
		if contract != nil {
			storageSnapshots[tx.To] = contract.Storage.CreateSnapshot()
		}
	}

	// ====================================
	// FASE 3: RESERVAR GAS
	// ====================================

	// Reservar gas máximo
	if err := state.SubtractBalance(tx.From, maxGasCost); err != nil {
		return err
	}

	// ====================================
	// FASE 4: INCREMENTAR NONCE (NO SE REVIERTE)
	// ====================================

	state.IncrementNonce(tx.From)

	// ====================================
	// FASE 5: EJECUTAR TRANSACCIÓN
	// ====================================

	var executionError error

	// Transferir fondos si aplica
	if tx.Amount > 0 {
		if err := state.SubtractBalance(tx.From, tx.Amount); err != nil {
			executionError = err
		} else if tx.To != "" {
			state.AddBalance(tx.To, tx.Amount)
		}
	}

	// Ejecutar contrato si aplica
	if executionError == nil && (len(tx.Data) > 0 || tx.IsContractCall(bc)) {
		if err := tx.ExecuteContract(bc); err != nil {
			executionError = fmt.Errorf("error ejecutando contrato: %v", err)
		}

		// Si no se registró GasUsed, significa que falló antes de calcular
		if tx.GasUsed == 0 {
			tx.GasUsed = gasLimit // Consumir todo el gas
		}
	} else if executionError == nil {
		// Transacción simple - gas base
		tx.GasUsed = 21000
	}

	// ====================================
	// FASE 6: APLICAR O REVERTIR
	// ====================================

	if executionError != nil {
		// ❌ EJECUCIÓN FALLÓ - REVERTIR ESTADO
		fmt.Printf("   ❌ Error en ejecución: %v\n", executionError)
		fmt.Printf("   🔄 Revirtiendo cambios de estado...\n")

		// Revertir estado de cuentas (excepto nonce y gas)
		currentNonce := state.GetAccount(tx.From).Nonce
		currentBalance := state.GetAccount(tx.From).Balance

		state.RevertToSnapshot(accountSnapshot)

		// Restaurar nonce (debe quedar incrementado)
		state.GetAccount(tx.From).Nonce = currentNonce

		// El gas YA fue restado, no lo devolvemos
		state.GetAccount(tx.From).Balance = currentBalance

		// Revertir storage de contratos
		for contractAddr, snapshot := range storageSnapshots {
			contract, _ := bc.GetContract(contractAddr)
			if contract != nil {
				contract.Storage.RevertToSnapshot(snapshot)
			}
		}

		// Consumir TODO el gas (penalización)
		tx.GasUsed = gasLimit
		gasCostUsed := float64(tx.GasUsed) * gasPrice

		fmt.Printf("   ⛽ Gas consumido (penalización): %.6f MTC (%d gas)\n", gasCostUsed, tx.GasUsed)

		// El gas ya fue restado, así que no hacemos nada más

	} else {
		// ✅ EJECUCIÓN EXITOSA
		gasCostUsed := float64(tx.GasUsed) * gasPrice
		gasRefund := maxGasCost - gasCostUsed

		// Devolver gas no usado
		if gasRefund > 0 {
			state.AddBalance(tx.From, gasRefund)
			fmt.Printf("   ⛽ Gas usado: %.6f MTC (%d gas)\n", gasCostUsed, tx.GasUsed)
			fmt.Printf("   💰 Gas devuelto: %.6f MTC\n", gasRefund)
		} else {
			fmt.Printf("   ⛽ Costo de gas: %.6f MTC (%d gas × %.6f)\n",
				gasCostUsed, tx.GasUsed, gasPrice)
		}
	}

	return nil
}

// ToJSON convierte la transacción a JSON
func (tx *Transaction) ToJSON() (string, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Print muestra la transacción de forma bonita
func (tx *Transaction) Print() {
	fmt.Println("\n┌────────────────────────────────────────┐")
	fmt.Println("│          💸 TRANSACCIÓN                │")
	fmt.Println("└────────────────────────────────────────┘")
	fmt.Printf("📤 From:      %s\n", tx.From[:16]+"...")
	fmt.Printf("📥 To:        %s\n", tx.To[:16]+"...")
	fmt.Printf("💰 Amount:    %.2f MTC\n", tx.Amount)
	fmt.Printf("🔢 Nonce:     %d\n", tx.Nonce)

	if tx.Signature != "" {
		fmt.Printf("✍️  Signature: %s...\n", tx.Signature[:16])
		fmt.Printf("✅ Firmada:   Sí\n")
		if tx.VerifySignature() {
			fmt.Printf("🔐 Válida:    Sí\n")
		} else {
			fmt.Printf("❌ Válida:    No\n")
		}
	} else {
		fmt.Printf("⚠️  Signature: (sin firmar)\n")
	}
}

// NewContractDeploymentTx crea una transacción para desplegar un contrato
func NewContractDeploymentTx(from string, bytecode []byte, nonce int) *Transaction {
	return &Transaction{
		From:   from,
		To:     "", // Vacío = deploy
		Amount: 0,
		Nonce:  nonce,
		Data:   bytecode,
	}
}

// NewContractCallTx crea una transacción para llamar a un contrato
func NewContractCallTx(from, contractAddr string, calldata []byte, nonce int) *Transaction {
	return &Transaction{
		From:   from,
		To:     contractAddr,
		Amount: 0,
		Nonce:  nonce,
		Data:   calldata,
	}
}

// ExecuteContract ejecuta un contrato (deploy o call)
func (tx *Transaction) ExecuteContract(bc *Blockchain) error {
	if tx.IsContractDeployment() {
		// DESPLEGAR CONTRATO
		contract, err := bc.DeployContract(tx.From, tx.Data)
		if err != nil {
			return fmt.Errorf("error desplegando contrato: %v", err)
		}

		// Guardar dirección del contrato en la transacción
		tx.ContractAddress = contract.Address

		// Cobrar gas por deployment (costo base)
		// En Ethereum real: ~32,000 gas por deploy + gas por bytecode
		baseGas := uint64(32000)
		bytecodeGas := uint64(len(tx.Data)) * 200 // 200 gas por byte
		tx.GasUsed = baseGas + bytecodeGas

		fmt.Printf("   📜 Contrato desplegado: %s\n", contract.Address[:16]+"...")
		fmt.Printf("   ⛽ Gas deployment: %d (base: %d + bytecode: %d)\n",
			tx.GasUsed, baseGas, bytecodeGas)

		return nil

	} else if tx.IsContractCall(bc) {
		// LLAMAR A CONTRATO
		contract, err := bc.GetContract(tx.To)
		if err != nil {
			return err
		}

		fmt.Printf("   ⚙️  Ejecutando contrato %s...\n\n", tx.To[:16]+"...")

		// Ejecutar con el intérprete global
		gasLeft, err := contract.Execute(1000000)
		if err != nil {
			return fmt.Errorf("error ejecutando contrato: %v", err)
		}

		tx.GasUsed = 1000000 - gasLeft
		fmt.Printf("\n   ✅ Contrato ejecutado. Gas usado: %d\n", tx.GasUsed)

		return nil
	}

	return nil
}
