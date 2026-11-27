package rundurations

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestTheHistoryRoundTripsThroughTheLocalCacheFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".smith", "cache")
	if got := Load(dir); len(got) != 0 {
		t.Fatalf("no file is no history: %v", got)
	}
	want := History{"web:build": {1200, 1300}}
	if err := Save(dir, want); err != nil {
		t.Fatal(err)
	}
	if got := Load(dir); !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 || entries[0].Name() != FileName {
		t.Fatalf("one file, no temporary one left: %v", entries)
	}
}

func TestAFileThatIsNotJSONIsNoHistory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(Path(dir), []byte("{nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Load(dir); len(got) != 0 {
		t.Fatalf("%v", got)
	}
	if Save("", History{}) == nil {
		t.Fatal("no folder must say so")
	}
}
