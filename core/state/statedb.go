package state

import (
	"fmt"
	"math/big"
	"minichain/rlp"
	"minichain/trie"
)

// StateDB gestiona el estado de todas las cuentas
// Basado en go-ethereum/core/state/statedb.go
type StateDB struct {
	db   Database              // Database wrapper
	trie *trie.SecureTrie      // Main state trie

	// State objects cache
	stateObjects map[string]*stateObject

	// Logging de cambios
	logs    []*Log
	logSize uint

	// Tracking
	refund uint64 // Gas refund acumulado
}

// New crea un nuevo StateDB
func New(root []byte, db Database) (*StateDB, error) {
	tr, err := trie.NewSecure(root, db.TrieDB())
	if err != nil {
		return nil, err
	}

	return &StateDB{
		db:           db,
		trie:         tr,
		stateObjects: make(map[string]*stateObject),
	}, nil
}

// getStateObject obtiene o carga un state object
func (s *StateDB) getStateObject(addr []byte) *stateObject {
	// Buscar en caché
	if obj, ok := s.stateObjects[string(addr)]; ok {
		return obj
	}

	// Cargar desde el trie
	data := s.trie.Get(addr)
	if len(data) == 0 {
		return nil
	}

	// Decodificar account
	var acc Account
	if err := rlp.Decode(data, &acc); err != nil {
		return nil
	}

	// Crear state object
	obj := newObject(s, addr, acc)
	s.stateObjects[string(addr)] = obj

	return obj
}

// getOrNewStateObject obtiene o crea un nuevo state object
func (s *StateDB) getOrNewStateObject(addr []byte) *stateObject {
	obj := s.getStateObject(addr)
	if obj == nil {
		obj = s.createObject(addr)
	}
	return obj
}

// createObject crea un nuevo state object
func (s *StateDB) createObject(addr []byte) *stateObject {
	newObj := newObject(s, addr, *NewAccount())
	s.stateObjects[string(addr)] = newObj
	return newObj
}

// Exist verifica si una cuenta existe
func (s *StateDB) Exist(addr []byte) bool {
	return s.getStateObject(addr) != nil
}

// Empty verifica si una cuenta está vacía
func (s *StateDB) Empty(addr []byte) bool {
	so := s.getStateObject(addr)
	return so == nil || so.empty()
}

// GetBalance obtiene el saldo de una cuenta
func (s *StateDB) GetBalance(addr []byte) *big.Int {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return big.NewInt(0)
}

// SetBalance establece el saldo de una cuenta
func (s *StateDB) SetBalance(addr []byte, amount *big.Int) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	}
}

// AddBalance añade al saldo de una cuenta
func (s *StateDB) AddBalance(addr []byte, amount *big.Int) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

// SubBalance resta del saldo de una cuenta
func (s *StateDB) SubBalance(addr []byte, amount *big.Int) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(amount)
	}
}

// GetNonce obtiene el nonce de una cuenta
func (s *StateDB) GetNonce(addr []byte) uint64 {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}
	return 0
}

// SetNonce establece el nonce de una cuenta
func (s *StateDB) SetNonce(addr []byte, nonce uint64) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

// GetCode obtiene el código de un contrato
func (s *StateDB) GetCode(addr []byte) []byte {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Code()
	}
	return nil
}

// GetCodeHash obtiene el hash del código de un contrato
func (s *StateDB) GetCodeHash(addr []byte) []byte {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.data.CodeHash
	}
	return trie.Keccak256(nil)
}

// SetCode establece el código de un contrato
func (s *StateDB) SetCode(addr []byte, code []byte) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(code)
	}
}

// GetState obtiene un valor del storage de un contrato
func (s *StateDB) GetState(addr []byte, key []byte) []byte {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetState(key)
	}
	return nil
}

// SetState establece un valor en el storage de un contrato
func (s *StateDB) SetState(addr []byte, key []byte, value []byte) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(key, value)
	}
}

// Suicide marca una cuenta para ser eliminada
func (s *StateDB) Suicide(addr []byte) bool {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return false
	}
	stateObject.suicided = true
	stateObject.data.Balance = new(big.Int)
	return true
}

// HasSuicided verifica si una cuenta se autodestruyó
func (s *StateDB) HasSuicided(addr []byte) bool {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

// AddRefund añade gas refund
func (s *StateDB) AddRefund(gas uint64) {
	s.refund += gas
}

// SubRefund resta gas refund
func (s *StateDB) SubRefund(gas uint64) {
	if gas > s.refund {
		s.refund = 0
	} else {
		s.refund -= gas
	}
}

// GetRefund obtiene el gas refund acumulado
func (s *StateDB) GetRefund() uint64 {
	return s.refund
}

// Commit escribe todos los cambios al trie y retorna el nuevo root
func (s *StateDB) Commit() ([]byte, error) {
	// Commit de todos los state objects
	for addr, stateObject := range s.stateObjects {
		if stateObject.suicided {
			// Eliminar cuenta suicidada
			s.trie.Delete([]byte(addr))
		} else if !stateObject.empty() {
			// Commit del state object
			if err := stateObject.commit(); err != nil {
				return nil, err
			}

			// Codificar account data
			data, err := rlp.Encode(stateObject.data)
			if err != nil {
				return nil, err
			}

			// Actualizar en el trie
			s.trie.Update([]byte(addr), data)
		}
	}

	// Commit del trie principal
	root, err := s.trie.Commit()
	if err != nil {
		return nil, err
	}

	// Limpiar caché
	s.stateObjects = make(map[string]*stateObject)
	s.refund = 0

	return root, nil
}

// IntermediateRoot calcula el root sin hacer commit
func (s *StateDB) IntermediateRoot() []byte {
	// Finalizar todos los state objects
	for _, stateObject := range s.stateObjects {
		if !stateObject.empty() {
			stateObject.updateStorageTrie()
		}
	}

	return s.trie.Hash()
}

// Root retorna el root hash actual
func (s *StateDB) Root() []byte {
	return s.trie.Hash()
}

// Copy crea una copia del StateDB
func (s *StateDB) Copy() *StateDB {
	// Crear nuevo StateDB con mismo root
	state, err := New(s.Root(), s.db)
	if err != nil {
		panic(fmt.Sprintf("failed to copy state: %v", err))
	}

	// Copiar state objects
	for addr, obj := range s.stateObjects {
		state.stateObjects[addr] = &stateObject{
			address:      obj.address,
			data:         *obj.data.Copy(),
			db:           state,
			dirtyStorage: make(map[string][]byte),
		}
		// Copiar dirty storage
		for k, v := range obj.dirtyStorage {
			state.stateObjects[addr].dirtyStorage[k] = v
		}
	}

	state.refund = s.refund

	return state
}

// Log representa un log de evento
type Log struct {
	Address []byte
	Topics  [][]byte
	Data    []byte
}

// AddLog añade un log
func (s *StateDB) AddLog(log *Log) {
	s.logs = append(s.logs, log)
}

// GetLogs retorna todos los logs
func (s *StateDB) GetLogs() []*Log {
	return s.logs
}
