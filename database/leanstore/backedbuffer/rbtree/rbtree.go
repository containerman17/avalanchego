// This implementation uses a Red-Black Tree to store []byte keys and []byte values.
// The tree stays balanced, offering O(log n) for Get and Set.
// GetNext and GetPrev find the successor or predecessor of a given key, even if that key doesn't exist.
// If the key isn't present, it treats it like it would be inserted, then finds the next or prev key relative to that position.

package rbtree

import (
	"fmt"
)

type color bool

const (
	red   color = true
	black color = false
)

type rbNode struct {
	data   []byte // Combined key and value
	keyLen uint8  // Length of the key portion

	color  color
	left   uint32
	right  uint32
	parent uint32
}

const nullNode uint32 = 0 // Index 0 represents nil

type RBTreeStore struct {
	nodes []rbNode // Pre-allocated array of nodes
	root  uint32   // Index of root node
	size  int
	free  uint32 // Index of first free node
}

func NewRBTreeStore() *RBTreeStore {
	// Pre-allocate with a reasonable initial size
	initialSize := 2
	nodes := make([]rbNode, initialSize)

	// Setup free list by linking through the 'left' field
	for i := uint32(0); i < uint32(initialSize-1); i++ {
		nodes[i].left = i + 1
	}
	nodes[initialSize-1].left = nullNode

	return &RBTreeStore{
		nodes: nodes,
		root:  nullNode,
		free:  1, // Start at 1 since 0 is nullNode
	}
}

// allocNode gets a free node from the arena
func (t *RBTreeStore) allocNode() uint32 {
	if t.free == nullNode {
		// Double the size when we run out
		oldSize := len(t.nodes)
		newSize := oldSize * 2
		newNodes := make([]rbNode, newSize)
		copy(newNodes, t.nodes)

		// Setup free list for new nodes
		for i := uint32(oldSize); i < uint32(newSize-1); i++ {
			newNodes[i].left = i + 1
		}
		newNodes[newSize-1].left = nullNode

		t.nodes = newNodes
		t.free = uint32(oldSize)
	}

	nodeIdx := t.free
	t.free = t.nodes[nodeIdx].left
	t.nodes[nodeIdx] = rbNode{} // Zero the node
	return nodeIdx
}

// freeNode returns a node to the free list
func (t *RBTreeStore) freeNode(idx uint32) {
	t.nodes[idx] = rbNode{} // Zero the node
	t.nodes[idx].left = t.free
	t.free = idx
}

func bytesLess(a []byte, aKeyLen uint8, b []byte, bKeyLen uint8) bool {
	minLen := aKeyLen
	if bKeyLen < minLen {
		minLen = bKeyLen
	}
	for i := uint8(0); i < minLen; i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return aKeyLen < bKeyLen
}

func bytesEqual(a []byte, aKeyLen uint8, b []byte, bKeyLen uint8) bool {
	if aKeyLen != bKeyLen {
		return false
	}
	for i := uint8(0); i < aKeyLen; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (t *RBTreeStore) Get(k []byte) ([]byte, error) {
	n := t.searchNode(k)
	if n == nullNode {
		return nil, nil
	}
	node := &t.nodes[n]
	return node.data[node.keyLen:], nil
}

func (t *RBTreeStore) Set(keys [][]byte, values [][]byte) error {
	if len(keys) != len(values) {
		return fmt.Errorf("keys and values length mismatch")
	}
	for i := range keys {
		t.insert(keys[i], values[i])
	}
	return nil
}

// GetNext returns the smallest key-value pair that is strictly larger than the given key k.
// If k exists in the tree, it returns the next larger key-value pair.
// If k doesn't exist, it returns the smallest key-value pair that is larger than k.
// Returns zero values if no such key exists (i.e., k is larger than or equal to the largest key in the tree).
func (t *RBTreeStore) GetNext(k []byte) ([]byte, []byte, error) {
	n := t.searchInsertPos(k)
	if n == nullNode {
		return nil, nil, nil
	}

	if bytesEqual(k, uint8(len(k)), t.nodes[n].data, t.nodes[n].keyLen) {
		succ := successor(t, n)
		if succ == nullNode {
			return nil, nil, nil
		}
		node := &t.nodes[succ]
		return node.data[:node.keyLen], node.data[node.keyLen:], nil
	}

	// For non-existent key, if the found node is greater than k,
	// that's our next value. Otherwise, we need its successor.
	if bytesLess(k, uint8(len(k)), t.nodes[n].data, t.nodes[n].keyLen) {
		node := &t.nodes[n]
		return node.data[:node.keyLen], node.data[node.keyLen:], nil
	}

	succ := successor(t, n)
	if succ == nullNode {
		return nil, nil, nil
	}
	node := &t.nodes[succ]
	return node.data[:node.keyLen], node.data[node.keyLen:], nil
}

// GetPrev returns the largest key-value pair that is strictly smaller than the given key k.
// If k exists in the tree, it returns the next smaller key-value pair.
// If k doesn't exist, it returns the largest key-value pair that is smaller than k.
// Returns zero values if no such key exists (i.e., k is smaller than or equal to the smallest key in the tree).
func (t *RBTreeStore) GetPrev(k []byte) ([]byte, []byte, error) {
	n := t.searchInsertPos(k)
	if n == nullNode {
		return nil, nil, nil
	}

	if bytesEqual(k, uint8(len(k)), t.nodes[n].data, t.nodes[n].keyLen) {
		pred := predecessor(t, n)
		if pred == nullNode {
			return nil, nil, nil
		}
		node := &t.nodes[pred]
		return node.data[:node.keyLen], node.data[node.keyLen:], nil
	}

	var candidate uint32
	if bytesLess(k, uint8(len(k)), t.nodes[n].data, t.nodes[n].keyLen) {
		candidate = n
	} else {
		candidate = predecessor(t, n)
	}

	if candidate == nullNode {
		return nil, nil, nil
	}
	if !bytesLess(t.nodes[candidate].data, t.nodes[candidate].keyLen, k, uint8(len(k))) && !bytesEqual(t.nodes[candidate].data, t.nodes[candidate].keyLen, k, uint8(len(k))) {
		pred := predecessor(t, candidate)
		if pred == nullNode {
			return nil, nil, nil
		}
		node := &t.nodes[pred]
		return node.data[:node.keyLen], node.data[node.keyLen:], nil
	}
	return t.nodes[candidate].data[:t.nodes[candidate].keyLen], t.nodes[candidate].data[t.nodes[candidate].keyLen:], nil
}

// searchNode finds the node with key k
func (t *RBTreeStore) searchNode(k []byte) uint32 {
	cur := t.root
	for cur != nullNode {
		node := &t.nodes[cur]
		if bytesLess(k, uint8(len(k)), node.data, node.keyLen) {
			cur = node.left
		} else if bytesEqual(k, uint8(len(k)), node.data, node.keyLen) {
			return cur
		} else {
			cur = node.right
		}
	}
	return nullNode
}

// searchInsertPos finds the node where k would be inserted or the node that matches k
func (t *RBTreeStore) searchInsertPos(k []byte) uint32 {
	cur := t.root
	var last uint32 = nullNode
	for cur != nullNode {
		last = cur
		node := &t.nodes[cur]
		// Compare full keys
		kLen := len(k)
		if bytesLess(k, uint8(kLen), node.data, node.keyLen) {
			cur = node.left
		} else if bytesEqual(k, uint8(kLen), node.data, node.keyLen) {
			return cur
		} else {
			cur = node.right
		}
	}
	return last
}

func successor(t *RBTreeStore, n uint32) uint32 {
	if n == nullNode {
		return nullNode
	}
	if t.nodes[n].right != nullNode {
		// go right then all the way left
		x := t.nodes[n].right
		for t.nodes[x].left != nullNode {
			x = t.nodes[x].left
		}
		return x
	}
	// go up until we come from left
	x := n
	y := t.nodes[n].parent
	for y != nullNode && x == t.nodes[y].right {
		x = y
		y = t.nodes[y].parent
	}
	return y
}

func predecessor(t *RBTreeStore, n uint32) uint32 {
	if n == nullNode {
		return nullNode
	}
	if t.nodes[n].left != nullNode {
		// go left then all the way right
		x := t.nodes[n].left
		for t.nodes[x].right != nullNode {
			x = t.nodes[x].right
		}
		return x
	}
	// go up until we come from right
	x := n
	y := t.nodes[n].parent
	for y != nullNode && x == t.nodes[y].left {
		x = y
		y = t.nodes[y].parent
	}
	return y
}

func (t *RBTreeStore) insert(k []byte, v []byte) {
	if len(k) > 255 {
		panic("key length must be less than 255")
	}
	newNodeIdx := t.allocNode()
	node := &t.nodes[newNodeIdx]

	// Combine key and value into single slice
	node.data = make([]byte, len(k)+len(v))
	copy(node.data, k)
	copy(node.data[len(k):], v)
	node.keyLen = uint8(len(k))

	node.color = red
	node.left = nullNode
	node.right = nullNode
	node.parent = nullNode

	var y uint32 = nullNode
	x := t.root
	for x != nullNode {
		y = x
		xNode := &t.nodes[x]
		if bytesLess(k, uint8(len(k)), xNode.data, xNode.keyLen) {
			x = xNode.left
		} else if bytesEqual(k, uint8(len(k)), xNode.data, xNode.keyLen) {
			// Update value for existing key
			newData := make([]byte, xNode.keyLen+uint8(len(v)))
			copy(newData, xNode.data[:xNode.keyLen])
			copy(newData[xNode.keyLen:], v)
			xNode.data = newData
			t.freeNode(newNodeIdx)
			return
		} else {
			x = xNode.right
		}
	}

	node.parent = y
	if y == nullNode {
		t.root = newNodeIdx
	} else {
		yNode := &t.nodes[y]
		if bytesLess(k, uint8(len(k)), yNode.data, yNode.keyLen) {
			yNode.left = newNodeIdx
		} else {
			yNode.right = newNodeIdx
		}
	}

	t.insertFixUp(newNodeIdx)
	t.size++
}

func (t *RBTreeStore) insertFixUp(z uint32) {
	for z != nullNode && t.nodes[z].parent != nullNode && t.nodes[t.nodes[z].parent].color == red {
		parent := t.nodes[z].parent
		grandparent := t.nodes[parent].parent
		if parent == t.nodes[grandparent].left {
			y := t.nodes[grandparent].right
			if y != nullNode && t.nodes[y].color == red {
				t.nodes[parent].color = black
				t.nodes[y].color = black
				t.nodes[grandparent].color = red
				z = grandparent
			} else {
				if z == t.nodes[parent].right {
					z = parent
					t.leftRotate(z)
				}
				t.nodes[t.nodes[z].parent].color = black
				t.nodes[t.nodes[t.nodes[z].parent].parent].color = red
				t.rightRotate(t.nodes[t.nodes[z].parent].parent)
			}
		} else {
			y := t.nodes[grandparent].left
			if y != nullNode && t.nodes[y].color == red {
				t.nodes[parent].color = black
				t.nodes[y].color = black
				t.nodes[grandparent].color = red
				z = grandparent
			} else {
				if z == t.nodes[parent].left {
					z = parent
					t.rightRotate(z)
				}
				t.nodes[t.nodes[z].parent].color = black
				t.nodes[t.nodes[t.nodes[z].parent].parent].color = red
				t.leftRotate(t.nodes[t.nodes[z].parent].parent)
			}
		}
	}
	t.nodes[t.root].color = black
}

func (t *RBTreeStore) leftRotate(x uint32) {
	y := t.nodes[x].right
	t.nodes[x].right = t.nodes[y].left
	if t.nodes[y].left != nullNode {
		t.nodes[t.nodes[y].left].parent = x
	}
	t.nodes[y].parent = t.nodes[x].parent
	if t.nodes[x].parent == nullNode {
		t.root = y
	} else if x == t.nodes[t.nodes[x].parent].left {
		t.nodes[t.nodes[x].parent].left = y
	} else {
		t.nodes[t.nodes[x].parent].right = y
	}
	t.nodes[y].left = x
	t.nodes[x].parent = y
}

func (t *RBTreeStore) rightRotate(x uint32) {
	y := t.nodes[x].left
	t.nodes[x].left = t.nodes[y].right
	if t.nodes[y].right != nullNode {
		t.nodes[t.nodes[y].right].parent = x
	}
	t.nodes[y].parent = t.nodes[x].parent
	if t.nodes[x].parent == nullNode {
		t.root = y
	} else if x == t.nodes[t.nodes[x].parent].right {
		t.nodes[t.nodes[x].parent].right = y
	} else {
		t.nodes[t.nodes[x].parent].left = y
	}
	t.nodes[y].right = x
	t.nodes[x].parent = y
}

func (t *RBTreeStore) Close() error {
	return nil
}

// AllItems returns all keys and values in the tree in sorted order
func (t *RBTreeStore) AllItems() ([][]byte, [][]byte) {
	keys := make([][]byte, 0, t.size)
	vals := make([][]byte, 0, t.size)

	var inorder func(uint32)
	inorder = func(n uint32) {
		if n == nullNode {
			return
		}
		inorder(t.nodes[n].left)
		node := &t.nodes[n]
		keys = append(keys, node.data[:node.keyLen])
		vals = append(vals, node.data[node.keyLen:])
		inorder(t.nodes[n].right)
	}

	inorder(t.root)
	return keys, vals
}
