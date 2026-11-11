package discovery

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// crateDir writes files (crate-relative, with the content "") into a new crate
// folder and parses manifest as its Cargo.toml.
func crateDir(t *testing.T, manifest string, files ...string) (string, Crate) {
	t.Helper()
	dir := t.TempDir()
	for _, f := range files {
		p := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	c, err := ParseCargo([]byte(manifest))
	if err != nil {
		t.Fatal(err)
	}
	return dir, c
}

func TestCrateBinaries(t *testing.T) {
	cases := []struct {
		name, manifest string
		files          []string
		want           []string
	}{
		{"main.rs is the crate's program", "[package]\nname = \"api\"\n", []string{"src/main.rs"}, []string{"api"}},
		{"a library has none", "[package]\nname = \"core\"\n", []string{"src/lib.rs"}, nil},
		{"a [[bin]] renames main.rs", "[package]\nname = \"loomas-cli\"\n\n[[bin]]\nname = \"loomas\"\npath = \"src/main.rs\"\n", []string{"src/main.rs"}, []string{"loomas"}},
		{"a [[bin]] beside main.rs", "[package]\nname = \"tool\"\n\n[[bin]]\nname = \"helper\"\npath = \"src/helper.rs\"\n", []string{"src/main.rs", "src/helper.rs"}, []string{"helper", "tool"}},
		{"src/bin files", "[package]\nname = \"kit\"\n", []string{"src/main.rs", "src/bin/b.rs", "src/bin/a.rs", "src/bin/own/main.rs"}, []string{"kit", "a", "b"}},
		{"a [[bin]] with the name of a src/bin file", "[package]\nname = \"kit\"\n\n[[bin]]\nname = \"a\"\n", []string{"src/bin/a.rs"}, []string{"a"}},
		{"autobins off", "[package]\nname = \"kit\"\nautobins = false\n\n[[bin]]\nname = \"only\"\npath = \"src/only.rs\"\n", []string{"src/main.rs", "src/bin/a.rs"}, []string{"only"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir, crate := crateDir(t, c.manifest, c.files...)
			if got := CrateBinaries(dir, crate); !slices.Equal(got, c.want) {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}
