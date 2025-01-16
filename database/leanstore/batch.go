// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package leanstore

import (
	"github.com/cockroachdb/pebble"

	"github.com/ava-labs/avalanchego/database"
)

var _ database.Batch = (*batch)(nil)

type operation struct {
	key    []byte
	value  []byte
	delete bool
}

// Not safe for concurrent use.
type batch struct {
	batch *pebble.Batch
	db    *Database
	size  int

	// True iff [batch] has been written to the database
	// since the last time [Reset] was called.
	written bool

	// Store operations in order
	ops []operation
}

// Delete implements database.Batch.
func (b *batch) Delete(key []byte) error {
	if b.written {
		return database.ErrClosed
	}
	// Make a copy of key
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	b.ops = append(b.ops, operation{
		key:    keyCopy,
		delete: true,
	})
	b.size += len(key) + pebbleByteOverHead
	return nil
}

// Inner implements database.Batch.
func (b *batch) Inner() database.Batch {
	return b
}

// Put implements database.Batch.
func (b *batch) Put(key []byte, value []byte) error {
	if b.written {
		return database.ErrClosed
	}
	// Make copies of key and value
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	b.ops = append(b.ops, operation{
		key:   keyCopy,
		value: valueCopy,
	})
	b.size += len(key) + len(value) + pebbleByteOverHead
	return nil
}

// Replay implements database.Batch.
func (b *batch) Replay(w database.KeyValueWriterDeleter) error {
	for _, op := range b.ops {
		if op.delete {
			if err := w.Delete(op.key); err != nil {
				return err
			}
		} else {
			if err := w.Put(op.key, op.value); err != nil {
				return err
			}
		}
	}
	return nil
}

// Reset implements database.Batch.
func (b *batch) Reset() {
	b.ops = b.ops[:0]
	b.size = 0
	b.written = false
}

// Size implements database.Batch.
func (b *batch) Size() int {
	return b.size
}

// Write implements database.Batch.
func (b *batch) Write() error {
	if b.db.closed {
		return database.ErrClosed
	}

	// Apply all operations in order
	for _, op := range b.ops {
		if op.delete {
			if err := b.db.Delete(op.key); err != nil {
				return err
			}
		} else {
			if err := b.db.Put(op.key, op.value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *Database) NewBatch() database.Batch {
	return &batch{
		db: db,
	}
}
