package valuestore

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/ava-labs/avalanchego/database/leanstore/blockstore"
	"github.com/ava-labs/avalanchego/database/leanstore/indexdb"

	"github.com/ava-labs/avalanchego/database"
)

const (
	numLockStripes = 10000
)

type ValueStore struct {
	index      *indexdb.IndexDB
	mutex      sync.RWMutex // Single mutex for all operations
	codec      *BlockDecoder
	blockStore *blockstore.RegularBlockStore
	blockSize  int
	closed     bool
}

func NewValueStore(dir string, blockSize int) (*ValueStore, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	blockStore, err := blockstore.CreateRegularBlockStore(dir+"/blocks.db", blockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create block store: %w", err)
	}

	indexPath := filepath.Join(dir, "index.db")
	index, err := indexdb.NewIndexDB(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create index db: %w", err)
	}

	decoder := NewBlockDecoder()

	if blockStore.NumBlocks() == 0 {
		emptyBlock, _, err := EncodeBlock(nil, [][]byte{}, [][]byte{}, blockSize)
		if err != nil {
			return nil, fmt.Errorf("failed to encode empty block: %w", err)
		}
		blockId, err := blockStore.Insert(emptyBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to insert empty block: %w", err)
		}
		if blockId != 0 {
			panic("expected block ID 0, got " + strconv.Itoa(int(blockId)))
		}
	}

	store := &ValueStore{
		index:      index,
		codec:      decoder,
		blockStore: blockStore,
		blockSize:  blockSize,
	}

	return store, nil
}

func (v *ValueStore) Close() error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.closed {
		return database.ErrClosed
	}

	v.closed = true
	err := v.index.Close()
	if err != nil {
		return fmt.Errorf("failed to close index db: %w", err)
	}

	err = v.blockStore.Close()
	if err != nil {
		return fmt.Errorf("failed to close block store: %w", err)
	}

	return nil
}

func (v *ValueStore) Delete(key []byte) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.closed {
		return database.ErrClosed
	}

	_, blockID, err := v.index.GetFloorKeyValue(key)
	if err != nil {
		if err.Error() == "no floor value found" {
			return nil
		}
		return fmt.Errorf("failed to get floor value: %w", err)
	}

	block, err := v.blockStore.Get(blockID)
	if err != nil {
		return fmt.Errorf("failed to get block: %w", err)
	}

	found, _, err := GetValue(block, key)
	if err != nil {
		return fmt.Errorf("failed to get value: %w", err)
	}
	if !found {
		return nil
	}

	decodedKeys, decodedValues, err := v.codec.Decode(block)
	if err != nil {
		return fmt.Errorf("failed to decode block: %w", err)
	}

	// Find and remove the key-value pair
	found = false
	newKeys := make([][]byte, 0, len(decodedKeys))
	newValues := make([][]byte, 0, len(decodedValues))

	for i, k := range decodedKeys {
		if bytes.Equal(k, key) {
			found = true
			continue
		}
		newKeys = append(newKeys, k)
		newValues = append(newValues, decodedValues[i])
	}

	if !found {
		return nil
	}

	updatedBlock, newBlocks, err := EncodeBlock(nil, newKeys, newValues, v.blockSize)
	if err != nil {
		return fmt.Errorf("failed to encode block: %w", err)
	}

	if len(newBlocks) > 0 {
		panic("Implementation error: deleting a key cannot cause a split in the block")
	}

	if err := v.blockStore.Update(blockID, updatedBlock); err != nil {
		return fmt.Errorf("failed to put block: %w", err)
	}

	return nil
}

func (v *ValueStore) Get(key []byte) ([]byte, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	if v.closed {
		return nil, database.ErrClosed
	}

	_, blockID, err := v.index.GetFloorKeyValue(key)
	if err != nil {
		if err.Error() == "no floor value found" {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get floor value: %w", err)
	}

	block, err := v.blockStore.Get(blockID)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	found, value, err := GetValue(block, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get value: %w", err)
	}
	if !found {
		return nil, database.ErrNotFound
	}

	return value, nil
}

func (v *ValueStore) Has(key []byte) (bool, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	if v.closed {
		return false, database.ErrClosed
	}

	_, blockID, err := v.index.GetFloorKeyValue(key)
	if err != nil {
		if err.Error() == "no floor value found" {
			return false, nil
		}
		return false, fmt.Errorf("failed to get floor value: %w", err)
	}

	block, err := v.blockStore.Get(blockID)
	if err != nil {
		return false, fmt.Errorf("failed to get block: %w", err)
	}

	found, _, err := GetValue(block, key)
	if err != nil {
		return false, fmt.Errorf("failed to get value: %w", err)
	}

	return found, nil
}

func (v *ValueStore) Put(key []byte, value []byte) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.closed {
		return database.ErrClosed
	}

	_, blockID, err := v.index.GetFloorKeyValue(key)
	if err != nil {
		return fmt.Errorf("failed to get floor value: %w", err)
	}

	block, err := v.blockStore.Get(blockID)
	if err != nil {
		return fmt.Errorf("failed to get block: %w", err)
	}

	updatedBlock, newBlocks, err := EncodeBlock(block, [][]byte{key}, [][]byte{value}, v.blockSize)
	if err != nil {
		return fmt.Errorf("failed to encode block: %w", err)
	}

	// Update the original block
	if err := v.blockStore.Update(blockID, updatedBlock); err != nil {
		return fmt.Errorf("failed to update block: %w", err)
	}

	// Handle any new blocks created from splitting
	for _, newBlock := range newBlocks {
		newBlockID, err := v.blockStore.Insert(newBlock.Block)
		if err != nil {
			return fmt.Errorf("failed to insert new block: %w", err)
		}

		if err := v.index.Put([][]byte{newBlock.StartingKey}, []uint32{newBlockID}); err != nil {
			return fmt.Errorf("failed to update index: %w", err)
		}
	}

	return nil
}

// NewIterator creates an iterator over the key-value pairs in the store.
// If prefix is not nil, only returns key-value pairs where key starts with prefix.
// If start is not nil, only returns key-value pairs where key >= start.
// If both prefix and start are provided, start must be >= prefix.
func (v *ValueStore) NewIterator(start, prefix []byte) database.Iterator {
	if v.closed {
		return &Iterator{valuestore: v, lastError: database.ErrClosed, closed: true, index: -1}
	}

	v.mutex.RLock() // Take read lock while loading snapshot
	defer v.mutex.RUnlock()

	// Take a snapshot of the current state by loading all data now
	var blockNumbers []uint32
	var keys [][]byte
	var values [][]byte

	// Get the first block number
	floorKey, blockNumber, err := v.index.GetFloorKeyValue(start)
	if err != nil && err.Error() != "no floor value found" {
		return &Iterator{valuestore: v, lastError: err, index: -1}
	}

	// Keep getting next block numbers until we exceed end range
	if err == nil {
		blockNumbers = append(blockNumbers, blockNumber)
		for {
			nextBlockFirstKey, nextBlockNumber, err := v.index.GetNextKeyValue(floorKey)
			if err != nil {
				return &Iterator{valuestore: v, lastError: err, index: -1}
			}
			if nextBlockFirstKey == nil {
				break
			}
			blockNumbers = append(blockNumbers, nextBlockNumber)
			floorKey = nextBlockFirstKey
		}
	}

	// Now extract all keys/values from the blocks
	for _, blockNum := range blockNumbers {
		blockBytes, err := v.blockStore.Get(blockNum)
		if err != nil {
			return &Iterator{valuestore: v, lastError: err, index: -1}
		}
		blockKeys, blockValues, err := v.codec.Decode(blockBytes)
		if err != nil {
			return &Iterator{valuestore: v, lastError: err, index: -1}
		}

		// Make copies of keys and values
		for i := range blockKeys {
			keys = append(keys, append([]byte(nil), blockKeys[i]...))
			values = append(values, append([]byte(nil), blockValues[i]...))
		}
	}

	// Filter by prefix if needed
	if prefix != nil {
		var filteredKeys [][]byte
		var filteredValues [][]byte
		for i, key := range keys {
			if bytes.HasPrefix(key, prefix) {
				filteredKeys = append(filteredKeys, key)
				filteredValues = append(filteredValues, values[i])
			}
		}
		keys = filteredKeys
		values = filteredValues
	}

	// Sort by key
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if bytes.Compare(keys[i], keys[j]) > 0 {
				keys[i], keys[j] = keys[j], keys[i]
				values[i], values[j] = values[j], values[i]
			}
		}
	}

	return &Iterator{
		valuestore: v,
		keys:       keys,
		values:     values,
		index:      -1,
	}
}

var _ database.KeyValueReaderWriterDeleter = (*ValueStore)(nil)
