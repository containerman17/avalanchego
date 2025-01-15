package blockstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegularBlockStore(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "blockstore_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an initializer function that returns a new BlockStore instance
	initializer := func() (IBlockPersister, func()) {
		// Create a unique file for each test
		path := filepath.Join(tmpDir, "test.db")

		// Ensure the file doesn't exist before creating the store
		os.Remove(path) // Remove any existing file

		store, err := CreateRegularBlockStore(path, 4096)
		if err != nil {
			t.Fatalf("Failed to create BlockStore: %v", err)
		}

		cleanup := func() {
			store.Close()
			os.Remove(path)
		}

		return store, cleanup
	}

	// Run the test suite
	RunTests(t, initializer)
}

func TestRegularBlockStoreErrorCases(t *testing.T) {
	// Test creating BlockStore with invalid path
	_, err := CreateRegularBlockStore("/invalid/path/test.db", 4096)
	if err == nil {
		t.Error("Expected error when creating BlockStore with invalid path")
	}

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "blockstore_error_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create a valid BlockStore
	store, err := CreateRegularBlockStore(tmpFile.Name(), 4096)
	if err != nil {
		t.Fatalf("Failed to create BlockStore: %v", err)
	}
	defer store.Close()

	// Test getting non-existent block
	_, err = store.Get(999)
	if err == nil {
		t.Error("Expected error when getting non-existent block")
	}

	// Test updating non-existent block
	data := []byte("Hello, World!")
	err = store.Update(999, data)
	if err == nil {
		t.Error("Expected error when updating non-existent block")
	}
}
