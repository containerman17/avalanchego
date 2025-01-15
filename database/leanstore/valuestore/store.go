package valuestore

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leanstore/blockstore"
	"github.com/ava-labs/avalanchego/database/leanstore/indexdb"
)

const (
	numLockStripes = 10000
)

type ValueStore struct {
	index      *indexdb.IndexDB
	mutexes    [numLockStripes]sync.RWMutex
	codec      *BlockDecoder
	blockStore *blockstore.RegularBlockStore
	blockSize  int
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

	store := &ValueStore{
		index:      index,
		codec:      NewBlockDecoder(),
		mutexes:    [numLockStripes]sync.RWMutex{},
		blockStore: blockStore,
		blockSize:  blockSize,
	}

	if err := store.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize value store: %w", err)
	}

	return store, nil
}

func (v *ValueStore) Close() error {
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

// manualy puts 0x00 value
func (v *ValueStore) initialize() error {
	_, err := v.index.GetFloorValue([]byte{0x00})
	if err == nil {
		return nil
	}

	if err != nil && err.Error() == "no floor value found" {
		// Create a minimal valid block with header, empty prefix, and one empty entry
		block := []byte{
			0x00, 0x07, // Block length (7 bytes)
			0x00,       // Prefix length (0)
			0x00,       // Key length (0)
			0x00, 0x00, // Value length (0)
		}

		newBlockID, err := v.blockStore.Insert(block)
		if err != nil {
			return fmt.Errorf("failed to insert empty block: %w", err)
		}

		err = v.index.Put([][]byte{{0x00}}, []uint32{newBlockID})
		if err != nil {
			return fmt.Errorf("failed to put floor value: %w", err)
		}
	}

	return nil
}

func (v *ValueStore) getMutex(blockID uint32) *sync.RWMutex {
	return &v.mutexes[blockID%numLockStripes]
}

// Delete implements database.KeyValueReaderWriterDeleter.
func (v *ValueStore) Delete(key []byte) error {
	blockID, err := v.index.GetFloorValue(key)
	if err != nil {
		if err.Error() == "no floor value found" {
			return database.ErrNotFound
		}
		return fmt.Errorf("failed to get floor value: %w", err)
	}

	mutex := v.getMutex(blockID)
	mutex.Lock()
	defer mutex.Unlock()

	block, err := v.blockStore.Get(blockID)
	if err != nil {
		return fmt.Errorf("failed to get block: %w", err)
	}

	found, _, err := GetValue(block, key)
	if err != nil {
		return fmt.Errorf("failed to get value: %w", err)
	}
	if !found {
		return database.ErrNotFound
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
		return database.ErrNotFound
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
	blockID, err := v.index.GetFloorValue(key)
	if err != nil {
		if err.Error() == "no floor value found" {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get floor value: %w", err)
	}

	mutex := v.getMutex(blockID)
	mutex.RLock()
	defer mutex.RUnlock()

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
	blockID, err := v.index.GetFloorValue(key)
	if err != nil {
		if err.Error() == "no floor value found" {
			return false, nil
		}
		return false, fmt.Errorf("failed to get floor value: %w", err)
	}

	mutex := v.getMutex(blockID)
	mutex.RLock()
	defer mutex.RUnlock()

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
	blockID, err := v.index.GetFloorValue(key)
	if err != nil {
		return fmt.Errorf("failed to get floor value: %w", err)
	}

	mutex := v.getMutex(blockID)
	mutex.Lock()
	defer mutex.Unlock()

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

var _ database.KeyValueReaderWriterDeleter = (*ValueStore)(nil)
