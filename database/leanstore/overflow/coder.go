package overflow

import "fmt"

const OverflowValueLength = 9

// Bytes 1-2 - fileId (max 2^16 - 1)
// Bytes 3-6 - start address in file
// Bytes 7-9 - length of data (uint24)
func GenerateOverflowValue(fileId int, startAddress int, length int) ([]byte, error) {
	if fileId < 0 || fileId >= (1<<16) {
		return nil, fmt.Errorf("fileId %d out of range [0, %d]", fileId, (1<<16)-1)
	}
	if startAddress < 0 {
		return nil, fmt.Errorf("startAddress %d cannot be negative", startAddress)
	}
	if length <= 0 || length >= (1<<24) {
		return nil, fmt.Errorf("length %d out of range [1, %d]", length, (1<<24)-1)
	}

	var key = make([]byte, 9)

	// Bytes 1-2 - fileId
	key[0] = byte(fileId >> 8)
	key[1] = byte(fileId)

	// Bytes 3-6 - start address
	key[2] = byte(startAddress >> 24)
	key[3] = byte(startAddress >> 16)
	key[4] = byte(startAddress >> 8)
	key[5] = byte(startAddress)

	// Bytes 7-9 - length (uint24)
	key[6] = byte(length >> 16)
	key[7] = byte(length >> 8)
	key[8] = byte(length)

	if len(key) != OverflowValueLength {
		panic(fmt.Sprintf("implementation error: key length %d does not match expected length %d", len(key), OverflowValueLength))
	}

	return key, nil
}
