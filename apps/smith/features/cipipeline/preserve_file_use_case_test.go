package cipipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreserveFileKeepsCRLF(t *testing.T) {
	root := t.TempDir()
	old := "name: ci\r\non: push\r\n"
	if err := os.WriteFile(filepath.Join(root, "ci.yml"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	got, lost := preserveFile(root, "ci.yml", old, "name: ci\non: push\njobs: {}\n")
	if want := "name: ci\r\non: push\r\njobs: {}\r\n"; got != want || len(lost) != 0 {
		t.Fatalf("got %q (lost %v), want %q", got, lost, want)
	}
}

func TestPreserveFileLeavesAnUpToDateCRLFFileAlone(t *testing.T) {
	root := t.TempDir()
	old := "name: ci\r\non: push\r\n"
	if err := os.WriteFile(filepath.Join(root, "ci.yml"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, _ := preserveFile(root, "ci.yml", old, "name: ci\non: push\n"); got != old {
		t.Fatalf("got %q, want the file unchanged", got)
	}
}
