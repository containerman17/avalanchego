// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package leanstore

import (
	"github.com/cockroachdb/pebble"

	"github.com/ava-labs/avalanchego/database"
)

var _ database.Batch = (*batch)(nil)

// Not safe for concurrent use.
type batch struct {
	batch *pebble.Batch
	db    *Database
	size  int

	// True iff [batch] has been written to the database
	// since the last time [Reset] was called.
	written bool
}

// Delete implements database.Batch.
func (b *batch) Delete(key []byte) error {
	panic("unimplemented")
}

// Inner implements database.Batch.
func (b *batch) Inner() database.Batch {
	panic("unimplemented")
}

// Put implements database.Batch.
func (b *batch) Put(key []byte, value []byte) error {
	panic("unimplemented")
}

// Replay implements database.Batch.
func (b *batch) Replay(w database.KeyValueWriterDeleter) error {
	panic("unimplemented")
}

// Reset implements database.Batch.
func (b *batch) Reset() {
	panic("unimplemented")
}

// Size implements database.Batch.
func (b *batch) Size() int {
	panic("unimplemented")
}

// Write implements database.Batch.
func (b *batch) Write() error {
	panic("unimplemented")
}

func (db *Database) NewBatch() database.Batch {
	panic("not implemented")
}
