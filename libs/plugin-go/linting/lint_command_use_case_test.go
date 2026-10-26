package linting

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWithoutAConfigSmithPassesItsDefault(t *testing.T) {
	ws := t.TempDir()
	lint, err := Command(ws, "libs/core")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(lint.Command, " "); got != "golangci-lint run --allow-serial-runners --config {workspaceRoot}/.smith/defaults/golangci.yml" {
		t.Fatalf("command: %s", got)
	}
	written, err := os.ReadFile(filepath.Join(ws, ".smith", "defaults", "golangci.yml"))
	if err != nil || string(written) != DefaultConfig() {
		t.Fatalf("the default is materialised: %q %v", written, err)
	}
}

func TestTheRepositorysConfigWinsAtAnyLevel(t *testing.T) {
	for _, where := range []string{".golangci.yml", "libs/core/.golangci.toml"} {
		ws := t.TempDir()
		full := filepath.Join(ws, filepath.FromSlash(where))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("version: \"2\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		lint, err := Command(ws, "libs/core")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.Join(lint.Command, " "), "--config") {
			t.Errorf("%s: smith must not override the repository's config: %v", where, lint.Command)
		}
		if _, err := os.Stat(filepath.Join(ws, ".smith")); err == nil {
			t.Errorf("%s: no default is written when the repository has its own", where)
		}
	}
}
