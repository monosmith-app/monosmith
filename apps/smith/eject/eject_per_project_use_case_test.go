package eject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

func TestAPerProjectToolIsCopiedIntoEachProjectThatUsesTheDefault(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".smith/defaults/phpstan.neon", "parameters:\n    level: 8\n")
	write(t, root, "packages/a/src/x.php", "<?php\n")
	write(t, root, "packages/b/src/x.php", "<?php\n")
	targets, err := Eject(root, tooldefaults.PHPStan, false, []string{"packages/a", "packages/b"})
	if err != nil || strings.Join(targets, " ") != "packages/a/phpstan.neon packages/b/phpstan.neon" {
		t.Fatalf("%v %v", targets, err)
	}
	if got := read(t, root, "packages/b/phpstan.neon"); !strings.Contains(got, "level: 8") {
		t.Fatalf("copied %q", got)
	}
	if _, err := Eject(root, tooldefaults.PHPStan, false, nil); err == nil {
		t.Fatal("no project uses the default: nothing to copy")
	}
}

func TestARefusalWritesNothing(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".smith/defaults/phpstan.neon", "default\n")
	write(t, root, "packages/b/phpstan.neon", "mine\n")
	if _, err := Eject(root, tooldefaults.PHPStan, false, []string{"packages/a", "packages/b"}); err == nil {
		t.Fatal("expected a refusal")
	}
	if _, err := os.Stat(filepath.Join(root, "packages", "a", "phpstan.neon")); err == nil {
		t.Fatal("packages/a was written although packages/b refused")
	}
}
