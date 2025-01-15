package valuestore

import (
	"bytes"
	"errors"
	"fmt"
)

type NewBlock struct {
	Block       []byte
	StartingKey []byte
}

func EncodeBlock(originalBlock []byte, keys [][]byte, values [][]byte, blockSize int) ([]byte, []NewBlock, error) {
	if blockSize > 256*256 {
		return nil, nil, errors.New("block size too large")
	}

	// Validate value sizes
	for _, value := range values {
		if len(value) > 65535 { // uint16 max
			return nil, nil, fmt.Errorf("value size %d exceeds maximum allowed size of 65535", len(value))
		}
	}

	allKeys, allValues := scanBlockAddKeys(originalBlock, keys, values)

	quickAndDirtyLengthEstimate := len(originalBlock)
	for i, key := range allKeys {
		quickAndDirtyLengthEstimate += 1 + 1 + len(key) + 2 + len(allValues[i]) // +2 for value length
	}

	//no split required
	if quickAndDirtyLengthEstimate < blockSize || willItFit(allKeys, allValues, blockSize) {
		_, updatedBlockBytes, err := packMaxKeys(allKeys, allValues, blockSize, blockSize)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to pack keys: %w", err)
		}
		return updatedBlockBytes, []NewBlock{}, nil
	}

	//split required, packing half-full blocks
	remainingKeys := allKeys
	remainingValues := allValues
	var firstBlock []byte
	newBlocks := make([]NewBlock, 0)

	// Pack the first block
	keysPacked, blockBytes, err := packMaxKeys(remainingKeys, remainingValues, blockSize/2, blockSize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to pack first block: %w", err)
	}
	firstBlock = blockBytes
	remainingKeys = remainingKeys[keysPacked:]
	remainingValues = remainingValues[keysPacked:]

	// Pack remaining blocks
	for len(remainingKeys) > 0 {
		keysPacked, blockBytes, err := packMaxKeys(remainingKeys, remainingValues, blockSize/2, blockSize)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to pack subsequent block: %w", err)
		}
		if keysPacked == 0 {
			return nil, nil, errors.New("failed to pack keys: zero keys packed")
		}

		newBlocks = append(newBlocks, NewBlock{
			Block:       blockBytes,
			StartingKey: remainingKeys[0],
		})

		remainingKeys = remainingKeys[keysPacked:]
		remainingValues = remainingValues[keysPacked:]
	}

	return firstBlock, newBlocks, nil
}

func scanBlockAddKeys(originalBlock []byte, keys [][]byte, values [][]byte) ([][]byte, [][]byte) {
	if len(originalBlock) < 2 {
		return keys, values
	}

	blockUsedLen := int(originalBlock[0])<<8 | int(originalBlock[1])
	if blockUsedLen > len(originalBlock) {
		return keys, values
	}

	// Sort the new keys and values first
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if bytes.Compare(keys[i], keys[j]) > 0 {
				keys[i], keys[j] = keys[j], keys[i]
				values[i], values[j] = values[j], values[i]
			}
		}
	}

	pos := 2
	lastKey := []byte{}
	mergedKeys := make([][]byte, 0)
	mergedValues := make([][]byte, 0)
	newKeyIndex := 0

	// Merge sorted original block with sorted new keys
	for pos < blockUsedLen {
		shared := int(originalBlock[pos])
		pos++
		keyLen := int(originalBlock[pos])
		pos++

		key := make([]byte, shared)
		copy(key, lastKey[:shared])
		key = append(key, originalBlock[pos:pos+keyLen]...)
		pos += keyLen

		valueLen := int(originalBlock[pos])<<8 | int(originalBlock[pos+1])
		pos += 2
		value := make([]byte, valueLen)
		copy(value, originalBlock[pos:pos+valueLen])
		pos += valueLen

		// Add any new keys that come before current key
		for newKeyIndex < len(keys) && bytes.Compare(keys[newKeyIndex], key) < 0 {
			mergedKeys = append(mergedKeys, keys[newKeyIndex])
			mergedValues = append(mergedValues, values[newKeyIndex])
			newKeyIndex++
		}

		// If current key exists in new keys, use the new value
		if newKeyIndex < len(keys) && bytes.Equal(key, keys[newKeyIndex]) {
			mergedKeys = append(mergedKeys, keys[newKeyIndex])
			mergedValues = append(mergedValues, values[newKeyIndex])
			newKeyIndex++
		} else {
			// Otherwise keep the existing key-value pair
			mergedKeys = append(mergedKeys, key)
			mergedValues = append(mergedValues, value)
		}

		lastKey = key
	}

	// Add any remaining new keys
	for newKeyIndex < len(keys) {
		mergedKeys = append(mergedKeys, keys[newKeyIndex])
		mergedValues = append(mergedValues, values[newKeyIndex])
		newKeyIndex++
	}

	return mergedKeys, mergedValues
}

const CMP_FIRST_IS_LESS = -1
const CMP_FIRST_IS_EQUAL = 0
const CMP_FIRST_IS_GREATER = 1

func FindFloorValue(block []byte, key []byte) (bool, []byte, error) {
	if len(block) < 2 {
		return false, nil, errors.New("block too short")
	}

	blockUsedLen := int(block[0])<<8 | int(block[1])
	if blockUsedLen > len(block) {
		return false, nil, errors.New("invalid block length")
	}

	offset := 2 // Start after header
	lastKey := []byte{}
	found := false
	var lastValue []byte

	for offset < blockUsedLen {
		shared := int(block[offset])
		offset++
		keyLen := int(block[offset])
		offset++

		currentKey := make([]byte, shared)
		copy(currentKey, lastKey[:shared])
		currentKey = append(currentKey, block[offset:offset+keyLen]...)
		offset += keyLen

		valueLen := int(block[offset])<<8 | int(block[offset+1])
		offset += 2
		value := make([]byte, valueLen)
		copy(value, block[offset:offset+valueLen])
		offset += valueLen

		cmp := bytes.Compare(currentKey, key)
		if cmp == CMP_FIRST_IS_EQUAL {
			return true, value, nil
		} else if cmp == CMP_FIRST_IS_LESS {
			found = true
			lastValue = value
		} else {
			break
		}

		lastKey = currentKey
	}

	if found {
		return true, lastValue, nil
	}

	return false, nil, nil
}

func willItFit(keys [][]byte, values [][]byte, hardMaxSize int) bool {
	blockSize := 2 // Start with 2 bytes for length
	lastKey := []byte{}

	for i := 0; i < len(keys); i++ {
		key := keys[i]
		value := values[i]

		shared := 0
		for shared < len(lastKey) && shared < len(key) && lastKey[shared] == key[shared] {
			shared++
		}

		expectedLen := 1 + 1 + (len(key) - shared) + 2 + len(value) // shared + keylen + key + valuelen + value

		if blockSize+expectedLen > hardMaxSize {
			return false
		}

		blockSize += expectedLen
		lastKey = key
	}

	return true
}

// packs maximum number of keys into set length, stops AFTER reaches preferredMinSize if possible, but not more than hardMaxSize
func packMaxKeys(keys [][]byte, values [][]byte, preferredMinSize, hardMaxSize int) (int, []byte, error) {
	block := make([]byte, 2, hardMaxSize)
	keysPackedLength := 0

	lastKey := []byte{}
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		value := values[i]

		shared := 0
		for shared < len(lastKey) && shared < len(key) && lastKey[shared] == key[shared] {
			shared++
		}

		expectedLen := 1 + 1 + (len(key) - shared) + 2 + len(value)

		if len(block)+expectedLen > hardMaxSize {
			break
		}

		block = append(block, uint8(shared))
		block = append(block, uint8(len(key)-shared))
		block = append(block, key[shared:]...)

		// Encode value length and value
		block = append(block, byte(len(value)>>8))
		block = append(block, byte(len(value)))
		block = append(block, value...)

		lastKey = key
		keysPackedLength++

		if len(block) >= preferredMinSize {
			break
		}
	}

	// Write the block length
	usedLength := len(block)
	block[0] = byte(usedLength >> 8)
	block[1] = byte(usedLength)

	return keysPackedLength, block, nil
}

// GetFirstKeyValue returns the first key-value pair in a block
func GetFirstKeyValue(block []byte) ([]byte, []byte, error) {
	if len(block) < 2 {
		return nil, nil, fmt.Errorf("block too short")
	}

	blockUsedLen := int(block[0])<<8 | int(block[1])
	if blockUsedLen > len(block) {
		return nil, nil, fmt.Errorf("invalid block length")
	}

	pos := 2
	if pos >= blockUsedLen {
		return nil, nil, fmt.Errorf("block is empty")
	}

	//no shared prefix
	pos++
	keyLen := int(block[pos])
	pos++

	if pos+keyLen+2 > blockUsedLen {
		return nil, nil, fmt.Errorf("invalid key length")
	}

	key := make([]byte, keyLen)
	copy(key, block[pos:pos+keyLen])
	pos += keyLen

	valueLen := int(block[pos])<<8 | int(block[pos+1])
	pos += 2
	value := make([]byte, valueLen)
	copy(value, block[pos:pos+valueLen])
	pos += valueLen

	return key, value, nil
}
