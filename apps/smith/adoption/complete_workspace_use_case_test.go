package adoption

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/lockfilesync"
)

const engineWarning = "npm warn EBADENGINE Unsupported engine {\nnpm warn EBADENGINE   package: 'a@1.0.0',\nnpm warn EBADENGINE   required: { node: '>=22' }\nnpm warn EBADENGINE }\n"

func npmRoot(t *testing.T, withLock bool) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if withLock {
		if err := os.WriteFile(filepath.Join(root, "package-lock.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestAnInstallThatWorkedFoldsNpmWarningsIntoOneLine(t *testing.T) {
	var out bytes.Buffer
	run := func(_ string, args ...string) (string, error) { return engineWarning, nil }
	settleLockfile(lockfilesync.Options{Root: npmRoot(t, true), Out: &out, Create: true, Run: run}, true, &out)
	text := out.String()
	if strings.Contains(text, "EBADENGINE") || !strings.Contains(text, "1 package wants a newer Node") || !strings.Contains(text, "  npm install done\n") {
		t.Fatalf("%s", text)
	}
}

func TestAnInstallThatFailedSaysSoAndLeavesTheAdoptionDone(t *testing.T) {
	var out bytes.Buffer
	run := func(_ string, args ...string) (string, error) { return engineWarning, errors.New("exit 1") }
	settleLockfile(lockfilesync.Options{Root: npmRoot(t, true), Out: &out, Create: true, Run: run, Tracked: func(string) bool { return false }}, true, &out)
	if text := out.String(); !strings.Contains(text, "EBADENGINE") || !strings.Contains(text, "not done") {
		t.Fatalf("%s", text)
	}
}

func TestWithoutInstallOnlyAnExistingLockfileIsUpdatedAndWithoutTouchingNodeModules(t *testing.T) {
	var out bytes.Buffer
	var calls [][]string
	run := func(_ string, args ...string) (string, error) { calls = append(calls, args); return "", nil }
	settleLockfile(lockfilesync.Options{Root: npmRoot(t, false), Out: &out, LockOnly: true, Run: run}, true, &out)
	if len(calls) != 0 || out.Len() != 0 {
		t.Fatalf("no lockfile, nothing to do: %v %q", calls, out.String())
	}
	settleLockfile(lockfilesync.Options{Root: npmRoot(t, true), Out: &out, LockOnly: true, Run: run}, true, &out)
	if len(calls) != 2 || !strings.Contains(strings.Join(calls[0], " "), "--package-lock-only") || !strings.Contains(out.String(), "package-lock.json updated") {
		t.Fatalf("%v %q", calls, out.String())
	}
}

func TestSettleLockfileDoesNothingWhenPackageJsonDidNotChange(t *testing.T) {
	var out bytes.Buffer
	run := func(string, ...string) (string, error) { t.Fatal("npm ran"); return "", nil }
	settleLockfile(lockfilesync.Options{Root: npmRoot(t, true), Out: &out, Run: run}, false, &out)
	if out.Len() != 0 {
		t.Fatal(out.String())
	}
}

func TestNoGoWorkIsWrittenForProjectsThatHaveNoGoMod(t *testing.T) {
	root := t.TempDir()
	if err := writeGoWork(root, []string{"apps/mvd", "libs/core"}, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.work")); err == nil {
		t.Error("projects declared in the root module have no go.mod to use: a go.work with nothing in it breaks every go command")
	}
}
