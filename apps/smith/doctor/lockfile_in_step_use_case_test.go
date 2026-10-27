package doctor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func lockfileRoot(t *testing.T, manifest string, lock bool) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if lock {
		if err := os.WriteFile(filepath.Join(root, "package-lock.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func fakeDryRun(output string, ran bool, err error) func(string) (string, bool, error) {
	return func(string) (string, bool, error) { return output, ran, err }
}

func TestALockfileNpmCiRefusesIsAWarningWithAnInstallFix(t *testing.T) {
	root := lockfileRoot(t, `{"workspaces":["packages/*"]}`, true)
	got := checkLockfile(root, fakeDryRun("npm error Invalid Version: \nnpm error A complete log\n", true, errors.New("exit 1")))
	if len(got) != 1 || got[0].OK || !got[0].Warning || got[0].Check != lockfileCheck {
		t.Fatalf("%+v", got)
	}
	f := got[0]
	if !strings.Contains(f.Detail, "Invalid Version") || !strings.Contains(f.Remedy, "run npm install, then commit package-lock.json") {
		t.Fatalf("%+v", f)
	}
	if f.Fix == nil || f.Fix.Safe || len(f.Fix.Steps) != 3 || f.Fix.Steps[0].Argv[1] != "install" || f.Fix.Steps[2].Argv[1] != "ci" {
		t.Fatalf("the fix must run npm install and be asked about: %+v", f.Fix)
	}
}

func TestALockfileNpmCiAcceptsPasses(t *testing.T) {
	root := lockfileRoot(t, `{"workspaces":{"packages":["a"]}}`, true)
	got := checkLockfile(root, fakeDryRun("", true, nil))
	if len(got) != 1 || !got[0].OK || got[0].Fix != nil {
		t.Fatalf("%+v", got)
	}
}

func TestNoFindingWithoutWorkspacesLockfileOrAnswer(t *testing.T) {
	dry := func(string) (string, bool, error) { t.Fatal("npm must not run"); return "", false, nil }
	if got := checkLockfile(lockfileRoot(t, `{"name":"x"}`, true), dry); got != nil {
		t.Errorf("not a workspace: %+v", got)
	}
	if got := checkLockfile(lockfileRoot(t, `{"workspaces":["a"]}`, false), dry); got != nil {
		t.Errorf("no lockfile: %+v", got)
	}
	if got := checkLockfile(lockfileRoot(t, `{"workspaces":["a"]}`, true), fakeDryRun("", false, nil)); got != nil {
		t.Errorf("npm gave no answer: %+v", got)
	}
}
