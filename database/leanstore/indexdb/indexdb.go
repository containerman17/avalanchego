package indexdb

import (
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

func (i *IndexDB) Delete(key []byte) error {
	return i.impl.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Delete(key)
	})
}

func (i *IndexDB) Get(key []byte) (uint32, error) {
	var value uint32
	err := i.impl.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get(key)
		if v == nil {
			return errors.New("key not found")
		}
		value = binary.BigEndian.Uint32(v)
		return nil
	})
	return value, err
}

func (i *IndexDB) Put(keys [][]byte, values []uint32) error {
	if len(keys) != len(values) {
		return errors.New("keys and values must have the same length")
	}

	return i.impl.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		buf := make([]byte, 4)
		for i := range keys {
			binary.BigEndian.PutUint32(buf, values[i])
			if err := b.Put(keys[i], buf); err != nil {
				return err
			}
		}
		return nil
	})
}
