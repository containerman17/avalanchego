package backedbuffer

import (
	"fmt"

	"github.com/ava-labs/avalanchego/database/leanstore/backedbuffer/rbtree"
	"github.com/ava-labs/avalanchego/database/leanstore/backedbuffer/wal"
)

type BackedBuffer struct {
	memStore *rbtree.RBTreeStore
	walStore *wal.WAL
}

func New(walPath string) (*BackedBuffer, error) {
	walStore, err := wal.New(walPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create WAL: %w", err)
	}
	memStore := rbtree.NewRBTreeStore()

	err = walStore.Replay(memStore)
	if err != nil {
		return nil, fmt.Errorf("failed to read wal: %w", err)
	}

	return &BackedBuffer{
		memStore: memStore,
		walStore: walStore,
	}, nil
}

func (b *BackedBuffer) Put(keys [][]byte, values [][]byte) error {
	// Write to WAL first
	if err := b.walStore.Write(keys, values); err != nil {
		return fmt.Errorf("failed to write to WAL: %w", err)
	}

	// Then update in-memory store
	return b.memStore.Set(keys, values)
}

func (b *BackedBuffer) Get(key []byte) ([]byte, error) {
	// Read from in-memory store
	return b.memStore.Get(key)
}

func (b *BackedBuffer) Has(key []byte) (bool, error) {
	val, err := b.Get(key)
	return val != nil, err
}

func (b *BackedBuffer) Delete(key []byte) error {
	// Write to WAL first
	if err := b.walStore.Write([][]byte{key}, [][]byte{{}}); err != nil {
		return fmt.Errorf("failed to write to WAL: %w", err)
	}

	// Then update in-memory store
	return b.memStore.Set([][]byte{key}, [][]byte{{}})
}

func (b *BackedBuffer) Close() error {
	return b.walStore.Close()
}

func (b *BackedBuffer) GetNext(key []byte) ([]byte, []byte, error) {
	return b.memStore.GetNext(key)
}
