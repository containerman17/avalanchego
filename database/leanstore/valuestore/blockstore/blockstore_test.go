package blockstore

import (
	"testing"
)

func RunTests(t *testing.T, initializer func() (IBlockPersister, func())) {
	persister, cleanup := initializer()
	defer cleanup()
	testBasicOperations(t, persister)

	persister, cleanup = initializer()
	defer cleanup()
	testMultipleBlocks(t, persister)
}

// TestBasicOperations tests the basic operations of a block persister
func testBasicOperations(t *testing.T, persister IBlockPersister) {
	// Test data
	data1 := []byte("Hello, World!")

	// Test Insert
	blockID1, err := persister.Insert(data1)
	if err != nil {
		t.Fatalf("Failed to insert first block: %v", err)
	}
	if blockID1 != 0 {
		t.Errorf("Expected first block ID to be 0, got %d", blockID1)
	}

	// Test Get
	retrieved1, err := persister.Get(blockID1)
	if err != nil {
		t.Fatalf("Failed to get first block: %v", err)
	}
	if string(retrieved1[:len(data1)]) != string(data1) {
		t.Errorf("Retrieved data doesn't match: expected %q, got %q", data1, retrieved1[:len(data1)])
	}

	// Test Update
	newData := []byte("Updated content!")
	err = persister.Update(blockID1, newData)
	if err != nil {
		t.Fatalf("Failed to update block: %v", err)
	}

	// Verify update
	retrieved2, err := persister.Get(blockID1)
	if err != nil {
		t.Fatalf("Failed to get updated block: %v", err)
	}
	if string(retrieved2[:len(newData)]) != string(newData) {
		t.Errorf("Updated data doesn't match: expected %q, got %q", newData, retrieved2[:len(newData)])
	}
}

// TestMultipleBlocks tests inserting and retrieving multiple blocks
func testMultipleBlocks(t *testing.T, persister IBlockPersister) {
	// Insert multiple blocks
	for i := uint32(0); i < 3; i++ {
		data := []byte("Block content")
		blockID, err := persister.Insert(data)
		if err != nil {
			t.Fatalf("Failed to insert block %d: %v", i+1, err)
		}
		if blockID != i {
			t.Errorf("Expected block ID %d, got %d", i, blockID)
		}

		// Verify the content
		retrieved, err := persister.Get(blockID)
		if err != nil {
			t.Fatalf("Failed to get block %d: %v", blockID, err)
		}
		if string(retrieved[:len(data)]) != string(data) {
			t.Errorf("Block %d: Retrieved data doesn't match: expected %q, got %q",
				blockID, data, retrieved[:len(data)])
		}
	}

	// Verify NumBlocks
	if persister.NumBlocks() != 3 {
		t.Errorf("Expected 3 blocks, got %d", persister.NumBlocks())
	}
}
