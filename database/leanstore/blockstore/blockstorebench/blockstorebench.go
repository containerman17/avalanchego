package main

import (
	"crypto/rand"
	"fmt"
	mathRand "math/rand/v2"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/database/leanstore/blockstore"
)

const (
	numOperations = 50000
	numThreads    = 32
	opsPerThread  = numOperations / numThreads
)

func generateRandomBlock(size int) []byte {
	block := make([]byte, size)
	rand.Read(block)
	return block
}

func runBenchmark(name string, store blockstore.IBlockPersister, blockSize int) {
	fmt.Printf("Running benchmark for %s with %d threads\n", name, numThreads)
	start := time.Now()

	// Insert phase
	blockIDs := make([]uint32, numOperations)
	insertStart := time.Now()
	var wg sync.WaitGroup

	for t := 0; t < numThreads; t++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			for i := 0; i < opsPerThread; i++ {
				block := generateRandomBlock(blockSize)
				id, err := store.Insert(block)
				if err != nil {
					fmt.Printf("Insert error: %v\n", err)
					return
				}
				blockIDs[threadID*opsPerThread+i] = id
			}
		}(t)
	}
	wg.Wait()
	insertDuration := time.Since(insertStart)
	fmt.Printf("Insert time: %v (%.1fK ops/sec)\n",
		insertDuration,
		float64(numOperations)/insertDuration.Seconds()/1000)

	// Random reads
	readStart := time.Now()
	wg = sync.WaitGroup{}
	for t := 0; t < numThreads; t++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerThread; i++ {
				id := blockIDs[(i+mathRand.IntN(len(blockIDs)))%len(blockIDs)]
				_, err := store.Get(id)
				if err != nil {
					fmt.Printf("Read error: %v\n", err)
					return
				}
			}
		}()
	}
	wg.Wait()
	readDuration := time.Since(readStart)
	fmt.Printf("Read time: %v (%.1fK ops/sec)\n",
		readDuration,
		float64(numOperations)/readDuration.Seconds()/1000)

	// Random updates
	updateStart := time.Now()
	wg = sync.WaitGroup{}
	for t := 0; t < numThreads; t++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerThread; i++ {
				id := blockIDs[(i+mathRand.IntN(len(blockIDs)))%len(blockIDs)]
				block := generateRandomBlock(blockSize)
				err := store.Update(id, block)
				if err != nil {
					fmt.Printf("Update error: %v\n", err)
					return
				}
			}
		}()
	}
	wg.Wait()
	updateDuration := time.Since(updateStart)
	fmt.Printf("Update time: %v (%.1fK ops/sec)\n",
		updateDuration,
		float64(numOperations)/updateDuration.Seconds()/1000)

	total := time.Since(start)
	fmt.Printf("Total time: %v\n\n", total)
}

func runBlockBenchmark(sizeName, fileName string, blockSize int, tmpDir string) error {
	store, err := blockstore.CreateRegularBlockStore(filepath.Join(tmpDir, fileName), blockSize)
	if err != nil {
		return fmt.Errorf("failed to create %s store: %w", sizeName, err)
	}
	defer store.Close()
	runBenchmark(sizeName, store, blockSize)
	return nil
}

type benchConfig struct {
	name      string
	fileName  string
	blockSize int
}

func main() {
	fmt.Println("---")
	tmpDir, err := os.MkdirTemp("", "blockstore_bench_*")
	if err != nil {
		fmt.Printf("Failed to create temp directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	configs := []benchConfig{
		// {"2KB Blocks", "blocks_2kb.db", 2 * 1024},
		// {"4KB Blocks", "blocks_4kb.db", 4 * 1024},
		// {"8KB Blocks", "blocks_8kb.db", 8 * 1024},
		{"16KB Blocks", "blocks_16kb.db", 16 * 1024},
		// {"32KB Blocks", "blocks_32kb.db", 32 * 1024},
	}

	for _, cfg := range configs {
		fmt.Printf("Testing %s\n", cfg.name)
		err := runBlockBenchmark(cfg.name, cfg.fileName, cfg.blockSize, tmpDir)
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			return
		}
	}
}
