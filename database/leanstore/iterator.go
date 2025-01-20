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
	end      []byte
	nextKeys [][]byte
	nextVals [][]byte
}

func NewIterator(db *Database, start []byte, prefix []byte) *iter {
	// If start is nil, use minimum possible key
	if start == nil {
		start = make([]byte, 1)
		// defaults to 0x00
	}

	// Pre-allocate end key with full capacity
	end := make([]byte, 255)
	if prefix != nil {
		// Copy prefix to the beginning
		copy(end, prefix)
	}
	// Fill the rest with 0xff bytes (works for both nil and non-nil prefix)
	for i := 0; i < 255; i++ {
		if prefix == nil || i >= len(prefix) {
			end[i] = 0xff
		}
	}

	return &iter{
		db:       db,
		start:    start,
		end:      end,
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
