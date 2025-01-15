package overflow

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

const AddressSpaceSize = math.MaxUint32

type Store struct {
	files         map[int]*os.File
	currentFileID int
	path          string
	capacityLimit int
}

func NewStoreWithCapacity(path string, capacityLimit int) (*Store, error) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create overflow directory: %w", err)
	}

	files := make(map[int]*os.File)
	currentFileID := 0

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 4 && entry.Name()[len(entry.Name())-4:] == ".bin" {
			file, err := os.OpenFile(filepath.Join(path, entry.Name()), os.O_RDWR, 0644)
			if err != nil {
				// Close any files we've already opened
				for _, f := range files {
					f.Close()
				}
				return nil, err
			}
			// Extract file number from name by removing .bin extension
			var fileNum int
			_, err = fmt.Sscanf(entry.Name(), "%d.bin", &fileNum)
			if err != nil {
				// Close all files including the one we just opened
				file.Close()
				for _, f := range files {
					f.Close()
				}
				return nil, err
			}
			files[fileNum] = file
			if fileNum > currentFileID {
				currentFileID = fileNum
			}
		}
	}

	// If no files exist, create the initial file
	if len(files) == 0 {
		fileName := fmt.Sprintf("%05d.bin", currentFileID)
		newFilePath := filepath.Join(path, fileName)
		newFile, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		files[currentFileID] = newFile
	}

	return &Store{files: files, currentFileID: currentFileID, path: path, capacityLimit: capacityLimit}, nil
}

func NewStore(path string) (*Store, error) {
	return NewStoreWithCapacity(path, AddressSpaceSize)
}

func (d *Store) Close() error {
	for _, file := range d.files {
		if err := file.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Store) Put(value []byte) ([]byte, error) {
	// Check if the size of the value is less than uint24 (16,777,215), or overflow would happen
	if len(value) >= (1 << 24) {
		return nil, fmt.Errorf("value size %d exceeds maximum allowed size", len(value))
	}

	file := d.files[d.currentFileID]

	// Get current file size
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	startAddress := stat.Size()

	// Check if the start address is beyond the AddressSpaceSize
	if startAddress+int64(len(value)) > int64(d.capacityLimit) {
		// Create a new file (as long as start fits)
		d.currentFileID++
		fileName := fmt.Sprintf("%05d.bin", d.currentFileID)
		newFilePath := filepath.Join(d.path, fileName)
		newFile, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		d.files[d.currentFileID] = newFile
		file = newFile
		startAddress = 0
	}

	// Write value to file
	_, err = file.Write(value)
	if err != nil {
		return nil, err
	}

	// Generate a key and return it
	key, err := GenerateOverflowValue(d.currentFileID, int(startAddress), len(value))
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (d *Store) Get(key []byte) ([]byte, error) {
	// Check if key length is correct
	if len(key) != OverflowValueLength {
		return nil, fmt.Errorf("invalid key length")
	}

	fileID := int(key[0])<<8 | int(key[1])
	startAddress := int64(key[2])<<24 | int64(key[3])<<16 | int64(key[4])<<8 | int64(key[5])
	length := int(key[6])<<16 | int(key[7])<<8 | int(key[8])

	file, exists := d.files[fileID]
	if !exists {
		return nil, fmt.Errorf("file with ID %d does not exist", fileID)
	}

	data := make([]byte, length)
	_, err := file.ReadAt(data, startAddress)
	if err != nil {
		return nil, err
	}

	return data, nil
}
