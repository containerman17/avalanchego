// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShortString(t *testing.T) {
	id := ShortID{1}

	xPrefixedID := id.PrefixedString("X-")
	pPrefixedID := id.PrefixedString("P-")

	newID, err := ShortFromPrefixedString(xPrefixedID, "X-")
	if err != nil {
		t.Fatal(err)
	}
	if newID != id {
		t.Fatalf("ShortFromPrefixedString did not produce the identical ID")
	}

	_, err = ShortFromPrefixedString(pPrefixedID, "X-")
	if err == nil {
		t.Fatal("Using the incorrect prefix did not cause an error")
	}

	tooLongPrefix := "hasnfaurnourieurn3eiur3nriu3nri34iurni34unr3iunrasfounaeouern3ur"
	_, err = ShortFromPrefixedString(xPrefixedID, tooLongPrefix)
	if err == nil {
		t.Fatal("Using the incorrect prefix did not cause an error")
	}
}

func TestShortIDMapMarshalling(t *testing.T) {
	originalMap := map[ShortID]int{
		{'e', 'v', 'a', ' ', 'l', 'a', 'b', 's'}: 1,
		{'a', 'v', 'a', ' ', 'l', 'a', 'b', 's'}: 2,
	}
	mapJSON, err := json.Marshal(originalMap)
	if err != nil {
		t.Fatal(err)
	}

	var unmarshalledMap map[ShortID]int
	err = json.Unmarshal(mapJSON, &unmarshalledMap)
	if err != nil {
		t.Fatal(err)
	}

	if len(originalMap) != len(unmarshalledMap) {
		t.Fatalf("wrong map lengths")
	}
	for originalID, num := range originalMap {
		if unmarshalledMap[originalID] != num {
			t.Fatalf("map was incorrectly Unmarshalled")
		}
	}
}

func TestShortIDsToStrings(t *testing.T) {
	shortIDs := []ShortID{{1}, {2}, {2}}
	expected := []string{"6HgC8KRBEhXYbF4riJyJFLSHt37UNuRt", "BaMPFdqMUQ46BV8iRcwbVfsam55kMqcp", "BaMPFdqMUQ46BV8iRcwbVfsam55kMqcp"}
	shortStrings := ShortIDsToStrings(shortIDs)
	require.EqualValues(t, expected, shortStrings)
}

func TestShortIDLess(t *testing.T) {
	require := require.New(t)

	id1 := ShortID{}
	id2 := ShortID{}
	require.False(id1.Less(id2))
	require.False(id2.Less(id1))

	id1 = ShortID{1}
	id2 = ShortID{}
	require.False(id1.Less(id2))
	require.True(id2.Less(id1))

	id1 = ShortID{1}
	id2 = ShortID{1}
	require.False(id1.Less(id2))
	require.False(id2.Less(id1))

	id1 = ShortID{1}
	id2 = ShortID{1, 2}
	require.True(id1.Less(id2))
	require.False(id2.Less(id1))
}
