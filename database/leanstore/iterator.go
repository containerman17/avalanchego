// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package leanstore

import (
	"github.com/ava-labs/avalanchego/database"
)

var (
	_ database.Iterator = (*iter)(nil)
)

type iter struct {
	db       *Database
	start    []byte
	prefix   []byte
	nextKeys [][]byte
	nextVals [][]byte
}

func NewIterator(db *Database, start []byte, prefix []byte) *iter {
	return &iter{
		db:       db,
		start:    start,
		prefix:   prefix,
		nextKeys: make([][]byte, 1),
		nextVals: make([][]byte, 1),
	}
}

// Error implements database.Iterator.
func (i *iter) Error() error {
	panic("unimplemented")
}

// Key implements database.Iterator.
func (i *iter) Key() []byte {
	panic("unimplemented")
}

// Next implements database.Iterator.
func (i *iter) Next() bool {
	panic("unimplemented")
}

// Release implements database.Iterator.
func (i *iter) Release() {
	panic("unimplemented")
}

// Value implements database.Iterator.
func (i *iter) Value() []byte {
	panic("unimplemented")
}
