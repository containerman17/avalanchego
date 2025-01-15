package backedbuffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValueRewrites(t *testing.T) {
	// Create a new BackedBuffer instance
	buffer, err := New("test_wal_path")
	require.NoError(t, err)
	defer buffer.Close()

	key := []byte("testKey")
	initialValue := []byte("initialValue")
	updatedValue := []byte("updatedValue")

	// Put the initial value
	require.NoError(t, buffer.Put([][]byte{key}, [][]byte{initialValue}))

	// Verify the initial value is stored
	value, err := buffer.Get(key)
	require.NoError(t, err)
	require.Equal(t, initialValue, value)

	// Update the value
	require.NoError(t, buffer.Put([][]byte{key}, [][]byte{updatedValue}))

	// Verify the updated value is stored
	value, err = buffer.Get(key)
	require.NoError(t, err)
	require.Equal(t, updatedValue, value)
}
