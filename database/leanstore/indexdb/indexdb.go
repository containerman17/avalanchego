package indexdb

import (
	"bytes"
	"encoding/binary"
	"errors"

	"go.etcd.io/bbolt"
)

type IndexDB struct {
	impl *bbolt.DB
}

const bucketName = "index"

func NewIndexDB(path string) (*IndexDB, error) {

	impl, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	// Create the bucket if it doesn't exist
	err = impl.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		impl.Close()
		return nil, err
	}

	return &IndexDB{impl: impl}, nil
}

func (i *IndexDB) GetFloorValue(key []byte) (uint32, error) {
	var value uint32
	err := i.impl.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()

		// Try to find the key or the next greater one
		k, v := c.Seek(key)

		// If we found an exact match, return it
		if k != nil && bytes.Equal(k, key) {
			value = binary.BigEndian.Uint32(v)
			return nil
		}

		// No exact match, get the previous key
		if k == nil || bytes.Compare(k, key) > 0 {
			k, v = c.Prev()
		}

		if k == nil {
			value = 0
			return nil
		}

		value = binary.BigEndian.Uint32(v)
		return nil
	})
	return value, err
}

// FIXME: might not need to return the key
func (i *IndexDB) GetNextKeyValue(key []byte) ([]byte, uint32, error) {
	var nextKey []byte
	var nextValue uint32

	err := i.impl.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()

		// Seek to the given key
		k, v := c.Seek(key)

		// If we found an exact match, move to the next entry
		if k != nil && bytes.Equal(k, key) {
			k, v = c.Next()
		}

		// If there is no next key, return nil
		if k == nil {
			return nil
		}

		// Copy the key and value
		nextKey = make([]byte, len(k))
		copy(nextKey, k)
		nextValue = binary.BigEndian.Uint32(v)
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	// If no next key was found, return nil
	if nextKey == nil {
		return nil, 0, nil
	}

	return nextKey, nextValue, nil
}

func (i *IndexDB) Put(keys [][]byte, values []uint32) error {
	if len(keys) != len(values) {
		return errors.New("keys and values must have the same length")
	}

	return i.impl.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		for i := range keys {
			// Create a new buffer for each value
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, values[i])
			if err := b.Put(keys[i], buf); err != nil {
				return err
			}
		}
		return nil
	})
}

func (i *IndexDB) Close() error {
	return i.impl.Close()
}
