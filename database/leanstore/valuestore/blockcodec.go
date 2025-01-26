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

// Block Layout:
// [2 bytes]: block length (uint16, big endian)
// [1 byte]:  shared prefix length
// [N bytes]: shared prefix bytes
// Entries:
//   [1 byte]:  remaining key length
//   [N bytes]: remaining key bytes (after shared prefix)
//   [2 bytes]: value length (uint16, big endian)
//   [M bytes]: value data

// BlockDecoder holds reusable buffers for decoding blocks
type BlockDecoder struct {
	keyBuf []byte
	keys   [][]byte
	values [][]byte
}

// NewBlockDecoder creates a decoder with pre-allocated buffers
func NewBlockDecoder() *BlockDecoder {
	return &BlockDecoder{
		keyBuf: make([]byte, 16384), // Initial size
		keys:   make([][]byte, 0, 99),
		values: make([][]byte, 0, 99),
	}
}

// Decode decodes a block using internal buffers
func (d *BlockDecoder) Decode(block []byte) ([][]byte, [][]byte, error) {
	if len(block) == 0 {
		return [][]byte{}, [][]byte{}, nil
	}

	if len(block) < 2 {
		return nil, nil, errors.New("block too short")
	}

	blockUsedLen := int(block[0])<<8 | int(block[1])
	if blockUsedLen > len(block) {
		return nil, nil, errors.New("invalid block length")
	}

	if blockUsedLen == 2 {
		//this is normal empty block
		return [][]byte{}, [][]byte{}, nil
	}

	// Read block prefix
	pos := 2
	prefixLen := int(block[pos])
	pos++
	if pos+prefixLen >= blockUsedLen {
		return nil, nil, errors.New("invalid prefix length in Decode")
	}
	prefix := block[pos : pos+prefixLen]
	pos += prefixLen

	// Pre-count entries and calculate total key space needed
	totalKeySpace := 0
	scanPos := pos
	for scanPos < blockUsedLen {
		if scanPos+1 >= blockUsedLen {
			return nil, nil, errors.New("malformed block: unexpected end while reading key length")
		}
		keyLen := int(block[scanPos])
		scanPos++
		if scanPos+keyLen+2 >= blockUsedLen {
			return nil, nil, errors.New("malformed block: unexpected end while reading key data")
		}
		totalKeySpace += prefixLen + keyLen
		valueLen := int(block[scanPos+keyLen])<<8 | int(block[scanPos+keyLen+1])
		if scanPos+keyLen+2+valueLen > blockUsedLen {
			return nil, nil, errors.New("malformed block: unexpected end while reading value data")
		}
		scanPos += keyLen + 2 + valueLen
	}

	// Grow key buffer if needed
	if len(d.keyBuf) < totalKeySpace {
		d.keyBuf = make([]byte, totalKeySpace)
	}

	// Reset slices
	d.keys = d.keys[:0]
	d.values = d.values[:0]
	keyBufPos := 0

	// Decode entries
	for pos < blockUsedLen {
		if pos+1 >= blockUsedLen {
			return nil, nil, errors.New("malformed block: unexpected end while reading key length")
		}
		keyLen := int(block[pos])
		pos++
		if pos+keyLen+2 >= blockUsedLen {
			return nil, nil, errors.New("malformed block: unexpected end while reading key data")
		}

		key := d.keyBuf[keyBufPos : keyBufPos+prefixLen+keyLen]
		copy(key[:prefixLen], prefix)
		copy(key[prefixLen:], block[pos:pos+keyLen])
		pos += keyLen
		keyBufPos += prefixLen + keyLen

		valueLen := int(block[pos])<<8 | int(block[pos+1])
		pos += 2
		if pos+valueLen > blockUsedLen {
			return nil, nil, errors.New("malformed block: unexpected end while reading value data")
		}

		value := block[pos : pos+valueLen]
		pos += valueLen

		d.keys = append(d.keys, key)
		d.values = append(d.values, value)
	}

	if len(d.keys) != len(d.values) {
		return nil, nil, errors.New("corrupted block: key count does not match value count")
	}

	return d.keys, d.values, nil
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

	// If both original block and new keys are empty, return empty block
	if len(originalBlock) == 0 && len(keys) == 0 {
		block := make([]byte, 2)
		block[0] = 0
		block[1] = 2
		return block, []NewBlock{}, nil
	}

	allKeys, allValues, err := scanBlockAddKeys(originalBlock, keys, values)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan and merge keys: %w", err)
	}

	quickAndDirtyLengthEstimate := len(originalBlock)
	for i, key := range allKeys {
		quickAndDirtyLengthEstimate += 1 + 1 + len(key) + 2 + len(allValues[i]) // +2 for value length
	}

	//no split required
	if quickAndDirtyLengthEstimate < blockSize {
		willFit, err := willItFit(allKeys, allValues, blockSize)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to check if keys fit: %w", err)
		}
		if willFit {
			_, updatedBlockBytes, err := packMaxKeys(allKeys, allValues, blockSize, blockSize)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to pack keys: %w", err)
			}
			return updatedBlockBytes, []NewBlock{}, nil
		}
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
	if keysPacked == 0 {
		return nil, nil, errors.New("failed to pack any keys into first block")
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
			return nil, nil, fmt.Errorf("failed to pack keys: zero keys packed with %s remaining", debugDescribeBytesSlices(remainingKeys))
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

func debugDescribeBytesSlices(slices [][]byte) string {
	maxLen := 0
	minLen := len(slices[0])
	for _, slice := range slices {
		if len(slice) > maxLen {
			maxLen = len(slice)
		}
		if len(slice) < minLen {
			minLen = len(slice)
		}
	}
	return fmt.Sprintf("%d slices with len from %d to %d", len(slices), minLen, maxLen)
}

func scanBlockAddKeys(originalBlock []byte, keys [][]byte, values [][]byte) ([][]byte, [][]byte, error) {
	if len(originalBlock) == 0 {
		return keys, values, nil
	}

	if len(originalBlock) < 2 {
		return nil, nil, errors.New("malformed block: block too short")
	}

	blockUsedLen := int(originalBlock[0])<<8 | int(originalBlock[1])
	if blockUsedLen > len(originalBlock) {
		return nil, nil, errors.New("malformed block: invalid block length")
	}

	// Empty block with just header
	if blockUsedLen <= 2 {
		return keys, values, nil
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

	// Read block prefix
	pos := 2
	prefixLen := int(originalBlock[pos])
	pos++
	if pos+prefixLen >= blockUsedLen {
		return nil, nil, errors.New("malformed block: invalid prefix length")
	}
	prefix := originalBlock[pos : pos+prefixLen]
	pos += prefixLen

	mergedKeys := make([][]byte, 0)
	mergedValues := make([][]byte, 0)
	newKeyIndex := 0

	// Merge sorted original block with sorted new keys
	for pos < blockUsedLen {
		if pos+1 >= blockUsedLen {
			return nil, nil, errors.New("malformed block: unexpected end while reading key length")
		}
		keyLen := int(originalBlock[pos])
		pos++

		if pos+keyLen+2 >= blockUsedLen {
			return nil, nil, errors.New("malformed block: unexpected end while reading key data")
		}

		// Reconstruct full key
		key := make([]byte, prefixLen+keyLen)
		copy(key, prefix)
		copy(key[prefixLen:], originalBlock[pos:pos+keyLen])
		pos += keyLen

		if pos+2 > blockUsedLen {
			return nil, nil, errors.New("malformed block: unexpected end while reading value length")
		}
		valueLen := int(originalBlock[pos])<<8 | int(originalBlock[pos+1])
		pos += 2
		if pos+valueLen > blockUsedLen {
			return nil, nil, errors.New("malformed block: unexpected end while reading value data")
		}
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
	}

	// Add any remaining new keys
	for newKeyIndex < len(keys) {
		mergedKeys = append(mergedKeys, keys[newKeyIndex])
		mergedValues = append(mergedValues, values[newKeyIndex])
		newKeyIndex++
	}

	if len(mergedKeys) != len(mergedValues) {
		return nil, nil, fmt.Errorf("internal error: key count (%d) does not match value count (%d) after merge", len(mergedKeys), len(mergedValues))
	}

	return mergedKeys, mergedValues, nil
}

func willItFit(keys [][]byte, values [][]byte, hardMaxSize int) (bool, error) {
	if len(keys) != len(values) {
		return false, fmt.Errorf("key count (%d) does not match value count (%d)", len(keys), len(values))
	}

	if len(keys) == 0 {
		return true, nil
	}

	// Find longest common prefix
	firstKey := keys[0]
	prefixLen := len(firstKey)
	for _, key := range keys[1:] {
		for i := 0; i < prefixLen; i++ {
			if i >= len(key) || key[i] != firstKey[i] {
				prefixLen = i
				break
			}
		}
		if prefixLen == 0 {
			break
		}
	}

	// Calculate total size
	totalSize := 2 // block length
	totalSize += 1 // prefix length
	totalSize += prefixLen

	for i, key := range keys {
		if len(key) < prefixLen {
			return false, fmt.Errorf("key length (%d) shorter than prefix length (%d)", len(key), prefixLen)
		}
		remainingKeyLen := len(key) - prefixLen
		if remainingKeyLen > 255 {
			return false, fmt.Errorf("remaining key length (%d) exceeds maximum (255)", remainingKeyLen)
		}
		totalSize += 1 // remaining key length
		totalSize += remainingKeyLen
		totalSize += 2 // value length
		totalSize += len(values[i])
	}

	return totalSize <= hardMaxSize, nil
}

// packs maximum number of keys into set length, stops AFTER reaches preferredMinSize if possible, but not more than hardMaxSize
func packMaxKeys(keys [][]byte, values [][]byte, preferredMinSize, hardMaxSize int) (int, []byte, error) {
	if len(keys) != len(values) {
		return 0, nil, fmt.Errorf("key count (%d) does not match value count (%d)", len(keys), len(values))
	}

	if len(keys) == 0 {
		// Empty block with just header
		block := make([]byte, 2)
		block[0] = 0
		block[1] = 2
		return 0, block, nil
	}

	if preferredMinSize > hardMaxSize {
		return 0, nil, fmt.Errorf("preferredMinSize (%d) greater than hardMaxSize (%d)", preferredMinSize, hardMaxSize)
	}

	// Find shared prefix length
	firstKey := keys[0]
	prefixLen := len(firstKey)
	for _, key := range keys[1:] {
		for i := 0; i < prefixLen; i++ {
			if i >= len(key) || key[i] != firstKey[i] {
				prefixLen = i
				break
			}
		}
		if prefixLen == 0 {
			break
		}
	}

	// Validate key lengths
	for _, key := range keys {
		if len(key) < prefixLen {
			return 0, nil, fmt.Errorf("key length (%d) shorter than prefix length (%d)", len(key), prefixLen)
		}
		remainingKeyLen := len(key) - prefixLen
		if remainingKeyLen > 255 {
			return 0, nil, fmt.Errorf("remaining key length (%d) exceeds maximum (255)", remainingKeyLen)
		}
	}

	// Calculate block size: header + prefix + entries
	blockSize := 2 + 1 + prefixLen // length + prefixLen + prefix
	keysPacked := 0

	// Pack as many keys as possible
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		value := values[i]

		remainingKeyLen := len(key) - prefixLen
		entrySize := 1 + remainingKeyLen + 2 + len(value) // keyLen + key + valueLen + value
		
		if blockSize+entrySize > hardMaxSize {
			break
		}

		blockSize += entrySize
		keysPacked++

		if blockSize >= preferredMinSize && keysPacked > 0 {
			break
		}
	}

	if keysPacked == 0 {
		return 0, nil, errors.New("could not pack any keys within size constraints")
	}

	// Now pack the actual block
	block := make([]byte, blockSize)
	block[0] = byte(blockSize >> 8)
	block[1] = byte(blockSize)
	block[2] = byte(prefixLen)
	copy(block[3:], firstKey[:prefixLen])
	pos := 3 + prefixLen

	for i := 0; i < keysPacked; i++ {
		key := keys[i]
		value := values[i]

		remainingKeyLen := len(key) - prefixLen
		block[pos] = byte(remainingKeyLen)
		pos++
		copy(block[pos:], key[prefixLen:])
		pos += remainingKeyLen

		block[pos] = byte(len(value) >> 8)
		block[pos+1] = byte(len(value))
		pos += 2
		copy(block[pos:], value)
		pos += len(value)
	}

	return keysPacked, block, nil
}

func GetValue(block []byte, key []byte) (bool, []byte, error) {
	if len(block) < 2 {
		return false, nil, errors.New("block too short")
	}

	// Read block length from header
	blockUsedLen := int(block[0])<<8 | int(block[1])

	// Special case for empty block (just header)
	if blockUsedLen == 2 {
		return false, nil, nil
	}

	if blockUsedLen > len(block) {
		return false, nil, errors.New("invalid block length")
	}

	// Read block prefix
	pos := 2
	prefixLen := int(block[pos])
	pos++
	if pos+prefixLen >= blockUsedLen {
		fmt.Printf("invalid prefix length in GetValue, full block: %x\n", block)
		return false, nil, errors.New("invalid prefix length in GetValue")
	}
	prefix := block[pos : pos+prefixLen]
	pos += prefixLen

	// Check if key matches prefix
	if len(key) < prefixLen {
		return false, nil, nil // Key too short to match prefix
	}
	if !bytes.Equal(key[:prefixLen], prefix) {
		return false, nil, nil // Prefix doesn't match
	}

	// Scan entries for exact match
	searchRemaining := key[prefixLen:]
	for pos < blockUsedLen {
		if pos+1 >= blockUsedLen {
			break
		}
		keyLen := int(block[pos])
		pos++

		if pos+keyLen+2 >= blockUsedLen {
			break
		}
		remainingKey := block[pos : pos+keyLen]
		pos += keyLen

		valueLen := int(block[pos])<<8 | int(block[pos+1])
		pos += 2
		if pos+valueLen > blockUsedLen {
			break
		}

		// Compare with search key
		if bytes.Equal(remainingKey, searchRemaining) {
			// Exact match
			value := make([]byte, valueLen)
			copy(value, block[pos:pos+valueLen])
			return true, value, nil
		}
		pos += valueLen
	}

	return false, nil, nil
}
