package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissingWorkspaceAdvisesFromWhatTheFolderHolds(t *testing.T) {
	cases := []struct {
		file string
		want string
	}{
		{"nx.json", "smith convert --dry-run"},
		{"turbo.json", "smith convert --dry-run"},
		{"package.json", "smith adopt --dry-run"},
		{"go.mod", "smith adopt --dry-run"},
		{"", "smith new <name>"},
	}
	for _, c := range cases {
		dir := t.TempDir()
		if c.file != "" {
			if err := os.WriteFile(filepath.Join(dir, c.file), []byte("{}"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		t.Setenv("SMITH_WORKSPACE_ROOT", "")
		_, err := FindRoot(dir)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("%q: not an ErrNotFound: %v", c.file, err)
		}
		if !strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), "→") {
			t.Errorf("%q: advice lacks %q:\n%v", c.file, c.want, err)
		}
		if c.file != "" && strings.Contains(err.Error(), "smith new") {
			t.Errorf("%q: an existing repository must not be told to run smith new:\n%v", c.file, err)
		}
	}
}
