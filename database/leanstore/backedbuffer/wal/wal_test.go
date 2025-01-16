package wal

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestWALWriteReplay(t *testing.T) {
	// Create temp file
	tmpfile := "test.wal"
	defer os.Remove(tmpfile)

	// Create new WAL with 4-byte keys and values
	wal, err := New(tmpfile)
	if err != nil {
		t.Fatal("Failed to create WAL")
	}

	// First write with simple 4-byte keys/values
	keys1 := [][]byte{
		[]byte("key1"),
		[]byte("key2"),
	}
	vals1 := [][]byte{
		[]byte("val1"),
		[]byte("val2"),
	}

	err = wal.Write(keys1, vals1)
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	err = wal.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	wal, err = New(tmpfile)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// ReadAll
	replayKeys, replayVals, err := wal.DebugGetAllItems()
	if err != nil {
		t.Fatalf("DebugGetAllItems failed: %v", err)
	}

	// Verify replay results
	if len(replayKeys) != len(keys1) {
		t.Fatalf("Expected %d keys, got %d", len(keys1), len(replayKeys))
	}

	for i, key := range keys1 {
		found := false
		for j, replayKey := range replayKeys {
			if bytes.Equal(key, replayKey) {
				if !bytes.Equal(vals1[i], replayVals[j]) {
					t.Errorf("Value mismatch for key %s: expected %s, got %s",
						string(key), string(vals1[i]), string(replayVals[j]))
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Key %s not found in replay", string(key))
		}
	}

	// Second write
	keys2 := [][]byte{
		[]byte("key3"),
		[]byte("key1"),
	}
	vals2 := [][]byte{
		[]byte("val3"),
		[]byte("new1"),
	}

	err = wal.Write(keys2, vals2)
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	// Second replay
	replayKeys2, replayVals2, err := wal.DebugGetAllItems()
	if err != nil {
		t.Fatalf("Second DebugGetAllItems failed: %v", err)
	}

	// Verify final state - all keys should be present, with key1 appearing twice
	expectedKeys := [][]byte{
		[]byte("key1"),
		[]byte("key2"),
		[]byte("key3"),
		[]byte("key1"),
	}
	expectedVals := [][]byte{
		[]byte("val1"),
		[]byte("val2"),
		[]byte("val3"),
		[]byte("new1"),
	}

	if len(replayKeys2) != len(expectedKeys) {
		t.Fatalf("Expected %d keys, got %d", len(expectedKeys), len(replayKeys2))
	}

	// Count occurrences of each key in replay results
	keyOccurrences := make(map[string]int)
	for _, key := range replayKeys2 {
		// Trim any null bytes when converting to string
		trimmedKey := string(bytes.TrimRight(key, "\x00"))
		keyOccurrences[trimmedKey]++
	}

	// Verify key1 appears twice and other keys once
	if keyOccurrences["key1"] != 2 {
		t.Errorf("Expected key1 to appear twice, got %d occurrences", keyOccurrences["key1"])
	}
	if keyOccurrences["key2"] != 1 {
		t.Errorf("Expected key2 to appear once, got %d occurrences", keyOccurrences["key2"])
	}
	if keyOccurrences["key3"] != 1 {
		t.Errorf("Expected key3 to appear once, got %d occurrences", keyOccurrences["key3"])
	}

	// Verify values appear in correct order
	for i, key := range expectedKeys {
		found := false
		for j, replayKey := range replayKeys2 {
			// Compare trimmed versions of the keys and values
			if bytes.Equal(bytes.TrimRight(key, "\x00"), bytes.TrimRight(replayKey, "\x00")) &&
				bytes.Equal(bytes.TrimRight(expectedVals[i], "\x00"), bytes.TrimRight(replayVals2[j], "\x00")) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Key-value pair (%s, %s) not found in expected order",
				string(bytes.TrimRight(key, "\x00")), string(bytes.TrimRight(expectedVals[i], "\x00")))
		}
	}
}

func TestWALVariableLength(t *testing.T) {
	tmpfile := "test_variable.wal"
	defer os.Remove(tmpfile)

	wal, err := New(tmpfile)
	if err != nil {
		t.Fatal("Failed to create WAL")
	}

	keys := [][]byte{
		[]byte("k1"),
		[]byte("key2"),
		[]byte("longkey3"),
	}
	values := [][]byte{
		[]byte("value1"),
		[]byte("v2"),
		[]byte("longvalue3"),
	}

	err = wal.Write(keys, values)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	replayKeys, replayVals, err := wal.DebugGetAllItems()
	if err != nil {
		t.Fatalf("DebugGetAllItems failed: %v", err)
	}

	if len(replayKeys) != len(keys) {
		t.Fatalf("Expected %d keys, got %d", len(keys), len(replayKeys))
	}

	for i := range keys {
		if !bytes.Equal(keys[i], replayKeys[i]) {
			t.Errorf("Key mismatch at index %d: expected %s, got %s",
				i, string(keys[i]), string(replayKeys[i]))
		}
		if !bytes.Equal(values[i], replayVals[i]) {
			t.Errorf("Value mismatch at index %d: expected %s, got %s",
				i, string(values[i]), string(replayVals[i]))
		}
	}
}

func TestWALReplayBatch(t *testing.T) {
	tmpfile := "test_batch.wal"
	defer os.Remove(tmpfile)

	wal, err := New(tmpfile)
	if err != nil {
		t.Fatal("Failed to create WAL")
	}

	// Create test data with more entries than the batch size
	numEntries := 2500
	keys := make([][]byte, numEntries)
	values := make([][]byte, numEntries)
	for i := 0; i < numEntries; i++ {
		keys[i] = []byte(fmt.Sprintf("key%d", i))
		values[i] = []byte(fmt.Sprintf("value%d", i))
	}

	// Write test data
	err = wal.Write(keys, values)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Create a test replayer that counts entries
	replayer := &testReplayer{
		t:               t,
		seenEntries:     0,
		expectedEntries: numEntries,
	}

	// Replay with small batch size to test batching
	err = wal.ReplayBatch(replayer, 100)
	if err != nil {
		t.Fatalf("ReplayBatch failed: %v", err)
	}

	if replayer.seenEntries != numEntries {
		t.Errorf("Expected %d entries, got %d", numEntries, replayer.seenEntries)
	}
}

type testReplayer struct {
	t               *testing.T
	seenEntries     int
	expectedEntries int
}

func (r *testReplayer) Set(keys [][]byte, values [][]byte) error {
	if len(keys) != len(values) {
		return fmt.Errorf("keys and values length mismatch")
	}

	for i := range keys {
		expectedKey := fmt.Sprintf("key%d", r.seenEntries)
		expectedValue := fmt.Sprintf("value%d", r.seenEntries)

		if !bytes.Equal(keys[i], []byte(expectedKey)) {
			r.t.Errorf("Key mismatch at index %d: expected %s, got %s",
				r.seenEntries, expectedKey, string(keys[i]))
		}
		if !bytes.Equal(values[i], []byte(expectedValue)) {
			r.t.Errorf("Value mismatch at index %d: expected %s, got %s",
				r.seenEntries, expectedValue, string(values[i]))
		}
		r.seenEntries++
	}
	return nil
}
