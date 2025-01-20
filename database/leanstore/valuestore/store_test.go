package valuestore

import (
	"bytes"
	"os"
	"testing"

	"github.com/ava-labs/avalanchego/database"
	"github.com/stretchr/testify/require"
)

func setupValueStore(t *testing.T) (*ValueStore, func()) {
	t.Helper()

	// Create temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "valuestore_test")
	require.NoError(t, err)

	// Create value store using the temp directory directly
	store, err := NewValueStore(tmpDir, 1024)
	require.NoError(t, err)

	cleanup := func() {
		err := store.Close()
		require.NoError(t, err)
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestValueStore_BasicOperations(t *testing.T) {
	store, cleanup := setupValueStore(t)
	defer cleanup()

	key := []byte("test-key")
	value := []byte("test-value")

	// Test Put
	err := store.Put(key, value)
	require.NoError(t, err)

	// Test Has
	exists, err := store.Has(key)
	require.NoError(t, err)
	require.True(t, exists)

	// Test Get
	retrievedValue, err := store.Get(key)
	require.NoError(t, err)
	require.True(t, bytes.Equal(value, retrievedValue))

	// Test Delete
	err = store.Delete(key)
	require.NoError(t, err)

	// Verify deletion
	exists, err = store.Has(key)
	require.NoError(t, err)
	require.False(t, exists)

	// Verify Get returns not found
	_, err = store.Get(key)
	require.ErrorIs(t, err, database.ErrNotFound)
}

func TestValueStore_MultipleEntries(t *testing.T) {
	store, cleanup := setupValueStore(t)
	defer cleanup()

	entries := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	// Insert all entries
	for k, v := range entries {
		err := store.Put([]byte(k), []byte(v))
		require.NoError(t, err)
	}

	// Verify all entries
	for k, v := range entries {
		// Check Has
		exists, err := store.Has([]byte(k))
		require.NoError(t, err)
		require.True(t, exists)

		// Check Get
		value, err := store.Get([]byte(k))
		require.NoError(t, err)
		require.Equal(t, []byte(v), value)
	}
}

func TestValueStore_NonExistentKey(t *testing.T) {
	store, cleanup := setupValueStore(t)
	defer cleanup()

	key := []byte("non-existent-key")

	// Test Has
	exists, err := store.Has(key)
	require.NoError(t, err)
	require.False(t, exists)

	// Test Get
	_, err = store.Get(key)
	require.ErrorIs(t, err, database.ErrNotFound)

	// Test Delete
	err = store.Delete(key)
	require.NoError(t, err)
}

func TestValueStore_UpdateExistingKey(t *testing.T) {
	store, cleanup := setupValueStore(t)
	defer cleanup()

	key := []byte("update-test-key")
	value1 := []byte("initial-value")
	value2 := []byte("updated-value")

	// Insert initial value
	err := store.Put(key, value1)
	require.NoError(t, err)

	// Verify initial value
	retrievedValue, err := store.Get(key)
	require.NoError(t, err)
	require.True(t, bytes.Equal(value1, retrievedValue))

	// Update value
	err = store.Put(key, value2)
	require.NoError(t, err)

	// Verify updated value
	retrievedValue, err = store.Get(key)
	require.NoError(t, err)
	require.True(t, bytes.Equal(value2, retrievedValue))
}

func TestValueStore_BlockSplitting(t *testing.T) {
	// Create store with small block size to force splits
	tmpDir, err := os.MkdirTemp("", "valuestore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Use small block size of 64 bytes to force splits
	store, err := NewValueStore(tmpDir, 64)
	require.NoError(t, err)
	defer store.Close()

	// Insert multiple entries that should cause block splits
	entries := []struct {
		key   string
		value string
	}{
		{"key1", "value1-with-some-padding"},
		{"key2", "value2-with-more-padding"},
		{"key3", "value3-with-even-more-padding"},
	}

	// Insert entries
	for _, entry := range entries {
		err := store.Put([]byte(entry.key), []byte(entry.value))
		require.NoError(t, err)
	}

	// Verify all entries are still accessible
	for _, entry := range entries {
		value, err := store.Get([]byte(entry.key))
		require.NoError(t, err)
		require.Equal(t, []byte(entry.value), value)
	}
}

func TestValueStore_Empty(t *testing.T) {
	store, cleanup := setupValueStore(t)
	defer cleanup()

	value, err := store.Get([]byte("key1"))
	require.ErrorIs(t, err, database.ErrNotFound)
	require.Nil(t, value)
}
