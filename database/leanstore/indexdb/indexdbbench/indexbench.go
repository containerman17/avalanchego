package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ava-labs/avalanchego/database/leanstore/indexdb"
)

const (
	numItems  = 1_000_000
	readItems = numItems * 10
	keySize   = 96
	batchSize = 5000
)

func generateRandomKey() []byte {
	key := make([]byte, keySize)
	rand.Read(key)
	return key
}

func main() {
	fmt.Printf("Starting benchmark\n")
	// Create temporary directory for the database
	tmpDir, err := os.MkdirTemp("", "indexdb_bench")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "bench.db")
	db, err := indexdb.NewIndexDB(dbPath)
	if err != nil {
		fmt.Printf("Failed to create IndexDB: %v\n", err)
		return
	}
	defer db.Close()

	// Generate random keys
	keys := make([][]byte, numItems)
	values := make([]uint32, numItems)
	for i := range keys {
		keys[i] = generateRandomKey()
		values[i] = uint32(i)
	}

	// Benchmark insertion in batches
	start := time.Now()
	batchStart := start
	for i := 0; i < numItems; i += batchSize {
		end := i + batchSize
		if end > numItems {
			end = numItems
		}
		if err := db.Put(keys[i:end], values[i:end]); err != nil {
			fmt.Printf("Failed to insert batch %d-%d: %v\n", i, end, err)
			return
		}
		batchDuration := time.Since(batchStart)
		fmt.Printf("Inserted batch %d-%dK items in %v (%.2fK items/sec)\n",
			i/1000, end/1000, batchDuration, float64(end-i)/batchDuration.Seconds()/1000)
		batchStart = time.Now()
	}
	insertDuration := time.Since(start)
	fmt.Printf("Total: Inserted %dK items in %v (%.2fK items/sec)\n",
		numItems/1000, insertDuration, float64(numItems)/insertDuration.Seconds()/1000)

	// Verify all values with more reads
	start = time.Now()
	errors := 0
	for i := 0; i < readItems; i++ {
		// Pick a random key from our dataset
		idx := i % numItems
		_, val, err := db.GetFloorKeyValue(keys[idx])
		if err != nil {
			errors++
			continue
		}
		if val != values[idx] {
			errors++
			fmt.Printf("Value mismatch for key %d: expected %d, got %d\n", idx, values[idx], val)
		}
	}
	queryDuration := time.Since(start)
	fmt.Printf("Queried %dK items in %v (%.2fK items/sec) with %d errors\n",
		readItems/1000, queryDuration, float64(readItems)/queryDuration.Seconds()/1000, errors)

	// Test floor queries with non-existent keys
	start = time.Now()
	errors = 0
	for i := 0; i < readItems; i++ {
		randomKey := generateRandomKey()
		_, _, err := db.GetFloorKeyValue(randomKey)
		if err != nil {
			errors++
		}
	}
	floorDuration := time.Since(start)
	fmt.Printf("Performed %dK floor queries in %v (%.2fK queries/sec) with %d errors\n",
		readItems/1000, floorDuration, float64(readItems)/floorDuration.Seconds()/1000, errors)

	fmt.Printf("Done\n")
}
