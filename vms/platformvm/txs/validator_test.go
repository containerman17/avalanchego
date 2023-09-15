// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto/secp256k1"
)

const defaultWeight = 10000

// each key controls an address that has [defaultBalance] AVAX at genesis
var keys = secp256k1.TestKeys()

func TestBoundedBy(t *testing.T) {
	require := require.New(t)

	// case 1: a starts, a finishes, b starts, b finishes
	aStartTime := uint64(0)
	aEndTIme := uint64(1)
	a := &Validator{
		NodeID: ids.ShortNodeID(keys[0].PublicKey().Address()),
		Start:  aStartTime,
		End:    aEndTIme,
		Wght:   defaultWeight,
	}

	bStartTime := uint64(2)
	bEndTime := uint64(3)
	b := &Validator{
		NodeID: ids.ShortNodeID(keys[0].PublicKey().Address()),
		Start:  bStartTime,
		End:    bEndTime,
		Wght:   defaultWeight,
	}
	require.False(BoundedBy(a.StartTime(), b.EndTime(), b.StartTime(), b.EndTime()))
	require.False(BoundedBy(b.StartTime(), b.EndTime(), a.StartTime(), a.EndTime()))

	// case 2: a starts, b starts, a finishes, b finishes
	a.Start = 0
	b.Start = 1
	a.End = 2
	b.End = 3
	require.False(BoundedBy(a.StartTime(), a.EndTime(), b.StartTime(), b.EndTime()))
	require.False(BoundedBy(b.StartTime(), b.EndTime(), a.StartTime(), a.EndTime()))

	// case 3: a starts, b starts, b finishes, a finishes
	a.Start = 0
	b.Start = 1
	b.End = 2
	a.End = 3
	require.False(BoundedBy(a.StartTime(), a.EndTime(), b.StartTime(), b.EndTime()))
	require.True(BoundedBy(b.StartTime(), b.EndTime(), a.StartTime(), a.EndTime()))

	// case 4: b starts, a starts, a finishes, b finishes
	b.Start = 0
	a.Start = 1
	a.End = 2
	b.End = 3
	require.True(BoundedBy(a.StartTime(), a.EndTime(), b.StartTime(), b.EndTime()))
	require.False(BoundedBy(b.StartTime(), b.EndTime(), a.StartTime(), a.EndTime()))

	// case 5: b starts, b finishes, a starts, a finishes
	b.Start = 0
	b.End = 1
	a.Start = 2
	a.End = 3
	require.False(BoundedBy(a.StartTime(), a.EndTime(), b.StartTime(), b.EndTime()))
	require.False(BoundedBy(b.StartTime(), b.EndTime(), a.StartTime(), a.EndTime()))

	// case 6: b starts, a starts, b finishes, a finishes
	b.Start = 0
	a.Start = 1
	b.End = 2
	a.End = 3
	require.False(BoundedBy(a.StartTime(), a.EndTime(), b.StartTime(), b.EndTime()))
	require.False(BoundedBy(b.StartTime(), b.EndTime(), a.StartTime(), a.EndTime()))

	// case 3: a starts, b starts, b finishes, a finishes
	a.Start = 0
	b.Start = 0
	b.End = 1
	a.End = 1
	require.True(BoundedBy(a.StartTime(), a.EndTime(), b.StartTime(), b.EndTime()))
	require.True(BoundedBy(b.StartTime(), b.EndTime(), a.StartTime(), a.EndTime()))
}
