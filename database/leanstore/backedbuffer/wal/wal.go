package wal

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	readBufferSize   = 1024 * 1024
	writeBufferSize  = 1024 * 1024
	maxKeySize       = 255     // 1 byte
	maxValueSize     = 1 << 16 // 2 bytes (64KB)
	flushInterval    = time.Second
	defaultBatchSize = 1000
)

type WAL struct {
	filename     string
	file         *os.File
	buffer       *bufio.Writer
	mu           sync.Mutex
	done         chan struct{} // Signal to stop background flushing
	lastFlush    time.Time
	flushPending bool
}

func New(filename string) (*WAL, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	w := &WAL{
		filename:  filename,
		file:      file,
		buffer:    bufio.NewWriterSize(file, writeBufferSize),
		done:      make(chan struct{}),
		lastFlush: time.Now(),
	}

	// Start background flush goroutine
	go w.periodicFlush()

	return w, nil
}

func (a *WAL) periodicFlush() {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.mu.Lock()
			if a.flushPending {
				if err := a.buffer.Flush(); err != nil {
					// Log error or handle it appropriately
					fmt.Printf("Error flushing buffer: %v\n", err)
				}
				a.flushPending = false
				a.lastFlush = time.Now()
			}
			a.mu.Unlock()
		case <-a.done:
			return
		}
	}
}

func (a *WAL) Write(keys, values [][]byte) error {
	if len(keys) != len(values) {
		return fmt.Errorf("keys and values length mismatch")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	for i := 0; i < len(keys); i++ {
		keyLen := len(keys[i])
		valueLen := len(values[i])

		if keyLen > maxKeySize {
			return fmt.Errorf("key too large: %d bytes (max %d)", keyLen, maxKeySize)
		}
		if valueLen > maxValueSize {
			return fmt.Errorf("value too large: %d bytes (max %d)", valueLen, maxValueSize)
		}

		// Write key length (1 byte)
		if err := a.buffer.WriteByte(byte(keyLen)); err != nil {
			return err
		}
		// Write value length (2 bytes)
		if err := binary.Write(a.buffer, binary.LittleEndian, uint16(valueLen)); err != nil {
			return err
		}
		// Write key
		if _, err := a.buffer.Write(keys[i]); err != nil {
			return err
		}
		// Write value
		if _, err := a.buffer.Write(values[i]); err != nil {
			return err
		}
	}

	a.flushPending = true
	return nil
}

type Replayable interface {
	Set(keys [][]byte, values [][]byte) error
}

func (a *WAL) ReplayBatch(target Replayable, batchSize int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.buffer.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	// Seek to start of file
	if _, err := a.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to start of file: %w", err)
	}

	reader := bufio.NewReaderSize(a.file, readBufferSize)
	keys := make([][]byte, 0, batchSize)
	values := make([][]byte, 0, batchSize)

	for {
		// Read key length (1 byte)
		keyLen, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read key length: %w", err)
		}

		// Read value length (2 bytes)
		var valueLen uint16
		if err := binary.Read(reader, binary.LittleEndian, &valueLen); err != nil {
			return fmt.Errorf("failed to read value length: %w", err)
		}

		// Read key
		key := make([]byte, keyLen)
		if _, err := io.ReadFull(reader, key); err != nil {
			return fmt.Errorf("failed to read key: %w", err)
		}

		// Read value
		value := make([]byte, valueLen)
		if _, err := io.ReadFull(reader, value); err != nil {
			return fmt.Errorf("failed to read value: %w", err)
		}

		keys = append(keys, key)
		values = append(values, value)

		// Process batch if we've reached batch size
		if len(keys) >= batchSize {
			if err := target.Set(keys, values); err != nil {
				return fmt.Errorf("failed to process batch: %w", err)
			}
			// Reset slices while preserving capacity
			keys = keys[:0]
			values = values[:0]
		}
	}

	// Process any remaining entries
	if len(keys) > 0 {
		if err := target.Set(keys, values); err != nil {
			return fmt.Errorf("failed to process final batch: %w", err)
		}
	}

	// Seek back to end for future writes
	if _, err := a.file.Seek(0, 2); err != nil {
		return fmt.Errorf("failed to seek to end of file: %w", err)
	}

	return nil
}

func (a *WAL) Replay(target Replayable) error {
	return a.ReplayBatch(target, defaultBatchSize)
}

func (a *WAL) DebugGetAllItems() ([][]byte, [][]byte, error) {
	collector := &debugCollector{}
	err := a.Replay(collector)
	if err != nil {
		return nil, nil, err
	}
	return collector.keys, collector.values, nil
}

// debugCollector is a helper type used only in tests to collect all WAL entries in memory
type debugCollector struct {
	keys   [][]byte
	values [][]byte
}

func (r *debugCollector) Set(keys [][]byte, values [][]byte) error {
	r.keys = append(r.keys, keys...)
	r.values = append(r.values, values...)
	return nil
}

func (a *WAL) Close() error {
	// Signal background goroutine to stop
	close(a.done)

	// Final flush
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.buffer.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}
	return a.file.Close()
}

// Ensure immediate flush
func (a *WAL) Sync() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.buffer.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}
	a.flushPending = false
	a.lastFlush = time.Now()
	return nil
}

// Deletes the current file and writes a new one
func (a *WAL) Flush() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Close current file
	if err := a.file.Close(); err != nil {
		return err
	}

	// Remove the file
	if err := os.Remove(a.filename); err != nil {
		return err
	}

	// Create new file
	file, err := os.OpenFile(a.filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	a.file = file
	return nil
}

func (a *WAL) DropCache() error {
	// Ensure all pending writes are synced to disk
	if err := a.buffer.Flush(); err != nil {
		return err
	}

	// On Linux, write "1" to /proc/sys/vm/drop_caches
	// This requires root privileges
	return os.WriteFile("/proc/sys/vm/drop_caches", []byte("1"), 0644)
}
