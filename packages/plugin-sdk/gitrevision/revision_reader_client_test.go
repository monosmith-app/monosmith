package gitrevision

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestReaderReadsManyFilesWithOneProcess(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	repo := t.TempDir()
	root := filepath.Join(repo, "ws") // the workspace is below the top of the repository
	if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "index.ts"), []byte("export const a = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init", "-q"}, {"add", "-A"},
		{"-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "init"},
		{"tag", "v1"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	src := NewReader(root, "v1")
	defer func() { _ = src.Close() }()
	if body, ok := src.Read("pkg/index.ts"); !ok || body != "export const a = 1\n" {
		t.Fatalf("read = %q, %v", body, ok)
	}
	if _, ok := src.Read("pkg/missing.ts"); ok {
		t.Fatal("a file the revision lacks must read as missing")
	}
	if body, ok := src.Read("pkg/index.ts"); !ok || body == "" {
		t.Fatal("a second read on the same process must still work")
	}
}
