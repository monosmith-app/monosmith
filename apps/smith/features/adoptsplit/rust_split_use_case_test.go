package adoptsplit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitRustMakesTheRootAWorkspace(t *testing.T) {
	root := t.TempDir()
	put(t, root, "apps/cli/Cargo.toml", "[package]\nname = \"cli\"\n")
	put(t, root, "apps/cli/Cargo.lock", "# lock\n")
	if _, err := SplitRust(Request{Root: root, Dir: "apps/cli"}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, root, "Cargo.toml"); got != "[workspace]\nresolver = \"2\"\nmembers = [\"apps/cli\"]\n" {
		t.Errorf("root Cargo.toml:\n%s", got)
	}
	if read(t, root, "Cargo.lock") != "# lock\n" {
		t.Error("Cargo.lock moves to the root")
	}
	if _, err := os.Stat(filepath.Join(root, "apps/cli/Cargo.lock")); err == nil {
		t.Error("the crate keeps a Cargo.lock")
	}
	if got := read(t, root, ".cargo/config.toml"); !strings.Contains(got, `target-dir = "dist/cargo"`) {
		t.Errorf(".cargo/config.toml:\n%s", got)
	}
}

func TestSplitRustKeepsWhatTheRootHas(t *testing.T) {
	root := t.TempDir()
	put(t, root, "apps/cli/Cargo.toml", "[package]\nname = \"cli\"\n")
	put(t, root, "Cargo.toml", "[workspace]\nmembers = []\n")
	put(t, root, ".cargo/config", "[build]\ntarget-dir = \"out\"\n")
	if _, err := SplitRust(Request{Root: root, Dir: "apps/cli"}); err != nil {
		t.Fatal(err)
	}
	if read(t, root, "Cargo.toml") != "[workspace]\nmembers = []\n" {
		t.Error("an existing root Cargo.toml is kept")
	}
	if _, err := os.Stat(filepath.Join(root, ".cargo/config.toml")); err == nil {
		t.Error("an existing .cargo/config is kept as the only one")
	}
}
