package taskrun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLookPathPrefersTheChildPath(t *testing.T) {
	dir := t.TempDir()
	name := "smith-probe-tool"
	file := filepath.Join(dir, name)
	if isWindows {
		file += ".cmd"
	}
	if err := os.WriteFile(file, []byte("@echo off\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := lookPath(name, []string{"PATH=" + dir + string(os.PathListSeparator) + os.Getenv("PATH")})
	if err != nil || !strings.EqualFold(got, file) {
		t.Fatalf("got %q, %v; want %q", got, err, file)
	}
	if _, err := lookPath(name, []string{"PATH=" + t.TempDir()}); err == nil {
		t.Fatal("a program on no PATH is not found")
	}
}

func TestVenvBin(t *testing.T) {
	root := t.TempDir()
	if VenvBin(root) != "" {
		t.Fatal("no .venv, no bin")
	}
	bin := filepath.Join(root, ".venv", "bin")
	if isWindows {
		bin = filepath.Join(root, ".venv", "Scripts")
	}
	_ = os.MkdirAll(bin, 0o755)
	if VenvBin(root) != bin {
		t.Fatalf("got %q", VenvBin(root))
	}
}
