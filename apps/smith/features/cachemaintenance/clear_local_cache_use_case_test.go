package cachemaintenance

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClearDeletesTheStoreAndOnlyTheStore(t *testing.T) {
	store, dir := newStore(t)
	seedEntry(t, store, "web:build", []byte("out"))
	seedEntry(t, store, "api:test", []byte("out2"))
	// Whatever else lives in the configured directory is not smith's to delete.
	other := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(other, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Clear(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Entries != 2 || result.Blobs != 4 || result.Bytes == 0 || result.Dir != dir {
		t.Fatalf("result %+v", result)
	}
	if _, err := os.Stat(store.Root()); err == nil {
		t.Fatal("the store must be gone")
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatal("a file that is not smith's must stay")
	}
	status, err := Inspect(dir, 0)
	if err != nil || status.Exists {
		t.Fatalf("after clearing: %+v (%v)", status, err)
	}
}

func TestClearOfACacheThatDoesNotExistIsNotAnError(t *testing.T) {
	result, err := Clear(filepath.Join(t.TempDir(), "never"))
	if err != nil || result.Entries != 0 || result.Bytes != 0 {
		t.Fatalf("result %+v, err %v", result, err)
	}
}

func TestACacheCanBeUsedAgainAfterItIsCleared(t *testing.T) {
	store, dir := newStore(t)
	seedEntry(t, store, "web:build", []byte("out"))
	if _, err := Clear(dir); err != nil {
		t.Fatal(err)
	}
	fresh, err := newStoreAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	seedEntry(t, fresh, "web:build", []byte("out"))
	if status, _ := Inspect(dir, 0); status.Entries != 1 {
		t.Fatalf("status %+v", status)
	}
}
