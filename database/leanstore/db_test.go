// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package leanstore

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanchego/database/dbtest"
	"github.com/ava-labs/avalanchego/utils/logging"
)

func newDB(t testing.TB) *Database {
	folder := t.TempDir()
	db, err := New(folder, nil, logging.NoLog{}, prometheus.NewRegistry())
	require.NoError(t, err)
	return db.(*Database)
}

func TestBasic(t *testing.T) {
	for name, test := range dbtest.TestsBasic {
		t.Run(name, func(t *testing.T) {
			db := newDB(t)
			test(t, db)
			_ = db.Close()
		})
	}
}

func TestInterface(t *testing.T) {
	for name, test := range dbtest.Tests {
		if strings.Contains(name, "Iterator") || strings.Contains(name, "Compact") || strings.Contains(name, "Prefix") || strings.Contains(name, "Clear") { //TODO: remove
			continue
		}
		// if !strings.Contains(name, "ModifyValueAfterBatchPut") {
		// 	continue
		// }
		t.Run(name, func(t *testing.T) {
			db := newDB(t)
			test(t, db)
			_ = db.Close()
		})
	}
}

func FuzzKeyValue(f *testing.F) {
	db := newDB(f)
	dbtest.FuzzKeyValue(f, db)
	_ = db.Close()
}

func FuzzNewIteratorWithPrefix(f *testing.F) {
	db := newDB(f)
	dbtest.FuzzNewIteratorWithPrefix(f, db)
	_ = db.Close()
}

func FuzzNewIteratorWithStartAndPrefix(f *testing.F) {
	db := newDB(f)
	dbtest.FuzzNewIteratorWithStartAndPrefix(f, db)
	_ = db.Close()
}

// func BenchmarkInterface(b *testing.B) {
// 	for _, size := range dbtest.BenchmarkSizes {
// 		keys, values := dbtest.SetupBenchmark(b, size[0], size[1], size[2])
// 		for name, bench := range dbtest.Benchmarks {
// 			b.Run(fmt.Sprintf("leanstore_%d_pairs_%d_keys_%d_values_%s", size[0], size[1], size[2], name), func(b *testing.B) {
// 				db := newDB(b)
// 				bench(b, db, keys, values)
// 				_ = db.Close()
// 			})
// 		}
// 	}
// }
