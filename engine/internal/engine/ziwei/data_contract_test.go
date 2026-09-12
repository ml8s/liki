package ziwei

import (
	"os"
	"testing"
)

func TestZiweiDataDirectoryContainsOnlyLoadedTables(t *testing.T) {
	entries, err := os.ReadDir("data")
	if err != nil {
		t.Fatal(err)
	}
	loaded := map[string]bool{
		"miao_wang.json": true,
		"tables.json":    true,
	}
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("unexpected directory data/%s", entry.Name())
		}
		if !loaded[entry.Name()] {
			t.Fatalf("data/%s is not embedded or read by the ziwei engine", entry.Name())
		}
	}
	if len(entries) != len(loaded) {
		t.Fatalf("data directory has %d files, want %d", len(entries), len(loaded))
	}
}
