package trie

import (
	"hash"
	"sync"

	"golang.org/x/crypto/sha3"
)

// hasher es responsable de calcular hashes de nodos
// Basado en go-ethereum/trie/hasher.go
type hasher struct {
	sha      hash.Hash  // Keccak256 hasher
	tmp      []byte     // Buffer temporal para encoding
	encbuf   encBuffer  // Buffer para RLP encoding
}

// hasherPool mantiene un pool de hashers para reutilizar
var hasherPool = sync.Pool{
	New: func() interface{} {
		return &hasher{
			sha: sha3.NewLegacyKeccak256(),
			tmp: make([]byte, 0, 550), // Tamaño típico de nodo
		}
	},
}

// newHasher obtiene un hasher del pool
func newHasher() *hasher {
	return hasherPool.Get().(*hasher)
}

// returnHasher devuelve un hasher al pool
func returnHasher(h *hasher) {
	h.tmp = h.tmp[:0]
	h.encbuf.buf = h.encbuf.buf[:0]
	hasherPool.Put(h)
}

// hash calcula el hash de un nodo
// Si el nodo codificado es < 32 bytes, retorna el nodo directamente (embedded)
// Si es >= 32 bytes, retorna el hash del nodo
func (h *hasher) hash(n node, force bool) (node, error) {
	// Si el nodo ya tiene hash, retornarlo
	if hash, cached := n.cache(); cached {
		return hash, nil
	}

	// Codificar el nodo en RLP
	h.encbuf.buf = h.encbuf.buf[:0]
	if err := n.encode(&h.encbuf); err != nil {
		return nil, err
	}

	// Si el nodo es pequeño (< 32 bytes), se embebe directamente
	// Si es >= 32 bytes, se reemplaza con su hash
	if len(h.encbuf.buf) < 32 && !force {
		// Retornar el nodo sin cambios (se embebe en el padre)
		return n, nil
	}

	// Calcular hash Keccak256
	hash := h.makeHashNode(h.encbuf.buf)
	return hash, nil
}

// makeHashNode calcula Keccak256 de los datos
func (h *hasher) makeHashNode(data []byte) hashNode {
	h.sha.Reset()
	h.sha.Write(data)
	hash := h.sha.Sum(h.tmp[:0])
	return hashNode(hash)
}

// hashChildren procesa recursivamente los hijos de un fullNode
func (h *hasher) hashChildren(n *fullNode) (*fullNode, error) {
	// Crear una copia para no modificar el original
	collapsed := *n

	for i := 0; i < 16; i++ {
		if n.Children[i] != nil {
			// Hashear hijo recursivamente
			child, err := h.hash(n.Children[i], false)
			if err != nil {
				return nil, err
			}
			collapsed.Children[i] = child
		}
	}

	// Hashear el valor si existe
	if n.Children[16] != nil {
		child, err := h.hash(n.Children[16], false)
		if err != nil {
			return nil, err
		}
		collapsed.Children[16] = child
	}

	return &collapsed, nil
}

// hashShortNodeChildren procesa los hijos de un shortNode
func (h *hasher) hashShortNodeChildren(n *shortNode) (*shortNode, error) {
	// Hashear el valor
	val, err := h.hash(n.Val, false)
	if err != nil {
		return nil, err
	}

	// Crear nueva shortNode con valor hasheado
	collapsed := &shortNode{
		Key: n.Key,
		Val: val,
	}

	return collapsed, nil
}

// hashRoot calcula el hash root de todo el trie
// Procesa el nodo y todos sus hijos recursivamente
func (h *hasher) hashRoot(n node) (node, error) {
	if n == nil {
		// Trie vacío
		return emptyRoot, nil
	}

	return h.hash(n, true) // force=true para siempre calcular hash del root
}

// Keccak256 calcula el hash Keccak256 de data
func Keccak256(data ...[]byte) []byte {
	h := sha3.NewLegacyKeccak256()
	for _, b := range data {
		h.Write(b)
	}
	return h.Sum(nil)
}

// Keccak256Hash calcula el hash y retorna como hashNode
func Keccak256Hash(data ...[]byte) hashNode {
	return hashNode(Keccak256(data...))
}
