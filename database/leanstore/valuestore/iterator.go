package valuestore

import (
	"bytes"
	"sync"

	"github.com/ava-labs/avalanchego/database"
)

var (
	_ database.Iterator = (*Iterator)(nil)
)

type Iterator struct {
	lock sync.Mutex

	valuestore *ValueStore
	start      []byte
	end        []byte

	// Current block tracking
	currentBlockID uint32
	currentKeys    [][]byte
	currentValues  [][]byte
	currentIndex   int

	// Error tracking
	lastError error
	released  bool
}

// Error implements database.Iterator.
func (i *Iterator) Error() error {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.lastError
}

// Key implements database.Iterator.
func (i *Iterator) Key() []byte {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.released || i.currentIndex < 0 || i.currentIndex >= len(i.currentKeys) {
		return nil
	}

	return i.currentKeys[i.currentIndex]
}

// Value implements database.Iterator.
func (i *Iterator) Value() []byte {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.released || i.currentIndex < 0 || i.currentIndex >= len(i.currentValues) {
		return nil
	}

	return i.currentValues[i.currentIndex]
}

// Next implements database.Iterator.
func (i *Iterator) Next() bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.released || i.lastError != nil {
		return false
	}

	// First call to Next()
	if i.currentIndex == -1 {
		return i.loadNextBlock()
	}

	// Try moving to next item in current block
	i.currentIndex++
	if i.currentIndex < len(i.currentKeys) {
		return true
	}

	// Need to load next block
	return i.loadNextBlock()
}

// Release implements database.Iterator.
func (i *Iterator) Release() {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.clearCurrentBlock()
	i.released = true
}

// Helper methods
func (i *Iterator) loadNextBlock() bool {
	// Get the next block's starting key and ID
	var nextKey []byte
	var nextBlockID uint32
	var err error

	if i.currentKeys == nil {
		// First block - use the start key
		nextBlockID, err = i.valuestore.index.GetFloorValue(i.start)
		if err != nil {
			if err.Error() == "no floor value found" {
				// For empty database, use block 0 which was created during initialization
				nextBlockID = 0
			} else {
				i.lastError = err
				return false
			}
		}
	} else {
		// Only try to get next block if we have keys in the current block
		if len(i.currentKeys) > 0 {
			nextKey, nextBlockID, err = i.valuestore.index.GetNextKeyValue(i.currentKeys[len(i.currentKeys)-1])
			if err != nil {
				i.lastError = err
				return false
			}
			// No more blocks
			if nextKey == nil {
				return false
			}
			// If we have an end boundary and the next block starts beyond it, we're done
			if i.end != nil && bytes.Compare(nextKey, i.end) >= 0 {
				return false
			}
		} else {
			// Current block is empty, no need to continue
			return false
		}
	}

	// Clear current block data before loading new block
	i.clearCurrentBlock()
	i.currentBlockID = nextBlockID

	return i.loadCurrentBlock()
}

func (i *Iterator) loadCurrentBlock() bool {
	// Get mutex for the block
	mutex := i.valuestore.getMutex(i.currentBlockID)
	mutex.RLock()
	defer mutex.RUnlock()

	// Load the block
	block, err := i.valuestore.blockStore.Get(i.currentBlockID)
	if err != nil {
		i.lastError = err
		return false
	}

	// Decode the block
	keys, values, err := i.valuestore.codec.Decode(block)
	if err != nil {
		i.lastError = err
		return false
	}

	// Filter keys based on range
	i.currentKeys = make([][]byte, 0, len(keys))
	i.currentValues = make([][]byte, 0, len(values))

	for idx, key := range keys {
		// Skip keys before start
		if i.start != nil && bytes.Compare(key, i.start) < 0 {
			continue
		}
		// Stop if we've exceeded end
		if i.end != nil && bytes.Compare(key, i.end) >= 0 {
			break
		}
		i.currentKeys = append(i.currentKeys, key)
		i.currentValues = append(i.currentValues, values[idx])
	}

	if len(i.currentKeys) == 0 {
		return i.loadNextBlock()
	}

	i.currentIndex = 0
	return true
}

func (i *Iterator) clearCurrentBlock() {
	i.currentKeys = nil
	i.currentValues = nil
	i.currentIndex = -1
}
