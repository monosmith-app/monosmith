package formatting

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, ws, rel, body string) {
	t.Helper()
	full := filepath.Join(ws, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWithoutAConfigSmithPassesItsDefault(t *testing.T) {
	ws := t.TempDir()
	args, own, err := Args(ws, "crates/core")
	if err != nil || own {
		t.Fatalf("%v %v", own, err)
	}
	if got := strings.Join(args, " "); got != "--config-path {workspaceRoot}/.smith/defaults/rustfmt.toml" {
		t.Fatalf("args: %s", got)
	}
	written, err := os.ReadFile(filepath.Join(ws, ".smith", "defaults", "rustfmt.toml"))
	if err != nil || string(written) != DefaultConfig() {
		t.Fatalf("the default is materialised: %q %v", written, err)
	}
}

func TestTheRepositorysRustfmtConfigWins(t *testing.T) {
	for _, where := range []string{"rustfmt.toml", ".rustfmt.toml", "crates/core/rustfmt.toml", "crates/.rustfmt.toml"} {
		ws := t.TempDir()
		write(t, ws, where, "tab_spaces = 2\n")
		args, own, err := Args(ws, "crates/core")
		if err != nil || !own || len(args) != 0 {
			t.Errorf("%s: %v %v %v", where, args, own, err)
		}
		if _, err := os.Stat(filepath.Join(ws, ".smith")); err == nil {
			t.Errorf("%s: no default is written when the repository has its own", where)
		}
	}
}
