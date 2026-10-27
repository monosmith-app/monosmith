package linting

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
	args, own, err := Args(ws, "packages/core")
	if err != nil || own {
		t.Fatalf("%v %v", own, err)
	}
	if got := strings.Join(args, " "); got != "-c {workspaceRoot}/.smith/defaults/phpstan.neon" {
		t.Fatalf("args: %s", got)
	}
	written, err := os.ReadFile(filepath.Join(ws, ".smith", "defaults", "phpstan.neon"))
	if err != nil || string(written) != DefaultConfig() {
		t.Fatalf("the default is materialised: %q %v", written, err)
	}
}

func TestThePackagesPHPStanConfigWins(t *testing.T) {
	for _, name := range configNames {
		ws := t.TempDir()
		write(t, ws, "packages/core/"+name, "parameters:\n    level: 5\n")
		args, own, err := Args(ws, "packages/core")
		if err != nil || !own || len(args) != 0 {
			t.Errorf("%s: %v %v %v", name, args, own, err)
		}
		if _, err := os.Stat(filepath.Join(ws, ".smith")); err == nil {
			t.Errorf("%s: no default is written when the package has its own", name)
		}
	}
}

func TestAConfigAboveThePackageDoesNotCountBecausePHPStanDoesNotLookThere(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "phpstan.neon", "parameters:\n    level: 5\n")
	if _, own, _ := Args(ws, "packages/core"); own {
		t.Fatal("PHPStan reads only the working directory's file")
	}
}
