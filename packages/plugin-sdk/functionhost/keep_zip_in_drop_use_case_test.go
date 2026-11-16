package functionhost

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeepZipInDropCopiesTheZipBesideTheAppSettings(t *testing.T) {
	root := t.TempDir()
	build := filepath.Join(root, "dist", "apps", "orders")
	if err := os.MkdirAll(build, 0o755); err != nil {
		t.Fatal(err)
	}
	zip := filepath.Join(build, "orders.zip")
	if err := os.WriteFile(zip, []byte("PK zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	rel, err := KeepZipInDrop(root, "orders", zip)
	if err != nil {
		t.Fatal(err)
	}
	if rel != "dist/drop/orders/orders.zip" {
		t.Fatalf("copy at %q, want dist/drop/orders/orders.zip", rel)
	}
	got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil || string(got) != "PK zip" {
		t.Fatalf("copy holds %q (%v), want the zip's bytes", got, err)
	}
	if _, err := os.Stat(zip); err != nil {
		t.Fatalf("the zip in the build folder is gone: %v", err)
	}
}
