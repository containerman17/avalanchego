package valuestore

import (
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

	// Test exact matches
	for i, key := range keys {
		found, val, err := GetValue(updatedBlock, key)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, values[i], val)
	}

	// Test non-existent keys
	nonExistentKeys := [][]byte{
		[]byte("aaaaaaaaaaaaa_baaaaaa"),
		[]byte("aaaaaaaaaaaaa_zzzzz"),
		[]byte("aaaaaaaaaaaaa_aaaaa"),
	}

	for _, key := range nonExistentKeys {
		found, val, err := GetValue(updatedBlock, key)
		require.NoError(t, err)
		require.False(t, found)
		require.Nil(t, val)
	}
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
		found, val, err := GetValue(updatedBlock, key)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, values[i], val)
	}

	// Test non-existent keys
	tests := []struct {
		searchKey []byte
		wantFound bool
	}{
		{[]byte("ant"), false},     // Before first key
		{[]byte("apricot"), false}, // Between apple and banana
		{[]byte("cat"), false},     // Between banana and cherry
		{[]byte("zebra"), false},   // After last key
	}

	for _, tt := range tests {
		found, val, err := GetValue(updatedBlock, tt.searchKey)
		require.NoError(t, err)
		require.False(t, found)
		require.Nil(t, val)
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
		{[]byte("apple"), true, []byte{1, 0, 0, 0}},  // First key
		{[]byte("banana"), true, []byte{2, 0, 0, 0}}, // Second key
		{[]byte("ant"), false, nil},                  // Before first key
		{[]byte("between"), false, nil},              // Non-existent key
	}

	for _, tt := range firstBlockTests {
		found, val, err := GetValue(updatedBlock, tt.searchKey)
		require.NoError(t, err, "First block - GetValue failed for key %s", tt.searchKey)
		require.Equal(t, tt.wantFound, found, "First block - For key %s", tt.searchKey)
		if found {
			require.Equal(t, tt.wantValue, val, "First block - For key %s", tt.searchKey)
		} else {
			require.Nil(t, val)
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
		{[]byte("dog"), false, nil},                  // Non-existent key
		{[]byte("zebra"), false, nil},                // After last key
	}

	for _, tt := range secondBlockTests {
		found, val, err := GetValue(newBlocks[0].Block, tt.searchKey)
		require.NoError(t, err, "Second block - GetValue failed for key %s", tt.searchKey)
		require.Equal(t, tt.wantFound, found, "Second block - For key %s", tt.searchKey)
		if found {
			require.Equal(t, tt.wantValue, val, "Second block - For key %s", tt.searchKey)
		} else {
			require.Nil(t, val)
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
		{[]byte("zebra"), false, nil},                 // After last key
	}

	for _, tt := range tests {
		found, val, err := GetValue(updatedBlock, tt.searchKey)
		require.NoError(t, err)
		require.Equal(t, tt.wantFound, found, "for key: %s", tt.searchKey)
		if found {
			require.Equal(t, tt.wantValue, val, "for key: %s", tt.searchKey)
		} else {
			require.Nil(t, val)
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
	found, val, err := GetValue(updatedBlock, keys[0])
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, values[0], val)

	// Test last block
	lastBlock := newBlocks[len(newBlocks)-1].Block
	found, val, err = GetValue(lastBlock, keyGenerator(keyCount-1))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []byte{byte(keyCount - 1), 0, 0, 0}, val)
}

func BenchmarkBlockDecoderParallel(b *testing.B) {
	// Create 1000 different blocks
	numBlocks := 1000
	blocks := make([][]byte, numBlocks)
	numEntriesPerBlock := 99

	for blockIdx := 0; blockIdx < numBlocks; blockIdx++ {
		// Create different shared prefix for each block
		sharedPrefix := make([]byte, 64)
		for i := range sharedPrefix {
			sharedPrefix[i] = byte((blockIdx * i) % 256)
		}

		keys := make([][]byte, numEntriesPerBlock)
		values := make([][]byte, numEntriesPerBlock)

		for i := 0; i < numEntriesPerBlock; i++ {
			// Create key with shared prefix + random suffix
			key := make([]byte, 96)
			copy(key, sharedPrefix)
			for j := 64; j < 96; j++ {
				key[j] = byte((blockIdx * i * j) % 256)
			}
			keys[i] = key

			// Create value
			value := make([]byte, 128)
			for j := range value {
				value[j] = byte((blockIdx * i * j) % 256)
			}
			values[i] = value
		}

		block, newBlocks, err := EncodeBlock(nil, keys, values, 16384)
		if err != nil || len(newBlocks) > 0 {
			b.Fatalf("Failed to create test block %d: err=%v newBlocks=%d", blockIdx, err, len(newBlocks))
		}
		blocks[blockIdx] = block
	}

	b.Run("Decode", func(b *testing.B) {
		b.ResetTimer()
		b.SetBytes(int64(len(blocks[0])))
		b.RunParallel(func(pb *testing.PB) {
			decoder := NewBlockDecoder()
			blockIdx := 0
			for pb.Next() {
				block := blocks[blockIdx%numBlocks]
				decodedKeys, decodedValues, err := decoder.Decode(block)
				if err != nil {
					b.Fatal(err)
				}
				if len(decodedKeys) != numEntriesPerBlock || len(decodedValues) != numEntriesPerBlock {
					b.Fatalf("Incorrect number of entries decoded: got %d, want %d", len(decodedKeys), numEntriesPerBlock)
				}
				blockIdx++
			}
		})
	})
}
