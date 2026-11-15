package prompt

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestChangedFilesListsModifiedAndUntrackedFiles(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	root := t.TempDir()
	git(t, root, "init", "-q")
	if err := os.MkdirAll(filepath.Join(root, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "api", "a.go"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "add", ".")
	git(t, root, "commit", "-q", "-m", "first")
	if got := ChangedFiles(root); len(got) != 0 {
		t.Fatalf("clean tree: got %v", got)
	}
	if err := os.WriteFile(filepath.Join(root, "api", "a.go"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "new.txt"), []byte("n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := ChangedFiles(root)
	slices.Sort(got)
	if want := []string{"api/a.go", "new.txt"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestChangedFilesOfAFolderInsideALargerRepositoryAreRelativeToTheFolder(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	top := t.TempDir()
	git(t, top, "init", "-q")
	ws := filepath.Join(top, "ws")
	if err := os.MkdirAll(filepath.Join(ws, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(top, "outside.txt"), []byte("o"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "api", "a.go"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := ChangedFiles(ws)
	if want := []string{"api/a.go"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
