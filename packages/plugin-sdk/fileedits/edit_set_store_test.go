package fileedits

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditsKeepTheLineEndingOfAnEditedFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "ci.yml"), []byte("a: 1\r\nb: 2\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	set := New(root)
	set.Write("ci.yml", "a: 1\nb: 3\n")
	edits := set.Edits()
	if len(edits) != 1 || edits[0].Content != "a: 1\r\nb: 3\r\n" {
		t.Fatalf("got %+v", edits)
	}
}

func TestEditsDropAFileWhoseOnlyChangeWasTheEnding(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "ci.yml"), []byte("a: 1\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	set := New(root)
	set.Write("ci.yml", "a: 1\n")
	if edits := set.Edits(); len(edits) != 0 {
		t.Fatalf("got %+v", edits)
	}
}
