package blockstore

import (
	"fmt"
	"os"
	"sync"
)

const (
	// Number of stripes for lock striping
	numStripes = 100
)

// BlockStore saves blocks to a file using regular file I/O operations
type RegularBlockStore struct {
	file      *os.File
	blockSize int
	// Use fixed-size array instead of slice
	mutexes     [numStripes]sync.RWMutex
	nextBlockID uint32
	// Add a mutex to protect nextBlockID at the instance level
	nextBlockIDMutex sync.RWMutex
	path             string
}

// Helper method to get the appropriate mutex for a block ID
func (bs *RegularBlockStore) getMutex(blockID uint32) *sync.RWMutex {
	return &bs.mutexes[blockID%numStripes]
}

// NewBlockStore creates a new BlockStore
// path is the path to the file to save the blocks to
// blockSizeBytes is the size in bytes
func CreateRegularBlockStore(path string, blockSizeBytes int) (*RegularBlockStore, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Calculate nextBlockID based on current file size
	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := fileInfo.Size()
	nextBlockID := uint32(fileSize / int64(blockSizeBytes))

	return &RegularBlockStore{
		file:             file,
		blockSize:        blockSizeBytes,
		path:             path,
		nextBlockID:      nextBlockID,
		mutexes:          [numStripes]sync.RWMutex{},
		nextBlockIDMutex: sync.RWMutex{},
	}, nil
}

// Get retrieves a block by its ID
func (bs *RegularBlockStore) Get(blockID uint32) ([]byte, error) {
	// Lock only the stripe for this block
	mutex := bs.getMutex(blockID)
	mutex.RLock()
	defer mutex.RUnlock()

	if blockID >= bs.nextBlockID {
		return nil, fmt.Errorf("invalid block ID: %d", blockID)
	}

	offset := int64(blockID) * int64(bs.blockSize)
	data := make([]byte, bs.blockSize)
	_, err := bs.file.ReadAt(data, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read block: %w", err)
	}

	return data, nil
}

// Update modifies an existing block with new data
func (bs *RegularBlockStore) Update(blockID uint32, data []byte) error {
	// Lock only the stripe for this block
	mutex := bs.getMutex(blockID)
	mutex.Lock()
	defer mutex.Unlock()

	if blockID >= bs.nextBlockID {
		return fmt.Errorf("invalid block ID: %d", blockID)
	}

	if len(data) > bs.blockSize {
		return fmt.Errorf("data size %d exceeds block size %d", len(data), bs.blockSize)
	}

	offset := int64(blockID) * int64(bs.blockSize)

	// Prepare data with padding if necessary
	paddedData := make([]byte, bs.blockSize)
	copy(paddedData, data)

	_, err := bs.file.WriteAt(paddedData, offset)
	if err != nil {
		return fmt.Errorf("failed to write block: %w", err)
	}

	return nil
}

// Insert adds a new block and returns its ID
func (bs *RegularBlockStore) Insert(data []byte) (uint32, error) {
	if len(data) > bs.blockSize {
		return 0, fmt.Errorf("data size %d exceeds block size %d", len(data), bs.blockSize)
	}

	// Use the instance-level mutex instead of the global one
	bs.nextBlockIDMutex.Lock()
	blockID := bs.nextBlockID
	bs.nextBlockID++
	bs.nextBlockIDMutex.Unlock()

	// Then lock only the stripe for this block
	mutex := bs.getMutex(blockID)
	mutex.Lock()
	defer mutex.Unlock()

	offset := int64(blockID) * int64(bs.blockSize)

	// Prepare data with padding if necessary
	paddedData := make([]byte, bs.blockSize)
	copy(paddedData, data)

	_, err := bs.file.WriteAt(paddedData, offset)
	if err != nil {
		return 0, fmt.Errorf("failed to write block: %w", err)
	}

	return blockID, nil
}

// Close closes the underlying file
func (bs *RegularBlockStore) Close() error {
	// Lock all stripes before closing
	for i := range bs.mutexes {
		bs.mutexes[i].Lock()
		defer bs.mutexes[i].Unlock()
	}

	if err := bs.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// NumBlocks returns the number of blocks stored
func (bs *RegularBlockStore) NumBlocks() uint32 {
	bs.nextBlockIDMutex.RLock()
	defer bs.nextBlockIDMutex.RUnlock()
	return bs.nextBlockID
}
