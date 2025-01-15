package indexdb

import (
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
			name:        "no floor value",
			searchKey:   []byte("block_05"),
			wantErr:     true,
			errContains: "no floor value found",
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
