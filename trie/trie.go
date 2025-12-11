package trie

import (
	"bytes"
	"fmt"
	"minichain/database"
)

// emptyRoot es el hash del trie vacío
var emptyRoot = Keccak256Hash(nil)

// Trie es un Merkle Patricia Trie
// Basado en go-ethereum/trie/trie.go
type Trie struct {
	db   *Database  // Database para persistir nodos
	root node       // Nodo raíz del trie

	// Caché de nodos unhashed (modificados pero no hasheados todavía)
	unhashed int
}

// New crea un nuevo trie vacío
func New(root []byte, db *Database) (*Trie, error) {
	if db == nil {
		panic("trie.New called without a database")
	}

	trie := &Trie{
		db: db,
	}

	if len(root) == 0 || bytes.Equal(root, emptyRoot) {
		// Trie vacío
		trie.root = nil
		return trie, nil
	}

	// Cargar root desde database
	rootnode, err := trie.resolveHash(root, nil)
	if err != nil {
		return nil, err
	}
	trie.root = rootnode

	return trie, nil
}

// Get retorna el valor asociado a la key
func (t *Trie) Get(key []byte) []byte {
	res, err := t.TryGet(key)
	if err != nil {
		panic(fmt.Sprintf("trie get error: %v", err))
	}
	return res
}

// TryGet retorna el valor asociado a la key, con manejo de errores
func (t *Trie) TryGet(key []byte) ([]byte, error) {
	key = keybytesToHex(key)
	value, _, err := t.tryGet(t.root, key, 0)
	return value, err
}

// tryGet es la implementación recursiva de Get
func (t *Trie) tryGet(origNode node, key []byte, pos int) (value []byte, newnode node, err error) {
	if origNode == nil {
		return nil, nil, nil
	}

	switch n := origNode.(type) {
	case *shortNode:
		if len(key)-pos < len(n.Key) || !bytes.Equal(n.Key, key[pos:pos+len(n.Key)]) {
			// Key no coincide
			return nil, n, nil
		}
		value, _, err = t.tryGet(n.Val, key, pos+len(n.Key))
		return value, n, err

	case *fullNode:
		value, _, err = t.tryGet(n.Children[key[pos]], key, pos+1)
		return value, n, err

	case hashNode:
		// Cargar nodo desde database
		child, err := t.resolveHash(n, key[:pos])
		if err != nil {
			return nil, n, err
		}
		value, _, err = t.tryGet(child, key, pos)
		return value, n, err

	case valueNode:
		return []byte(n), n, nil

	default:
		panic(fmt.Sprintf("invalid node type: %T", origNode))
	}
}

// Update asocia key con value en el trie
func (t *Trie) Update(key, value []byte) {
	if err := t.TryUpdate(key, value); err != nil {
		panic(fmt.Sprintf("trie update error: %v", err))
	}
}

// TryUpdate asocia key con value, con manejo de errores
func (t *Trie) TryUpdate(key, value []byte) error {
	t.unhashed++
	k := keybytesToHex(key)
	if len(value) != 0 {
		_, t.root, _ = t.insert(t.root, nil, k, valueNode(value))
	} else {
		_, t.root, _ = t.delete(t.root, nil, k)
	}
	return nil
}

// insert inserta un valor en el trie
func (t *Trie) insert(n node, prefix, key []byte, value node) (bool, node, error) {
	if len(key) == 0 {
		if v, ok := n.(valueNode); ok {
			return !bytes.Equal(v, value.(valueNode)), value, nil
		}
		return true, value, nil
	}

	switch n := n.(type) {
	case *shortNode:
		matchlen := prefixLen(key, n.Key)
		// Si la key coincide exactamente con el shortNode
		if matchlen == len(n.Key) {
			dirty, nn, err := t.insert(n.Val, append(prefix, key[:matchlen]...), key[matchlen:], value)
			if !dirty || err != nil {
				return false, n, err
			}
			return true, &shortNode{n.Key, nn, nodeFlag{dirty: true}}, nil
		}

		// Si no coincide completamente, necesitamos crear un branch node
		branch := &fullNode{flags: nodeFlag{dirty: true}}
		var err error

		// Agregar el nodo existente al branch
		_, branch.Children[n.Key[matchlen]], err = t.insert(nil, append(prefix, n.Key[:matchlen+1]...), n.Key[matchlen+1:], n.Val)
		if err != nil {
			return false, nil, err
		}

		// Agregar el nuevo valor al branch
		_, branch.Children[key[matchlen]], err = t.insert(nil, append(prefix, key[:matchlen+1]...), key[matchlen+1:], value)
		if err != nil {
			return false, nil, err
		}

		// Reemplazar el shortNode con un shortNode + branch
		if matchlen == 0 {
			return true, branch, nil
		}
		return true, &shortNode{key[:matchlen], branch, nodeFlag{dirty: true}}, nil

	case *fullNode:
		dirty, nn, err := t.insert(n.Children[key[0]], append(prefix, key[0]), key[1:], value)
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.flags = nodeFlag{dirty: true}
		n.Children[key[0]] = nn
		return true, n, nil

	case nil:
		// Crear nuevo leaf node
		return true, &shortNode{key, value, nodeFlag{dirty: true}}, nil

	case hashNode:
		// Cargar nodo desde database
		rn, err := t.resolveHash(n, prefix)
		if err != nil {
			return false, nil, err
		}
		dirty, nn, err := t.insert(rn, prefix, key, value)
		if !dirty || err != nil {
			return false, rn, err
		}
		return true, nn, nil

	default:
		panic(fmt.Sprintf("invalid node type: %T", n))
	}
}

// Delete elimina una key del trie
func (t *Trie) Delete(key []byte) {
	if err := t.TryDelete(key); err != nil {
		panic(fmt.Sprintf("trie delete error: %v", err))
	}
}

// TryDelete elimina una key, con manejo de errores
func (t *Trie) TryDelete(key []byte) error {
	t.unhashed++
	k := keybytesToHex(key)
	_, t.root, _ = t.delete(t.root, nil, k)
	return nil
}

// delete es la implementación recursiva de Delete
func (t *Trie) delete(n node, prefix, key []byte) (bool, node, error) {
	switch n := n.(type) {
	case *shortNode:
		matchlen := prefixLen(key, n.Key)
		if matchlen < len(n.Key) {
			return false, n, nil // Key no encontrada
		}
		if matchlen == len(key) {
			return true, nil, nil // Eliminar este nodo
		}

		// Continuar eliminación recursiva
		dirty, child, err := t.delete(n.Val, append(prefix, key[:len(n.Key)]...), key[len(n.Key):])
		if !dirty || err != nil {
			return false, n, err
		}

		switch child := child.(type) {
		case *shortNode:
			// Combinar shortNodes
			return true, &shortNode{concat(n.Key, child.Key...), child.Val, nodeFlag{dirty: true}}, nil
		default:
			return true, &shortNode{n.Key, child, nodeFlag{dirty: true}}, nil
		}

	case *fullNode:
		dirty, nn, err := t.delete(n.Children[key[0]], append(prefix, key[0]), key[1:])
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.flags = nodeFlag{dirty: true}
		n.Children[key[0]] = nn

		// Si el fullNode tiene solo un hijo, convertir a shortNode
		pos := -1
		for i := 0; i < 17; i++ {
			if n.Children[i] != nil {
				if pos == -1 {
					pos = i
				} else {
					pos = -2
					break
				}
			}
		}
		if pos >= 0 {
			if pos != 16 {
				// Convertir a shortNode
				child := n.Children[pos]
				if short, ok := child.(*shortNode); ok {
					return true, &shortNode{concat([]byte{byte(pos)}, short.Key...), short.Val, nodeFlag{dirty: true}}, nil
				}
				return true, &shortNode{[]byte{byte(pos)}, child, nodeFlag{dirty: true}}, nil
			}
		}
		return true, n, nil

	case valueNode:
		return true, nil, nil

	case nil:
		return false, nil, nil

	case hashNode:
		// Cargar nodo desde database
		rn, err := t.resolveHash(n, prefix)
		if err != nil {
			return false, nil, err
		}
		dirty, nn, err := t.delete(rn, prefix, key)
		if !dirty || err != nil {
			return false, rn, err
		}
		return true, nn, nil

	default:
		panic(fmt.Sprintf("invalid node type: %T", n))
	}
}

// Hash retorna el hash root del trie
func (t *Trie) Hash() []byte {
	hash, _ := t.hashRoot()
	return hash
}

// hashRoot calcula el hash root del trie
func (t *Trie) hashRoot() ([]byte, error) {
	if t.root == nil {
		return emptyRoot, nil
	}

	h := newHasher()
	defer returnHasher(h)

	hashed, err := h.hashRoot(t.root)
	if err != nil {
		return nil, err
	}

	if hn, ok := hashed.(hashNode); ok {
		return hn, nil
	}

	// Si el root es pequeño, calcular hash manualmente
	h.encbuf.buf = h.encbuf.buf[:0]
	if err := t.root.encode(&h.encbuf); err != nil {
		return nil, err
	}
	return h.makeHashNode(h.encbuf.buf), nil
}

// Commit escribe todos los nodos del trie a la database
func (t *Trie) Commit() ([]byte, error) {
	if t.root == nil {
		return emptyRoot, nil
	}

	// Hashear el trie
	rootHash, err := t.hashRoot()
	if err != nil {
		return nil, err
	}

	// Escribir nodos a database
	batch := t.db.db.NewBatch()
	if err := t.commitNode(batch, t.root); err != nil {
		return nil, err
	}

	if err := batch.Write(); err != nil {
		return nil, err
	}

	t.unhashed = 0
	return rootHash, nil
}

// commitNode escribe un nodo y sus hijos a la database
func (t *Trie) commitNode(batch database.Batch, n node) error {
	// Codificar nodo en RLP
	h := newHasher()
	defer returnHasher(h)

	h.encbuf.buf = h.encbuf.buf[:0]
	if err := n.encode(&h.encbuf); err != nil {
		return err
	}

	// Solo guardar nodos grandes (>= 32 bytes)
	if len(h.encbuf.buf) >= 32 {
		hash := h.makeHashNode(h.encbuf.buf)
		if err := batch.Put(hash, h.encbuf.buf); err != nil {
			return err
		}
	}

	// Procesar hijos recursivamente
	switch n := n.(type) {
	case *shortNode:
		return t.commitNode(batch, n.Val)
	case *fullNode:
		for i := 0; i < 16; i++ {
			if n.Children[i] != nil {
				if err := t.commitNode(batch, n.Children[i]); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// resolveHash carga un nodo desde la database usando su hash
func (t *Trie) resolveHash(n hashNode, prefix []byte) (node, error) {
	hash := []byte(n)
	enc, err := t.db.Node(hash)
	if err != nil {
		return nil, err
	}
	return mustDecodeNode(hash, enc), nil
}

// copy crea una copia de un fullNode
func (n *fullNode) copy() *fullNode {
	copy := *n
	return &copy
}

// concat concatena slices de bytes
func concat(s1 []byte, s2 ...byte) []byte {
	r := make([]byte, len(s1)+len(s2))
	copy(r, s1)
	copy(r[len(s1):], s2)
	return r
}
