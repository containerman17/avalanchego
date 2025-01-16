// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package leanstore

import (
	"context"
	"encoding/json"
	"path"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/ava-labs/avalanchego/database"
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
)

var (
	_ database.Database = (*Database)(nil)

	DefaultConfig = Config{
		BlockSize:             16 * units.KiB,
		OverflowThresholdSize: 1 * units.KiB,
	}
)

type Database struct {
	closed        bool
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

	valStore, err := valuestore.NewValueStore(path.Join(file, "valuestore"), cfg.BlockSize)
	if err != nil {
		return nil, err
	}

	return &Database{
		config:        cfg,
		overflowStore: overflowStore,
		valueStore:    valStore,
	}, nil
}

func (db *Database) Close() error {
	if db.closed {
		return database.ErrClosed
	}

	db.closed = true

	err := db.overflowStore.Close()
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

	return db.valueStore.Has(key)
}

func (db *Database) Get(key []byte) ([]byte, error) {
	if db.closed {
		return nil, database.ErrClosed
	}

	// fmt.Printf("Get key: %x\n", key)
	value, err := db.valueStore.Get(key)
	// fmt.Printf("Get raw value: %x, err: %v\n", value, err)

	if value == nil {
		return nil, database.ErrNotFound
	}

	result, err := db.untangleRemote(value)
	// fmt.Printf("Get untangled value: %x, err: %v\n", result, err)
	return result, err
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

	// fmt.Printf("Put key: %x, value: %x\n", key, value)
	var metadataByte byte
	if len(value) > db.config.OverflowThresholdSize {
		metadataByte = valuemeta.Remote
	} else {
		metadataByte = valuemeta.NoFlags
	}

	valueWithMeta := make([]byte, len(value)+1)
	valueWithMeta[0] = metadataByte
	copy(valueWithMeta[1:], value)
	// fmt.Printf("Put valueWithMeta: %x\n", valueWithMeta)

	return db.valueStore.Put(key, valueWithMeta)
}

func (db *Database) Delete(key []byte) error {
	if db.closed {
		return database.ErrClosed
	}

	// fmt.Printf("Delete key: %x\n", key)
	return db.valueStore.Delete(key)
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
	return NewIterator(db, start, prefix)
}
