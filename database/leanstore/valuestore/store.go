package valuestore

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leanstore/indexdb"
)

type ValueStore struct {
	index *indexdb.IndexDB
}

// Delete implements database.KeyValueReaderWriterDeleter.
func (v *ValueStore) Delete(key []byte) error {
	panic("unimplemented")
}

// Get implements database.KeyValueReaderWriterDeleter.
func (v *ValueStore) Get(key []byte) ([]byte, error) {
	panic("unimplemented")
}

// Has implements database.KeyValueReaderWriterDeleter.
func (v *ValueStore) Has(key []byte) (bool, error) {
	panic("unimplemented")
}

// Put implements database.KeyValueReaderWriterDeleter.
func (v *ValueStore) Put(key []byte, value []byte) error {
	panic("unimplemented")
}

var _ database.KeyValueReaderWriterDeleter = (*ValueStore)(nil)
