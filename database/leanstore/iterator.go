// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package leanstore

import (
	"errors"
	"sync"

	"github.com/cockroachdb/pebble"

	"github.com/ava-labs/avalanchego/database"
)

var (
	_ database.Iterator = (*iter)(nil)

	errCouldNotGetValue = errors.New("could not get iterator value")
)

type iter struct {
	// [lock] ensures that only one goroutine can access [iter] at a time.
	// Note that [Database.Close] calls [iter.Release] so we need [lock] to ensure
	// that the user and [Database.Close] don't execute [iter.Release] concurrently.
	// Invariant: [Database.lock] is never grabbed while holding [lock].
	lock sync.Mutex

	db   *Database
	iter *pebble.Iterator

	initialized bool
	closed      bool
	err         error

	hasNext bool
	nextKey []byte
	nextVal []byte
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
