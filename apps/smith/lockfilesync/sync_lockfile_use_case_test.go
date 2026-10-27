package lockfilesync

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeNpm answers each npm call from a script and records the calls; a call
// past the script succeeds. onInstall runs when an install is asked for.
type fakeNpm struct {
	results   []error
	calls     []string
	onInstall func()
}

func (f *fakeNpm) run(_ string, args ...string) (string, error) {
	f.calls = append(f.calls, args[0])
	if args[0] == "install" && f.onInstall != nil {
		f.onInstall()
	}
	if len(f.results) == 0 {
		return "", nil
	}
	err := f.results[0]
	f.results = f.results[1:]
	if err != nil {
		return "npm error Invalid Version:\n", err
	}
	return "", nil
}

func workspaceWith(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range append([]string{"package.json"}, files...) {
		if err := os.WriteFile(filepath.Join(root, name), []byte("{\"name\":\"x\"}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

var errNpm = errors.New("exit status 1")

func TestAnInstallThatPassesTheCheckIsQuietAndLeavesTheLockfile(t *testing.T) {
	root := workspaceWith(t, "package-lock.json")
	npm := &fakeNpm{}
	var out bytes.Buffer
	res, err := Sync(Options{Root: root, Out: &out, Run: npm.run})
	if err != nil || res.Regenerated || out.Len() != 0 || strings.Join(npm.calls, ",") != "install,ci" {
		t.Fatalf("%v %+v %q %v", err, res, out.String(), npm.calls)
	}
}

func TestAFailedCheckIsRetriedOnceWithAnotherInstall(t *testing.T) {
	root := workspaceWith(t, "package-lock.json")
	npm := &fakeNpm{results: []error{nil, errNpm, nil, nil}}
	res, err := Sync(Options{Root: root, Out: &bytes.Buffer{}, Run: npm.run})
	if err != nil || res.Regenerated || strings.Join(npm.calls, ",") != "install,ci,install,ci" {
		t.Fatalf("%v %+v %v", err, res, npm.calls)
	}
}

func TestATrackedLockfileThatStaysBrokenIsRebuiltAndTheUserIsTold(t *testing.T) {
	root := workspaceWith(t, "package-lock.json")
	lock := filepath.Join(root, "package-lock.json")
	installs := 0
	npm := &fakeNpm{
		// install, ci fails, install, ci fails, then the rebuild: install, ci
		results: []error{nil, errNpm, nil, errNpm, nil, nil},
		onInstall: func() {
			installs++
			if installs == 3 {
				if _, err := os.Stat(lock); err == nil {
					t.Error("the lockfile should have been deleted before the rebuild")
				}
				_ = os.WriteFile(lock, []byte("new"), 0o644)
			}
		},
	}
	var out bytes.Buffer
	res, err := Sync(Options{Root: root, Out: &out, Run: npm.run, Tracked: func(string) bool { return true }})
	if err != nil || !res.Regenerated {
		t.Fatalf("%v %+v", err, res)
	}
	if text := out.String(); !strings.Contains(text, "versions may have moved") && !strings.Contains(text, "Package versions may have moved") || !strings.Contains(text, "git diff package-lock.json") {
		t.Fatalf("the user must be told: %q", text)
	}
}

func TestAnUntrackedLockfileIsNeverDeletedAndTheRemedyIsPrinted(t *testing.T) {
	root := workspaceWith(t, "package-lock.json")
	npm := &fakeNpm{results: []error{nil, errNpm, nil, errNpm}}
	var out bytes.Buffer
	_, err := Sync(Options{Root: root, Out: &out, Run: npm.run, Tracked: func(string) bool { return false }})
	if err == nil || !strings.Contains(err.Error(), "npm install") {
		t.Fatalf("want a remedy, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "package-lock.json")); statErr != nil {
		t.Fatal("the lockfile was deleted")
	}
	if !strings.Contains(out.String(), "Invalid Version") {
		t.Fatalf("npm's output must be shown when a step fails: %q", out.String())
	}
}

func TestARebuildThatFailsPutsTheOriginalLockfileBack(t *testing.T) {
	root := workspaceWith(t, "package-lock.json")
	npm := &fakeNpm{results: []error{nil, errNpm, nil, errNpm, errNpm}}
	_, err := Sync(Options{Root: root, Out: &bytes.Buffer{}, Run: npm.run, Tracked: func(string) bool { return true }})
	if err == nil {
		t.Fatal("want an error")
	}
	raw, readErr := os.ReadFile(filepath.Join(root, "package-lock.json"))
	if readErr != nil || !strings.Contains(string(raw), "\"name\"") {
		t.Fatalf("the original lockfile must be back: %v %q", readErr, raw)
	}
}

func TestWithoutALockfileNothingRunsUnlessCreateIsSetAndNpmIsTheManager(t *testing.T) {
	npm := &fakeNpm{}
	res, err := Sync(Options{Root: workspaceWith(t), Out: &bytes.Buffer{}, Run: npm.run})
	if err != nil || res.Skipped == "" || len(npm.calls) != 0 {
		t.Fatalf("%v %+v %v", err, res, npm.calls)
	}
	res, err = Sync(Options{Root: workspaceWith(t, "yarn.lock"), Out: &bytes.Buffer{}, Run: npm.run, Create: true})
	if err != nil || res.Skipped == "" || len(npm.calls) != 0 {
		t.Fatalf("a yarn workspace is not npm's: %v %+v %v", err, res, npm.calls)
	}
	res, err = Sync(Options{Root: workspaceWith(t), Out: &bytes.Buffer{}, Run: npm.run, Create: true})
	if err != nil || res.Skipped != "" || len(npm.calls) != 2 {
		t.Fatalf("%v %+v %v", err, res, npm.calls)
	}
}

func TestAFailureShowsTheTailOfTheLogFileNpmNames(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "debug.log")
	if err := os.WriteFile(logFile, []byte("line one\nthe real cause\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := failureTail("npm error boom\nnpm error A complete log of this run can be found in: " + logFile + "\n")
	if !strings.Contains(got, "the real cause") || !strings.Contains(got, "npm error boom") {
		t.Fatalf("%q", got)
	}
}
