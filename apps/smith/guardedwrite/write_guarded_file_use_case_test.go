package guardedwrite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.html")
	var out strings.Builder

	// A new file is written.
	if err := WriteFile(path, []byte("one\n"), false, "--force", &out); err != nil {
		t.Fatal(err)
	}
	// The same content is not a change.
	if err := WriteFile(path, []byte("one\n"), false, "--force", &out); err != nil {
		t.Fatal(err)
	}
	// Other content is refused: the file, the diff and the flag are named.
	err := WriteFile(path, []byte("two\n"), false, "--force", &out)
	if err == nil {
		t.Fatal("want a refusal")
	}
	for _, want := range []string{"report.html", "-one", "+two", "--force"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal lacks %q:\n%v", want, err)
		}
	}
	if raw, _ := os.ReadFile(path); string(raw) != "one\n" {
		t.Fatalf("a refusal must not write: %q", raw)
	}
	// The flag replaces it and prints what was lost.
	if err := WriteFile(path, []byte("two\n"), true, "--force", &out); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(path); string(raw) != "two\n" || !strings.Contains(out.String(), "-one") {
		t.Fatalf("%q\n%s", raw, out.String())
	}
}
