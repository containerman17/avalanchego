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

	// FIXME: Implement database-level snapshots to ensure iterator consistency
	// Currently, iterators may see modifications made after their creation

	// In-memory storage of all keys/values
	keys   [][]byte
	values [][]byte
	index  int

	// Error tracking
	lastError error
	released  bool
	closed    bool
}

// Error implements database.Iterator.
func (i *Iterator) Error() error {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.closed {
		return database.ErrClosed
	}
	return i.lastError
}

// Key implements database.Iterator.
func (i *Iterator) Key() []byte {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.released || i.closed || i.index < 0 || i.index >= len(i.keys) {
		return nil
	}

	return i.keys[i.index]
}

// Value implements database.Iterator.
func (i *Iterator) Value() []byte {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.released || i.closed || i.index < 0 || i.index >= len(i.values) {
		return nil
	}

	return i.values[i.index]
}

// Next implements database.Iterator.
func (i *Iterator) Next() bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.released || i.lastError != nil {
		return false
	}

	if i.closed || i.valuestore.closed {
		i.keys = make([][]byte, 0)
		i.values = make([][]byte, 0)
		i.index = -1
		return false
	}

	// First call to Next()
	if i.index == -1 {
		return i.loadAllData()
	}

	// Move to next item
	i.index++
	return i.index < len(i.keys)
}

// Release implements database.Iterator.
func (i *Iterator) Release() {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.keys = nil
	i.values = nil
	i.released = true
}

func (it *Iterator) loadAllData() bool {
	// First collect all block numbers we need to scan
	var blockNumbers []uint32

	// Get the first block number
	floorKey, blockNumber, err := it.valuestore.index.GetFloorKeyValue(it.start)
	if err != nil {
		it.lastError = err
		return false
	}

	// Keep getting next block numbers until we exceed end range
	blockNumbers = append(blockNumbers, blockNumber)

	for {
		nextBlockFirstKey, nextBlockNumber, err := it.valuestore.index.GetNextKeyValue(floorKey)
		if err != nil {
			it.lastError = err
			return false
		}
		if nextBlockFirstKey == nil || bytes.Compare(nextBlockFirstKey, it.end) >= 0 {
			break
		}
		blockNumbers = append(blockNumbers, nextBlockNumber)
		floorKey = nextBlockFirstKey
	}

	// Now extract all keys/values from the blocks
	it.keys = make([][]byte, 0)
	it.values = make([][]byte, 0)

	for _, blockNum := range blockNumbers {
		blockBytes, err := it.valuestore.blockStore.Get(blockNum)
		if err != nil {
			it.lastError = err
			return false
		}
		keys, values, err := it.valuestore.codec.Decode(blockBytes)
		if err != nil {
			it.lastError = err
			return false
		}
		// Make copies of keys and values
		keysCopy := make([][]byte, len(keys))
		valuesCopy := make([][]byte, len(values))
		for i := range keys {
			keysCopy[i] = append([]byte(nil), keys[i]...)
			valuesCopy[i] = append([]byte(nil), values[i]...)
		}
		keys = keysCopy
		values = valuesCopy

		// Filter keys in range
		for i, key := range keys {
			if bytes.Compare(key, it.start) < 0 {
				continue
			}
			if bytes.Compare(key, it.end) >= 0 {
				continue
			}
			it.keys = append(it.keys, key)
			it.values = append(it.values, values[i])
		}
	}

	it.index = 0

	return len(it.keys) > 0
}
