package trie

// SecureTrie es un wrapper sobre Trie que hace hash de las keys
// Esto es lo que usa Ethereum para el state trie
// Las keys originales se guardan en un "preimage store" para poder recuperarlas
// Basado en go-ethereum/trie/secure_trie.go
type SecureTrie struct {
	trie *Trie

	// Preimage store para recuperar keys originales desde sus hashes
	preimages map[string][]byte
	db        *Database
}

// NewSecure crea un nuevo secure trie
func NewSecure(root []byte, db *Database) (*SecureTrie, error) {
	trie, err := New(root, db)
	if err != nil {
		return nil, err
	}

	return &SecureTrie{
		trie:      trie,
		preimages: make(map[string][]byte),
		db:        db,
	}, nil
}

// Get retorna el valor asociado a la key
// La key se hashea antes de buscarla
func (t *SecureTrie) Get(key []byte) []byte {
	res, err := t.TryGet(key)
	if err != nil {
		panic("secure trie get error: " + err.Error())
	}
	return res
}

// TryGet retorna el valor asociado a la key, con manejo de errores
func (t *SecureTrie) TryGet(key []byte) ([]byte, error) {
	return t.trie.TryGet(t.hashKey(key))
}

// Update asocia key con value en el trie
// La key se hashea antes de insertarla
func (t *SecureTrie) Update(key, value []byte) {
	if err := t.TryUpdate(key, value); err != nil {
		panic("secure trie update error: " + err.Error())
	}
}

// TryUpdate asocia key con value, con manejo de errores
func (t *SecureTrie) TryUpdate(key, value []byte) error {
	hk := t.hashKey(key)
	// Guardar preimage
	t.preimages[string(hk)] = key
	return t.trie.TryUpdate(hk, value)
}

// Delete elimina una key del trie
func (t *SecureTrie) Delete(key []byte) {
	if err := t.TryDelete(key); err != nil {
		panic("secure trie delete error: " + err.Error())
	}
}

// TryDelete elimina una key, con manejo de errores
func (t *SecureTrie) TryDelete(key []byte) error {
	hk := t.hashKey(key)
	delete(t.preimages, string(hk))
	return t.trie.TryDelete(hk)
}

// Hash retorna el hash root del trie
func (t *SecureTrie) Hash() []byte {
	return t.trie.Hash()
}

// Commit escribe todos los nodos del trie a la database
// También escribe los preimages
func (t *SecureTrie) Commit() ([]byte, error) {
	// Escribir preimages a la database
	if len(t.preimages) > 0 {
		batch := t.db.db.NewBatch()
		for hash, key := range t.preimages {
			// Prefijo "secure-key-" para preimages
			preimageKey := append([]byte("secure-key-"), []byte(hash)...)
			if err := batch.Put(preimageKey, key); err != nil {
				return nil, err
			}
		}
		if err := batch.Write(); err != nil {
			return nil, err
		}
	}

	// Limpiar preimages después de commit
	t.preimages = make(map[string][]byte)

	// Commit del trie
	return t.trie.Commit()
}

// GetKey retorna la key original desde su hash (si existe en preimages)
func (t *SecureTrie) GetKey(shaKey []byte) []byte {
	if key, ok := t.preimages[string(shaKey)]; ok {
		return key
	}

	// Intentar cargar desde database
	preimageKey := append([]byte("secure-key-"), shaKey...)
	key, err := t.db.db.Get(preimageKey)
	if err == nil {
		return key
	}

	return nil
}

// hashKey calcula Keccak256 de la key
func (t *SecureTrie) hashKey(key []byte) []byte {
	return Keccak256(key)
}

// Root retorna el hash root del trie
func (t *SecureTrie) Root() []byte {
	return t.Hash()
}

// Copy crea una copia del secure trie
func (t *SecureTrie) Copy() *SecureTrie {
	// Crear nuevo trie con mismo root
	newTrie, err := NewSecure(t.Hash(), t.db)
	if err != nil {
		panic("copy error: " + err.Error())
	}

	// Copiar preimages
	for k, v := range t.preimages {
		newTrie.preimages[k] = v
	}

	return newTrie
}
