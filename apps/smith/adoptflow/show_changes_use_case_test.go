package adoptflow

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
)

func TestTheConversionsFilesShowAsTheDoctorFixDialogShowsAFix(t *testing.T) {
	root, scratch := t.TempDir(), t.TempDir()
	write := func(dir, rel, text string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(root, "package.json", "{\n  \"name\": \"acme\",\n  \"devDependencies\": {\n    \"turbo\": \"2\"\n  }\n}\n")
	write(scratch, "package.json", "{\n  \"name\": \"acme\"\n}\n")
	write(scratch, "smith.toml", "[workspace]\nname = \"acme\"\n")
	write(root, "turbo.json", "{}\n")
	write(scratch, "big.json", strings.Repeat("x\n", maxCreatedDiffBytes))
	write(scratch, "logo.bin", "a\x00b")

	c := Changes{Created: []string{"smith.toml", "big.json", "logo.bin"}, Edited: []string{"package.json"}, Removed: []string{"turbo.json"}}
	fix, left := changesAsFix(scratch, c)
	var b strings.Builder
	if err := doctorfix.ShowFix(&b, root, fix); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	for _, want := range []string{
		"● The files smith writes: 3 created, 1 edited, 1 removed",
		"+++ b/smith.toml", "+name = \"acme\"",
		"--- a/package.json", "-    \"turbo\": \"2\"",
		"--- a/turbo.json",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q:\n%s", want, got)
		}
	}
	if len(left) != 2 || !strings.Contains(left[0], "big.json") || !strings.Contains(left[0], "too big") || !strings.Contains(left[1], "logo.bin: not text") {
		t.Fatalf("left out: %v", left)
	}
	if bytes.Contains([]byte(got), []byte("big.json")) {
		t.Fatalf("a file too big to preview was diffed:\n%s", got[:200])
	}
}

func TestShowingChangesNeedsThePreparedCopy(t *testing.T) {
	if _, err := (&Session{}).ShowChanges(Changes{}); err != ErrNotConvertible {
		t.Fatalf("err %v", err)
	}
}
