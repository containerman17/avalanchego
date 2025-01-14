package overflow

import (
	"os"
	"testing"
)

func TestStore_BasicOperations(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "store_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Ensure the directory exists with correct permissions
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create store with default capacity
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Test simple put and get
	testData := []byte("Hello, World!")
	key, err := store.Put(testData)
	if err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	// Verify key length
	if len(key) != OverflowValueLength {
		t.Errorf("Expected key length %d, got %d", OverflowValueLength, len(key))
	}

	retrieved, err := store.Get(key)
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if string(retrieved) != string(testData) {
		t.Errorf("Expected %q, got %q", testData, retrieved)
	}
}

func TestStore_CapacityLimit(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "store_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Ensure the directory exists with correct permissions
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create store with small capacity limit (100 bytes)
	store, err := NewStoreWithCapacity(tmpDir, 100)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Put data that will force file splitting
	data1 := make([]byte, 60) // First file
	data2 := make([]byte, 60) // Should force new file

	// Fill test data
	for i := range data1 {
		data1[i] = byte(i)
	}
	for i := range data2 {
		data2[i] = byte(i + 100)
	}

	// Store first chunk
	key1, err := store.Put(data1)
	if err != nil {
		t.Fatalf("Failed to put first chunk: %v", err)
	}

	// Store second chunk (should create new file)
	key2, err := store.Put(data2)
	if err != nil {
		t.Fatalf("Failed to put second chunk: %v", err)
	}

	// Verify both chunks can be retrieved
	retrieved1, err := store.Get(key1)
	if err != nil {
		t.Fatalf("Failed to get first chunk: %v", err)
	}
	retrieved2, err := store.Get(key2)
	if err != nil {
		t.Fatalf("Failed to get second chunk: %v", err)
	}

	// Check file splitting occurred
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 2 {
		t.Error("Expected at least 2 files due to capacity limit")
	}

	// Verify data integrity
	for i := range data1 {
		if data1[i] != retrieved1[i] {
			t.Errorf("Data mismatch in first chunk at position %d", i)
		}
	}
	for i := range data2 {
		if data2[i] != retrieved2[i] {
			t.Errorf("Data mismatch in second chunk at position %d", i)
		}
	}
}

func TestStore_InvalidOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Ensure the directory exists with correct permissions
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Test getting with invalid key length
	invalidKey := make([]byte, OverflowValueLength-1) // Wrong length
	_, err = store.Get(invalidKey)
	if err == nil {
		t.Error("Expected error when getting with invalid key length")
	}

	// Test putting too large value
	largeData := make([]byte, 1<<24) // 16MB
	_, err = store.Put(largeData)
	if err == nil {
		t.Error("Expected error when putting too large value")
	}
}
