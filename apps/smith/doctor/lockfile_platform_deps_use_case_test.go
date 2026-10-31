package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLock(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package-lock.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestALockfileMissingPlatformPackagesIsAWarningWithARegenerateFix(t *testing.T) {
	root := writeLock(t, `{"packages":{"":{},"node_modules/a":{"optionalDependencies":{"@emnapi/core":"1"}}}}`)
	got := checkLockfilePlatformDeps(root)
	if len(got) != 1 || got[0].OK || !got[0].Warning || got[0].Check != lockfilePlatformCheck || !strings.Contains(got[0].Detail, "@emnapi/core") {
		t.Fatalf("%+v", got)
	}
	if got[0].Fix == nil || got[0].Fix.Safe || len(got[0].Fix.Steps) != 2 || got[0].Fix.Steps[1].Argv[2] != "--package-lock-only" {
		t.Fatalf("%+v", got[0].Fix)
	}
}

func TestACompleteLockfilePassesAndNoLockfileSaysNothing(t *testing.T) {
	got := checkLockfilePlatformDeps(writeLock(t, `{"packages":{"":{},"node_modules/a":{}}}`))
	if len(got) != 1 || !got[0].OK || got[0].Fix != nil {
		t.Fatalf("%+v", got)
	}
	if got := checkLockfilePlatformDeps(t.TempDir()); got != nil {
		t.Fatalf("%+v", got)
	}
}
