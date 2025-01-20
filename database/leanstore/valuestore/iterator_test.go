package valuestore

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T) (*ValueStore, func()) {
	dir, err := os.MkdirTemp("", "valuestore_test_*")
	require.NoError(t, err)

	store, err := NewValueStore(dir, 1024)
	require.NoError(t, err)

	cleanup := func() {
		store.Close()
		os.RemoveAll(dir)
	}

	return store, cleanup
}

func TestIterator_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	iter := store.NewIterator(nil, nil)
	defer iter.Release()

	// Empty store should return false on first Next()
	require.False(t, iter.Next())
	require.Nil(t, iter.Key())
	require.Nil(t, iter.Value())
	require.NoError(t, iter.Error())
}

func TestIterator_SingleKey(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Put a single key-value pair
	key := []byte("key1")
	value := []byte("value1")
	require.NoError(t, store.Put(key, value))

	iter := store.NewIterator(nil, nil)
	defer iter.Release()

	// Should be able to iterate once
	require.True(t, iter.Next())
	require.True(t, bytes.Equal(key, iter.Key()))
	require.True(t, bytes.Equal(value, iter.Value()))

	// No more items
	require.False(t, iter.Next())
	require.NoError(t, iter.Error())
}

func TestIterator_MultipleKeys(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Put multiple key-value pairs
	pairs := []struct {
		key   []byte
		value []byte
	}{
		{[]byte("key1"), []byte("value1")},
		{[]byte("key2"), []byte("value2")},
		{[]byte("key3"), []byte("value3")},
	}

	for _, pair := range pairs {
		require.NoError(t, store.Put(pair.key, pair.value))
	}

	iter := store.NewIterator(nil, nil)
	defer iter.Release()

	// Should iterate through all pairs in order
	for _, pair := range pairs {
		require.True(t, iter.Next())
		require.True(t, bytes.Equal(pair.key, iter.Key()))
		require.True(t, bytes.Equal(pair.value, iter.Value()))
	}

	// No more items
	require.False(t, iter.Next())
	require.NoError(t, iter.Error())
}

func TestIterator_Range(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Put multiple key-value pairs
	pairs := []struct {
		key   []byte
		value []byte
	}{
		{[]byte("key1"), []byte("value1")},
		{[]byte("key2"), []byte("value2")},
		{[]byte("key3"), []byte("value3")},
		{[]byte("key4"), []byte("value4")},
		{[]byte("key5"), []byte("value5")},
	}

	for _, pair := range pairs {
		require.NoError(t, store.Put(pair.key, pair.value))
	}

	// Test range iteration
	iter := store.NewIterator([]byte("key2"), []byte("key4"))
	defer iter.Release()

	// Should only iterate through key2 and key3
	expectedPairs := pairs[1:3] // key2 and key3
	for _, pair := range expectedPairs {
		require.True(t, iter.Next())
		require.True(t, bytes.Equal(pair.key, iter.Key()))
		require.True(t, bytes.Equal(pair.value, iter.Value()))
	}

	// No more items
	require.False(t, iter.Next())
	require.NoError(t, iter.Error())
}

func TestIterator_Release(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	require.NoError(t, store.Put([]byte("key1"), []byte("value1")))

	iter := store.NewIterator(nil, nil)

	// Should work before release
	require.True(t, iter.Next())
	require.NotNil(t, iter.Key())
	require.NotNil(t, iter.Value())

	// Release the iterator
	iter.Release()

	// Should not work after release
	require.False(t, iter.Next())
	require.Nil(t, iter.Key())
	require.Nil(t, iter.Value())
}

func TestIterator_MultipleBlocks(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Put enough key-value pairs to span multiple blocks
	pairs := make([]struct {
		key   []byte
		value []byte
	}, 100)

	for i := 0; i < 100; i++ {
		pairs[i].key = []byte(fmt.Sprintf("key%03d", i))
		pairs[i].value = []byte(fmt.Sprintf("value%03d", i))
		require.NoError(t, store.Put(pairs[i].key, pairs[i].value))
	}

	iter := store.NewIterator(nil, nil)
	defer iter.Release()

	// Should iterate through all pairs in order
	for _, pair := range pairs {
		require.True(t, iter.Next())
		require.Equal(t, string(pair.key), string(iter.Key()))
		require.Equal(t, string(pair.value), string(iter.Value()))
	}

	require.False(t, iter.Next())
	require.NoError(t, iter.Error())
}

func TestIterator_OverlappingBlocks(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Insert keys in non-sequential order to force multiple blocks
	testData := []struct {
		key   string
		value string
	}{
		{"key3", "value3"},
		{"key1", "value1"},
		{"key4", "value4"},
		{"key2", "value2"},
		{"key5", "value5"},
	}

	// Insert data
	for _, data := range testData {
		require.NoError(t, store.Put([]byte(data.key), []byte(data.value)))
	}

	// Create iterator for full range
	iter := store.NewIterator(nil, nil)
	defer iter.Release()

	// Expected order after sorting
	expected := []struct {
		key   string
		value string
	}{
		{"key1", "value1"},
		{"key2", "value2"},
		{"key3", "value3"},
		{"key4", "value4"},
		{"key5", "value5"},
	}

	// Verify iteration order
	for _, exp := range expected {
		require.True(t, iter.Next())
		require.Equal(t, exp.key, string(iter.Key()))
		require.Equal(t, exp.value, string(iter.Value()))
	}

	require.False(t, iter.Next())
	require.NoError(t, iter.Error())
}
