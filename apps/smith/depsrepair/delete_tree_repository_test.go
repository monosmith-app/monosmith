package depsrepair

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// link makes a folder link at at that points to target: a junction on Windows
// (what npm makes for a workspace package), a symlink elsewhere.
func link(t *testing.T, target, at string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		if out, err := exec.Command("cmd", "/c", "mklink", "/J", at, target).CombinedOutput(); err != nil {
			t.Fatalf("mklink /J: %v: %s", err, out)
		}
		return
	}
	if err := os.Symlink(target, at); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A workspace package linked into node_modules keeps every file when
// node_modules is deleted: the link goes, its target stays.
func TestDeleteTreeDeletesALinkAsALinkAndKeepsItsTarget(t *testing.T) {
	root := t.TempDir()
	lib := filepath.Join(root, "libs", "ui")
	write(t, filepath.Join(lib, "src", "index.ts"), "export const ui = 1\n")
	write(t, filepath.Join(root, "node_modules", "left-pad", "index.js"), "module.exports = 1\n")
	if err := os.MkdirAll(filepath.Join(root, "node_modules", "@acme"), 0o755); err != nil {
		t.Fatal(err)
	}
	link(t, lib, filepath.Join(root, "node_modules", "@acme", "ui"))
	if _, err := os.Stat(filepath.Join(root, "node_modules", "@acme", "ui", "src", "index.ts")); err != nil {
		t.Fatalf("the link does not reach its target: %v", err)
	}

	if err := DeleteTree(context.Background(), filepath.Join(root, "node_modules")); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat(filepath.Join(root, "node_modules")); !os.IsNotExist(err) {
		t.Fatalf("node_modules is still there: %v", err)
	}
	if b, err := os.ReadFile(filepath.Join(lib, "src", "index.ts")); err != nil || string(b) != "export const ui = 1\n" {
		t.Fatalf("the target of the link lost its file: %q, %v", b, err)
	}
}

// The path itself may be a link: it goes, and its target stays.
func TestDeleteTreeOnALinkRemovesOnlyTheLink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "store")
	write(t, filepath.Join(target, "pkg", "a.txt"), "a")
	at := filepath.Join(root, "node_modules")
	link(t, target, at)

	if err := DeleteTree(context.Background(), at); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(at); !os.IsNotExist(err) {
		t.Fatalf("the link is still there: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "pkg", "a.txt")); err != nil {
		t.Fatalf("the target lost its file: %v", err)
	}
}

func TestDeleteTreeDeletesAFileAndIgnoresWhatIsNotThere(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "package-lock.json")
	write(t, file, "{}")
	if err := DeleteTree(context.Background(), file); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(file); !os.IsNotExist(err) {
		t.Fatalf("the file is still there: %v", err)
	}
	if err := DeleteTree(context.Background(), filepath.Join(root, "nothing")); err != nil {
		t.Fatalf("nothing to delete is not an error: %v", err)
	}
}

func TestDeleteTreeStopsWhenCancelled(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "node_modules", "a", "b.js"), "b")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := DeleteTree(ctx, filepath.Join(root, "node_modules")); err == nil {
		t.Fatal("a cancelled deletion says so")
	}
}
