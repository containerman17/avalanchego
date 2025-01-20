package indexdb

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestIndexDB_GetFloorValue(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "indexdb_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary file path within the temporary directory
	dbPath := filepath.Join(tmpDir, "test_index.db")

	// Create new IndexDB
	db, err := NewIndexDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create IndexDB: %v", err)
	}
	defer db.impl.Close()

	// Test data
	keys := [][]byte{
		[]byte("block_10"),
		[]byte("block_30"),
		[]byte("block_50"),
		[]byte("block_70"),
	}
	values := []uint32{10, 30, 50, 70}

	// Insert test data
	if err := db.Put(keys, values); err != nil {
		t.Fatalf("Failed to put test data: %v", err)
	}

	tests := []struct {
		name        string
		searchKey   []byte
		wantValue   uint32
		wantErr     bool
		errContains string
	}{
		{
			name:      "exact match",
			searchKey: []byte("block_30"),
			wantValue: 30,
			wantErr:   false,
		},
		{
			name:      "floor value - between blocks",
			searchKey: []byte("block_35"),
			wantValue: 30,
			wantErr:   false,
		},
		{
			name:      "floor value - at start",
			searchKey: []byte("block_15"),
			wantValue: 10,
			wantErr:   false,
		},
		{
			name:      "before first - return 0",
			searchKey: []byte("block_05"),
			wantValue: 0,
			wantErr:   false,
		},
		{
			name:      "last block",
			searchKey: []byte("block_80"),
			wantValue: 70,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetFloorValue(tt.searchKey)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetFloorValue() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("GetFloorValue() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("GetFloorValue() unexpected error = %v", err)
				return
			}
			if got != tt.wantValue {
				t.Errorf("GetFloorValue() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestIndexDB_GetNextKeyValue(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "indexdb_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary file path
	dbPath := filepath.Join(tmpDir, "test_index.db")

	// Create new IndexDB
	db, err := NewIndexDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create IndexDB: %v", err)
	}
	defer db.impl.Close()

	// Test data
	keys := [][]byte{
		[]byte("block_10"),
		[]byte("block_30"),
		[]byte("block_50"),
		[]byte("block_70"),
	}
	values := []uint32{10, 30, 50, 70}

	// Insert test data
	if err := db.Put(keys, values); err != nil {
		t.Fatalf("Failed to put test data: %v", err)
	}

	tests := []struct {
		name          string
		searchKey     []byte
		wantNextKey   []byte
		wantNextValue uint32
		wantNil       bool // true if we expect nil key/value
	}{
		{
			name:          "exact match - get next",
			searchKey:     []byte("block_30"),
			wantNextKey:   []byte("block_50"),
			wantNextValue: 50,
		},
		{
			name:          "between blocks - get next",
			searchKey:     []byte("block_20"),
			wantNextKey:   []byte("block_30"),
			wantNextValue: 30,
		},
		{
			name:          "before first - get first",
			searchKey:     []byte("block_05"),
			wantNextKey:   []byte("block_10"),
			wantNextValue: 10,
		},
		{
			name:      "after last - no next",
			searchKey: []byte("block_80"),
			wantNil:   true,
		},
		{
			name:      "at last - no next",
			searchKey: []byte("block_70"),
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotValue, err := db.GetNextKeyValue(tt.searchKey)
			if err != nil {
				t.Errorf("GetNextKeyValue() unexpected error = %v", err)
				return
			}

			if tt.wantNil {
				if gotKey != nil || gotValue != 0 {
					t.Errorf("GetNextKeyValue() = (%v, %v), want nil, 0", gotKey, gotValue)
				}
				return
			}

			if !bytes.Equal(gotKey, tt.wantNextKey) {
				t.Errorf("GetNextKeyValue() key = %s, want %s", gotKey, tt.wantNextKey)
			}
			if gotValue != tt.wantNextValue {
				t.Errorf("GetNextKeyValue() value = %v, want %v", gotValue, tt.wantNextValue)
			}
		})
	}
}
