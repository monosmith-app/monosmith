package plugintrust

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreKeepsDecisionsPerWorkspaceAndPlugin(t *testing.T) {
	store := &Store{Path: filepath.Join(t.TempDir(), "home", "trust.json")}
	ws1, ws2 := t.TempDir(), t.TempDir()
	f := fp([]string{"x"}, "aaaa")

	if _, ok, err := store.Find(ws1, "w"); ok || err != nil {
		t.Fatalf("empty store: %v %v", ok, err)
	}
	if err := store.Trust(ws1, "w", f); err != nil {
		t.Fatal(err)
	}
	if r, ok, _ := store.Find(ws1, "w"); !ok || r.Executables[0].SHA256 != "aaaa" {
		t.Fatalf("not found: %+v", r)
	}
	if _, ok, _ := store.Find(ws2, "w"); ok {
		t.Fatal("trust leaked to another workspace")
	}
	if _, ok, _ := store.Find(ws1, "other"); ok {
		t.Fatal("trust leaked to another plugin")
	}

	// Trusting again replaces; untrusting removes only that one.
	if err := store.Trust(ws1, "w", fp([]string{"x"}, "bbbb")); err != nil {
		t.Fatal(err)
	}
	if err := store.Trust(ws2, "w", f); err != nil {
		t.Fatal(err)
	}
	if r, _, _ := store.Find(ws1, "w"); r.Executables[0].SHA256 != "bbbb" {
		t.Fatalf("not replaced: %+v", r)
	}
	if removed, err := store.Untrust(ws1, "w"); !removed || err != nil {
		t.Fatalf("untrust: %v %v", removed, err)
	}
	if removed, _ := store.Untrust(ws1, "w"); removed {
		t.Fatal("untrusted twice")
	}
	if _, ok, _ := store.Find(ws2, "w"); !ok {
		t.Fatal("untrust removed another workspace's decision")
	}
}

func TestStoreRefusesAFileItCannotRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trust.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := (&Store{Path: path}).Find("/w", "x"); err == nil {
		t.Fatal("a broken file read as empty")
	}
}
