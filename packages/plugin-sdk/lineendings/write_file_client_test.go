package lineendings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileKeepsTheEndingOfTheFile(t *testing.T) {
	dir := t.TempDir()
	for name, c := range map[string]struct{ before, after, want string }{
		"crlf.json": {"{\r\n  \"a\": 1\r\n}\r\n", "{\n  \"a\": 2\n}\n", "{\r\n  \"a\": 2\r\n}\r\n"},
		"lf.json":   {"{\n  \"a\": 1\n}\n", "{\r\n  \"a\": 2\r\n}\r\n", "{\n  \"a\": 2\n}\n"},
		"new.json":  {"", "{\n}\n", "{\n}\n"},
	} {
		path := filepath.Join(dir, name)
		if c.before != "" {
			if err := os.WriteFile(path, []byte(c.before), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if err := WriteFile(path, []byte(c.after), 0o644); err != nil {
			t.Fatal(err)
		}
		if got, _ := os.ReadFile(path); string(got) != c.want {
			t.Errorf("%s = %q, want %q", name, got, c.want)
		}
	}
}
