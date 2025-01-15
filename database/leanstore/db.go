// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package leanstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leanstore/backedbuffer"
	"github.com/ava-labs/avalanchego/database/leanstore/overflow"
	"github.com/ava-labs/avalanchego/database/leanstore/valuemeta"
	"github.com/ava-labs/avalanchego/database/leanstore/valuestore"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/units"
)

const (
	Name = "leanstore"

	// pebbleByteOverHead is the number of bytes of constant overhead that
	// should be added to a batch size per operation.
	pebbleByteOverHead = 8

	defaultCacheSize = 512 * units.MiB
)

var (
	_ database.Database = (*Database)(nil)

	errInvalidOperation = errors.New("invalid operation")

	DefaultConfig = Config{
		BlockSize:             16 * units.KiB,
		OverflowThresholdSize: 1 * units.KiB,
	}
)

type Database struct {
	closed        bool
	backedBuffer1 *backedbuffer.BackedBuffer
	config        Config
	overflowStore *overflow.Store
	valueStore    *valuestore.ValueStore
}

type Config struct {
	OverflowThresholdSize int `json:"overflowThresholdSize"`
	BlockSize             int `json:"blockSize"`
}

func New(file string, configBytes []byte, log logging.Logger, _ prometheus.Registerer) (database.Database, error) {
	cfg := DefaultConfig
	if len(configBytes) > 0 {
		if err := json.Unmarshal(configBytes, &cfg); err != nil {
			return nil, err
		}
	}

	log.Info(
		"opening leanstore",
		zap.Reflect("config", cfg),
	)

	overflowStore, err := overflow.NewStore(path.Join(file, "overflow"))
	if err != nil {
		return nil, err
	}

	backedBuffer1, err := backedbuffer.New(path.Join(file, "backedbuffer1"))
	if err != nil {
		return nil, err
	}

	valStore, err := valuestore.NewValueStore(path.Join(file, "valuestore"), cfg.BlockSize)
	if err != nil {
		return nil, err
	}

	return &Database{
		config:        cfg,
		overflowStore: overflowStore,
		backedBuffer1: backedBuffer1,
		valueStore:    valStore,
	}, nil
}

func (db *Database) Close() error {
	if db.closed {
		return database.ErrClosed
	}

	db.closed = true

	err := db.backedBuffer1.Close()
	if err != nil {
		return err
	}

	err = db.overflowStore.Close()
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) HealthCheck(_ context.Context) (interface{}, error) {
	if db.closed {
		return nil, database.ErrClosed
	}
	return nil, nil
}

func (db *Database) Has(key []byte) (bool, error) {
	if db.closed {
		return false, database.ErrClosed
	}

	// Check in backedBuffer1
	has, err := db.backedBuffer1.Has(key)
	if err != nil {
		return false, err
	}

	if has {
		// If the key exists in backedBuffer1, we need to check if it's a tombstone
		value, err := db.backedBuffer1.Get(key)
		if err != nil {
			return false, err
		}
		if valuemeta.IsTombstone(value) {
			fmt.Printf("Has key=%x tombstone\n", key)
			return false, nil
		}
		fmt.Printf("Has key=%x\n", key)
		return true, nil
	}

	has, err = db.valueStore.Has(key)
	if err != nil {
		return false, err
	}
	fmt.Printf("Has key=%x from valueStore: %t\n", key, has)
	return has, nil
}

func (db *Database) Get(key []byte) ([]byte, error) {
	if db.closed {
		return nil, database.ErrClosed
	}

	var value []byte
	var err error

	//check in the temp store1
	value, err = db.backedBuffer1.Get(key)
	if err != nil {
		return nil, err
	}

	if value != nil {
		return db.untangleRemote(value)
	}

	//TODO: check in the temp store2 when implemented

	//get from the main store
	value, err = db.valueStore.Get(key)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, database.ErrNotFound
	}

	return db.untangleRemote(value)
}

func (db *Database) untangleRemote(value []byte) ([]byte, error) {
	if value == nil {
		panic("implementation error: value is nil in untangleRemote")
	}

	if valuemeta.IsTombstone(value) {
		return nil, database.ErrNotFound
	}

	// Strip off the metadata byte for non-remote values
	if !valuemeta.IsRemote(value) {
		return value[1:], nil
	}

	// For remote values, get from overflow store
	return db.overflowStore.Get(value[1:])
}

func (db *Database) Put(key []byte, value []byte) error {
	if db.closed {
		return database.ErrClosed
	}

	var metadataByte byte
	if len(value) > db.config.OverflowThresholdSize {
		metadataByte = valuemeta.Remote
	} else {
		metadataByte = valuemeta.NoFlags
	}

	valueWithMeta := make([]byte, len(value)+1)
	valueWithMeta[0] = metadataByte
	copy(valueWithMeta[1:], value)

	fmt.Printf("Put key=%x value=%x\n", key, valueWithMeta)

	return db.backedBuffer1.Put([][]byte{key}, [][]byte{valueWithMeta})
}

func (db *Database) Delete(key []byte) error {
	if db.closed {
		return database.ErrClosed
	}

	fmt.Printf("Delete key=%x\n", key)

	return db.backedBuffer1.Put([][]byte{key}, [][]byte{{valuemeta.Tombstone}})
}

func (db *Database) Compact(start []byte, end []byte) error {
	return nil //no need for compaction
}

func (db *Database) NewIterator() database.Iterator {
	return db.NewIteratorWithStartAndPrefix(nil, nil)
}

func (db *Database) NewIteratorWithStart(start []byte) database.Iterator {
	return db.NewIteratorWithStartAndPrefix(start, nil)
}

func (db *Database) NewIteratorWithPrefix(prefix []byte) database.Iterator {
	return db.NewIteratorWithStartAndPrefix(nil, prefix)
}

func (db *Database) NewIteratorWithStartAndPrefix(start, prefix []byte) database.Iterator {
	panic("not implemented")
}
