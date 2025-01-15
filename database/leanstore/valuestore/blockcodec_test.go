package valuestore

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeysWithPrefix(t *testing.T) {
	keys := [][]byte{
		[]byte("aaaaaaaaaaaaa_apple"),
		[]byte("aaaaaaaaaaaaa_banana"),
		[]byte("aaaaaaaaaaaaa_cherry"),
		[]byte("aaaaaaaaaaaaa_date"),
	}

	values := [][]byte{
		[]byte{1, 0, 0, 0},
		[]byte{2, 0, 0, 0},
		[]byte{3, 0, 0, 0},
		[]byte{4, 0, 0, 0},
	}

	updatedBlock, newBlocks, err := EncodeBlock([]byte{}, keys, values, 40000)
	require.NoError(t, err)
	require.Empty(t, newBlocks)

	found, val, err := FindFloorValue(updatedBlock, []byte("aaaaaaaaaaaaa_baaaaaa"))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []byte{1, 0, 0, 0}, val)

	found, val, err = FindFloorValue(updatedBlock, []byte("aaaaaaaaaaaaa_zzzzz"))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []byte{4, 0, 0, 0}, val)

	found, val, err = FindFloorValue(updatedBlock, []byte("aaaaaaaaaaaaa_aaaaa"))
	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, val)
}

func TestCodecSingleBlock(t *testing.T) {
	keys := [][]byte{
		[]byte("apple"),
		[]byte("banana"),
		[]byte("cherry"),
		[]byte("date"),
	}
	values := [][]byte{
		[]byte{1, 0, 0, 0},
		[]byte{2, 0, 0, 0},
		[]byte{3, 0, 0, 0},
		[]byte{4, 0, 0, 0},
	}

	updatedBlock, newBlocks, err := EncodeBlock([]byte{}, keys, values, 40000)
	require.NoError(t, err)
	require.Empty(t, newBlocks)

	// Test exact matches
	for i, key := range keys {
		found, val, err := FindFloorValue(updatedBlock, key)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, values[i], val)
	}

	// Test in-between values
	tests := []struct {
		searchKey []byte
		wantFound bool
		wantValue []byte
	}{
		{[]byte("ant"), false, nil},                   // Before first key
		{[]byte("apricot"), true, []byte{1, 0, 0, 0}}, // Between apple and banana
		{[]byte("cat"), true, []byte{2, 0, 0, 0}},     // Between banana and cherry
		{[]byte("zebra"), true, []byte{4, 0, 0, 0}},   // After last key
	}

	for _, tt := range tests {
		found, val, err := FindFloorValue(updatedBlock, tt.searchKey)
		require.NoError(t, err)
		require.Equalf(t, tt.wantFound, found, "searchKey: %s", tt.searchKey)
		if found {
			require.Equalf(t, tt.wantValue, val, "searchKey: %s", tt.searchKey)
		}
	}
}

func TestCodecTwoBlocks(t *testing.T) {
	minBlockSize := 35

	keys := [][]byte{
		[]byte("apple"),
		[]byte("banana"),
		[]byte("cherry"),
		[]byte("date"),
	}
	values := [][]byte{
		{1, 0, 0, 0},
		{2, 0, 0, 0},
		{3, 0, 0, 0},
		{4, 0, 0, 0},
	}

	updatedBlock, newBlocks, err := EncodeBlock([]byte{}, keys, values, minBlockSize)
	require.NoError(t, err)
	require.Equal(t, 1, len(newBlocks))

	// Test the original block
	firstBlockTests := []struct {
		searchKey []byte
		wantFound bool
		wantValue []byte
	}{
		{[]byte("ant"), false, nil},                   // Before first key
		{[]byte("apple"), true, []byte{1, 0, 0, 0}},   // First key
		{[]byte("banana"), true, []byte{2, 0, 0, 0}},  // Second key
		{[]byte("between"), true, []byte{2, 0, 0, 0}}, // After second key
	}

	for _, tt := range firstBlockTests {
		found, val, err := FindFloorValue(updatedBlock, tt.searchKey)
		require.NoError(t, err, "First block - FindFloorValue failed for key %s", tt.searchKey)
		require.Equal(t, tt.wantFound, found, "First block - For key %s", tt.searchKey)
		if found {
			require.Equal(t, tt.wantValue, val, "First block - For key %s", tt.searchKey)
		}
	}

	// Test second block
	secondBlockTests := []struct {
		searchKey []byte
		wantFound bool
		wantValue []byte
	}{
		{[]byte("cherry"), true, []byte{3, 0, 0, 0}}, // First key in second block
		{[]byte("date"), true, []byte{4, 0, 0, 0}},   // Last key
		{[]byte("zebra"), true, []byte{4, 0, 0, 0}},  // After last key
	}

	for _, tt := range secondBlockTests {
		found, val, err := FindFloorValue(newBlocks[0].Block, tt.searchKey)
		require.NoError(t, err, "Second block - FindFloorValue failed for key %s", tt.searchKey)
		require.Equal(t, tt.wantFound, found, "Second block - For key %s", tt.searchKey)
		if found {
			require.Equal(t, tt.wantValue, val, "Second block - For key %s", tt.searchKey)
		}
	}
}

func TestCodecInsertMerge(t *testing.T) {
	// Initial data
	initialKeys := [][]byte{
		[]byte("apple"),
		[]byte("cherry"),
		[]byte("fig"),
	}
	initialValues := [][]byte{
		[]byte{1, 0, 0, 0},
		[]byte{3, 0, 0, 0},
		[]byte{6, 0, 0, 0},
	}

	originalBlock, newBlocks, err := EncodeBlock([]byte{}, initialKeys, initialValues, 40000)
	require.NoError(t, err)
	require.Empty(t, newBlocks)

	// New data to merge
	newKeys := [][]byte{
		[]byte("apple"),  // Update existing value
		[]byte("banana"), // Insert between existing
		[]byte("cherry"), // Update existing value
		[]byte("date"),   // Insert between existing
		[]byte("fig"),    // Update existing value
		[]byte("grape"),  // Insert after existing
	}
	newValues := [][]byte{
		[]byte{10, 0, 0, 0},
		[]byte{20, 0, 0, 0},
		[]byte{30, 0, 0, 0},
		[]byte{40, 0, 0, 0},
		[]byte{60, 0, 0, 0},
		[]byte{70, 0, 0, 0},
	}

	// Merge new data into existing block
	updatedBlock, newBlocks, err := EncodeBlock(originalBlock, newKeys, newValues, 40000)
	require.NoError(t, err)
	require.Empty(t, newBlocks)

	// Test all keys after merge
	tests := []struct {
		searchKey []byte
		wantFound bool
		wantValue []byte
	}{
		{[]byte("apple"), true, []byte{10, 0, 0, 0}},  // Updated value
		{[]byte("banana"), true, []byte{20, 0, 0, 0}}, // New key
		{[]byte("cherry"), true, []byte{30, 0, 0, 0}}, // Updated value
		{[]byte("date"), true, []byte{40, 0, 0, 0}},   // New key
		{[]byte("fig"), true, []byte{60, 0, 0, 0}},    // Updated value
		{[]byte("grape"), true, []byte{70, 0, 0, 0}},  // New key
		{[]byte("ant"), false, nil},                   // Before first key
		{[]byte("zebra"), true, []byte{70, 0, 0, 0}},  // After last key
	}

	for _, tt := range tests {
		found, val, err := FindFloorValue(updatedBlock, tt.searchKey)
		require.NoError(t, err)
		require.Equal(t, tt.wantFound, found, "for key: %s", tt.searchKey)
		if found {
			require.Equal(t, tt.wantValue, val, "for key: %s", tt.searchKey)
		}
	}
}

func TestManyKeys(t *testing.T) {
	keyGenerator := func(i int) []byte {
		return []byte(fmt.Sprintf("prefix_%05d_key_%05d_suffix_to_make_this_even_longer_and_force_splits", i, i))
	}

	const keyCount = 12

	keys := make([][]byte, keyCount)
	values := make([][]byte, keyCount)
	for i := 0; i < keyCount; i++ {
		keys[i] = keyGenerator(i)
		values[i] = []byte{byte(i), 0, 0, 0}
	}

	updatedBlock, newBlocks, err := EncodeBlock(nil, keys, values, 200)
	require.NoError(t, err)
	require.NotEmpty(t, newBlocks)

	allBlocks := [][]byte{updatedBlock}
	for _, block := range newBlocks {
		allBlocks = append(allBlocks, block.Block)
	}

	// Test first block
	found, val, err := FindFloorValue(updatedBlock, keys[0])
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, values[0], val)

	// Test last block
	lastBlock := newBlocks[len(newBlocks)-1].Block
	found, val, err = FindFloorValue(lastBlock, keyGenerator(keyCount-1))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []byte{byte(keyCount - 1), 0, 0, 0}, val)
}

// decodeBlock is a test helper that decodes all key-value pairs from a block
func decodeBlock(block []byte) ([][]byte, [][]byte, error) {
	if len(block) < 2 {
		return nil, nil, errors.New("block too short")
	}

	blockUsedLen := int(block[0])<<8 | int(block[1])
	if blockUsedLen > len(block) {
		return nil, nil, errors.New("invalid block length")
	}

	// Pre-count entries
	entries := 0
	pos := 2
	for pos < blockUsedLen {
		pos++ // skip shared
		keyLen := int(block[pos])
		pos++
		valueLen := int(block[pos])<<8 | int(block[pos+1])
		pos += 2 + keyLen + valueLen
		entries++
	}

	// Pre-allocate slices
	keys := make([][]byte, 0, entries)
	values := make([][]byte, 0, entries)
	keyBuf := make([]byte, 96) // We know max key size is 96
	pos = 2
	lastKey := keyBuf[:0]

	for pos < blockUsedLen {
		shared := int(block[pos])
		pos++
		keyLen := int(block[pos])
		pos++

		key := keyBuf[:shared]
		copy(key, lastKey[:shared])
		key = append(key, block[pos:pos+keyLen]...)
		pos += keyLen

		valueLen := int(block[pos])<<8 | int(block[pos+1])
		pos += 2
		value := block[pos : pos+valueLen]
		pos += valueLen

		keyCopy := make([]byte, len(key))
		copy(keyCopy, key)
		keys = append(keys, keyCopy)
		values = append(values, value)

		lastKey = key
	}

	return keys, values, nil
}

func BenchmarkDecode16KBBlock(b *testing.B) {
	// Create shared prefix (64 bytes)
	sharedPrefix := make([]byte, 64)
	for i := range sharedPrefix {
		sharedPrefix[i] = byte(i % 256)
	}

	// Create test data - calculate how many entries fit in 16KB
	keySize := 96 // 64 bytes shared + 32 bytes random
	valueSize := 128

	numEntries := 99

	// Generate keys and values
	keys := make([][]byte, numEntries)
	values := make([][]byte, numEntries)

	for i := 0; i < numEntries; i++ {
		// Create key with shared prefix + random suffix
		key := make([]byte, keySize)
		copy(key, sharedPrefix)
		for j := 64; j < keySize; j++ {
			key[j] = byte(i * j % 256) // Deterministic but varying
		}
		keys[i] = key

		// Create value
		value := make([]byte, valueSize)
		for j := range value {
			value[j] = byte(i * j % 256)
		}
		values[i] = value
	}

	// Create the block once before benchmarking
	block, newBlocks, err := EncodeBlock(nil, keys, values, 16384)
	if err != nil || len(newBlocks) > 0 {
		b.Fatalf("Failed to create test block: err=%v newBlocks=%d", err, len(newBlocks))
	}

	b.Run("Decode", func(b *testing.B) {
		b.ResetTimer()
		b.SetBytes(int64(len(block)))
		for i := 0; i < b.N; i++ {
			decodedKeys, decodedValues, err := decodeBlock(block)
			if err != nil {
				b.Fatal(err)
			}
			if len(decodedKeys) != numEntries || len(decodedValues) != numEntries {
				b.Fatalf("Incorrect number of entries decoded: got %d, want %d", len(decodedKeys), numEntries)
			}
		}
	})
}

func BenchmarkDecode16KBBlockParallel(b *testing.B) {
	// Same setup as BenchmarkDecode16KBBlock
	sharedPrefix := make([]byte, 64)
	for i := range sharedPrefix {
		sharedPrefix[i] = byte(i % 256)
	}

	keySize := 96
	valueSize := 128
	numEntries := 99

	keys := make([][]byte, numEntries)
	values := make([][]byte, numEntries)

	for i := 0; i < numEntries; i++ {
		key := make([]byte, keySize)
		copy(key, sharedPrefix)
		for j := 64; j < keySize; j++ {
			key[j] = byte(i * j % 256)
		}
		keys[i] = key

		value := make([]byte, valueSize)
		for j := range value {
			value[j] = byte(i * j % 256)
		}
		values[i] = value
	}

	block, newBlocks, err := EncodeBlock(nil, keys, values, 16384)
	if err != nil || len(newBlocks) > 0 {
		b.Fatalf("Failed to create test block: err=%v newBlocks=%d", err, len(newBlocks))
	}

	b.Run("Parallel", func(b *testing.B) {
		b.ResetTimer()
		b.SetBytes(int64(len(block)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				decodedKeys, decodedValues, err := decodeBlock(block)
				if err != nil {
					b.Fatal(err)
				}
				if len(decodedKeys) != numEntries || len(decodedValues) != numEntries {
					b.Fatalf("Incorrect number of entries decoded: got %d, want %d", len(decodedKeys), numEntries)
				}
			}
		})
	})
}
