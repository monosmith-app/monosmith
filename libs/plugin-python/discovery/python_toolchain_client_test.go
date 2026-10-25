package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
)

func TestTheToolchainIsTheVenvsInterpreterWhenThereIsOne(t *testing.T) {
	root := t.TempDir()
	if got := venvPython(root); got != "" {
		t.Fatalf("no venv yet, got %q", got)
	}
	want := filepath.FromSlash(strings.ReplaceAll(venv.Interpreter(), "{workspaceRoot}", filepath.ToSlash(root)))
	if err := os.MkdirAll(filepath.Dir(want), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(want, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := venvPython(root); got != want {
		t.Fatalf("venvPython = %q, want %q", got, want)
	}
}
