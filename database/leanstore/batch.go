// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package leanstore

import (
	"github.com/cockroachdb/pebble"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leanstore/valuemeta"
)

var _ database.Batch = (*batch)(nil)

type operation struct {
	key          []byte
	value        []byte
	originalValue []byte
	delete       bool
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

	// Store the original value for replay
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	// Handle overflow values
	var metadataByte byte
	var valueToStore []byte

	if len(value) > b.db.config.OverflowThresholdSize {
		metadataByte = valuemeta.Remote
		// Store in overflow store
		overflowKey, err := b.db.overflowStore.Put(value)
		if err != nil {
			return err
		}
		valueToStore = overflowKey
	} else {
		metadataByte = valuemeta.NoFlags
		valueToStore = value
	}

	// Create value with metadata
	valueWithMeta := make([]byte, len(valueToStore)+1)
	valueWithMeta[0] = metadataByte
	copy(valueWithMeta[1:], valueToStore)

	b.ops = append(b.ops, operation{
		key:          keyCopy,
		value:        valueWithMeta,  // For writing to database
		originalValue: valueCopy, // For replay
	})
	// Size should reflect the original key and value sizes as expected by tests
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
			// Use original value for replay
			if err := w.Put(op.key, op.originalValue); err != nil {
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
			// Value already has metadata byte, so we can write directly to valueStore
			if err := b.db.valueStore.Put(op.key, op.value); err != nil {
				return err
			}
		}
	}

	b.written = true // Set written flag after successful write
	return nil
}

func (db *Database) NewBatch() database.Batch {
	return &batch{
		db: db,
	}
}
