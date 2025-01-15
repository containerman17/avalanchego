package rbtree_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ava-labs/avalanchego/database/leanstore/backedbuffer/rbtree"
)

func TestRBTreeBasicOperations(t *testing.T) {
	tree := rbtree.NewRBTreeStore()

	// Test empty tree
	val, err := tree.Get([]byte{1})
	if err != nil || val != nil {
		t.Errorf("Expected empty string and nil error for empty tree, got %v, %v", val, err)
	}

	// Test single insert and get
	err = tree.Set([][]byte{[]byte{1}}, [][]byte{{1}})
	if err != nil {
		t.Errorf("Unexpected error on Set: %v", err)
	}

	val, err = tree.Get([]byte{1})
	if err != nil || bytes.Compare(val, []byte{1}) != 0 {
		t.Errorf("Expected 'one' and nil error, got %v, %v", val, err)
	}
}
func TestRBTreeMultipleInserts(t *testing.T) {
	tree := rbtree.NewRBTreeStore()
	keys := [][]byte{[]byte{5}, []byte{3}, []byte{7}, []byte{1}, []byte{9}}
	vals := [][]byte{{5}, {3}, {7}, {1}, {9}}

	err := tree.Set(keys, vals)
	if err != nil {
		t.Errorf("Unexpected error on Set: %v", err)
	}

	// Verify all values
	for i, k := range keys {
		val, err := tree.Get(k)
		if err != nil || bytes.Compare(val, vals[i]) != 0 {
			t.Errorf("For key %v: expected %v, got %v with error %v", k, vals[i], val, err)
		}
	}
}
func TestRBTreeGetNextPrev(t *testing.T) {
	tree := rbtree.NewRBTreeStore()

	// Insert some ordered data
	keys := [][]byte{
		[]byte{10},
		[]byte{20},
		[]byte{30},
		[]byte{40},
		[]byte{50},
	}
	values := [][]byte{
		[]byte{10},
		[]byte{20},
		[]byte{30},
		[]byte{40},
		[]byte{50},
	}
	err := tree.Set(keys, values)
	if err != nil {
		t.Errorf("Unexpected error on Set: %v", err)
	}

	// Test GetNext
	k, v, err := tree.GetNext([]byte{20})
	if err != nil || bytes.Compare(k, []byte{30}) != 0 || bytes.Compare(v, []byte{30}) != 0 {
		t.Errorf("GetNext(20) expected (30, 30), got (%v, %v)", k, v)
	}

	// Test GetNext for non-existent key
	k, v, err = tree.GetNext([]byte{25})
	if err != nil || bytes.Compare(k, []byte{30}) != 0 || bytes.Compare(v, []byte{30}) != 0 {
		t.Errorf("GetNext(25) expected (30, 30), got (%v, %v)", k, v)
	}

	// Test GetPrev
	k, v, err = tree.GetPrev([]byte{30})
	if err != nil || bytes.Compare(k, []byte{20}) != 0 || bytes.Compare(v, []byte{20}) != 0 {
		t.Errorf("GetPrev(30) expected (20, 20), got (%v, %v)", k, v)
	}

	// Test GetPrev for non-existent key
	k, v, err = tree.GetPrev([]byte{25})
	if err != nil || bytes.Compare(k, []byte{20}) != 0 || bytes.Compare(v, []byte{20}) != 0 {
		t.Errorf("GetPrev(25) expected (20, 20), got (%v, %v)", k, v)
	}
}

func TestRBTreeCornerCases(t *testing.T) {
	tree := rbtree.NewRBTreeStore()

	// Test GetNext on empty tree
	k, v, err := tree.GetNext([]byte{1})
	if err != nil || k != nil || v != nil {
		t.Errorf("GetNext on empty tree expected (nil, nil), got (%v, %v)", k, v)
	}

	// Test GetPrev on empty tree
	k, v, err = tree.GetPrev([]byte{1})
	if err != nil || k != nil || v != nil {
		t.Errorf("GetPrev on empty tree expected (nil, nil), got (%v, %v)", k, v)
	}

	// Test GetNext at the end of tree
	err = tree.Set([][]byte{[]byte{1}}, [][]byte{{1}})
	if err != nil {
		t.Errorf("Unexpected error on Set: %v", err)
	}
	k, v, err = tree.GetNext([]byte{1})
	if err != nil || k != nil || v != nil {
		t.Errorf("GetNext at end expected (nil, nil), got (%v, %v)", k, v)
	}

	// Test GetPrev at the start of tree
	k, v, err = tree.GetPrev([]byte{1})
	if err != nil || k != nil || v != nil {
		t.Errorf("GetPrev at start expected (nil, nil), got (%v, %v)", k, v)
	}
}

func TestRBTreeUpdateExisting(t *testing.T) {
	tree := rbtree.NewRBTreeStore()

	// Insert initial value
	tree.Set([][]byte{[]byte{1}}, [][]byte{[]byte("one")})

	// Update existing key
	tree.Set([][]byte{[]byte{1}}, [][]byte{[]byte("ONE")})

	// Verify update
	val, err := tree.Get([]byte{1})
	if err != nil || bytes.Compare(val, []byte("ONE")) != 0 {
		t.Errorf("Expected updated value 'ONE', got %v", val)
	}
}

func TestRBTreeBalancing(t *testing.T) {
	tree := rbtree.NewRBTreeStore()

	// Insert values in ascending order to test balancing
	for i := 0; i < 10; i++ {
		err := tree.Set([][]byte{[]byte(fmt.Sprintf("%d", i))}, [][]byte{[]byte(fmt.Sprintf("val%d", i))})
		if err != nil {
			t.Errorf("Unexpected error on Set: %v", err)
		}
	}

	// Verify all values are accessible
	for i := 0; i < 10; i++ {
		val, err := tree.Get([]byte(fmt.Sprintf("%d", i)))
		if err != nil || bytes.Compare(val, []byte(fmt.Sprintf("val%d", i))) != 0 {
			t.Errorf("Expected val%d, got %v", i, val)
		}
	}
}
