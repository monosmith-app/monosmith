package scaffold

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestInspectFolderListsFilesAndSecrets(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"a.txt", ".env", "src/b.go", ".git/config", "node_modules/x/y.js"} {
		p := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := InspectFolder(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{".env", "a.txt", "node_modules/", "src/b.go"}; !reflect.DeepEqual(got.Files, want) {
		t.Errorf("files: got %v, want %v", got.Files, want)
	}
	if want := []string{".env"}; !reflect.DeepEqual(got.Secrets, want) {
		t.Errorf("secrets: got %v, want %v", got.Secrets, want)
	}
	missing, err := InspectFolder(filepath.Join(dir, "none"))
	if err != nil || len(missing.Files) != 0 {
		t.Errorf("missing folder: %v %v", missing, err)
	}
}
