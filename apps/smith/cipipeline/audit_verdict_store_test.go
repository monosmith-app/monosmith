package cipipeline

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestMain lets the test binary stand in for an audit tool: it appends a line
// to the file named by AUDIT_HELPER_LOG and exits with AUDIT_HELPER_EXIT.
func TestMain(m *testing.M) {
	if log := os.Getenv("AUDIT_HELPER_LOG"); log != "" {
		f, _ := os.OpenFile(log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		_, _ = f.WriteString("ran\n")
		_ = f.Close()
		code, _ := strconv.Atoi(os.Getenv("AUDIT_HELPER_EXIT"))
		os.Exit(code)
	}
	os.Exit(m.Run())
}

// auditFixture is a workspace with a lockfile and an audit step that logs
// each time it really runs.
type auditFixture struct {
	root     string
	log      string
	step     protocol.Step
	verdicts *verdicts
	day      time.Time
}

func newAuditFixture(t *testing.T) *auditFixture {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package-lock.json"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	self, _ := os.Executable()
	f := &auditFixture{root: root, log: filepath.Join(t.TempDir(), "runs"), day: time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC)}
	t.Setenv("AUDIT_HELPER_LOG", f.log)
	t.Setenv("AUDIT_HELPER_EXIT", "0")
	f.step = protocol.Step{Name: "npm audit", Command: []string{self}, Blocking: true, CacheFiles: []string{"package-lock.json"}}
	store, err := taskcache.NewLocalStorage(filepath.Join(root, ".cache"))
	if err != nil {
		t.Fatal(err)
	}
	f.verdicts = &verdicts{store: store, root: root, version: "test", files: taskcache.NewFileDigests(root)}
	real := verdictClock
	verdictClock = func() time.Time { return f.day }
	t.Cleanup(func() { verdictClock = real })
	return f
}

// run audits once and says how often the tool has really run, in total.
func (f *auditFixture) run(t *testing.T) (runs int, out string, err error) {
	t.Helper()
	f.verdicts.files = taskcache.NewFileDigests(f.root)
	var buf bytes.Buffer
	err = runSteps(context.Background(), f.root, []protocol.Step{f.step}, f.verdicts, &buf)
	raw, _ := os.ReadFile(f.log)
	return strings.Count(string(raw), "ran"), buf.String(), err
}

func TestAnUnchangedLockfileReplaysThePassedVerdict(t *testing.T) {
	f := newAuditFixture(t)
	if runs, _, err := f.run(t); runs != 1 || err != nil {
		t.Fatalf("first audit: %d runs, %v", runs, err)
	}
	runs, out, err := f.run(t)
	if runs != 1 || err != nil || !strings.Contains(out, "no change in package-lock.json since it passed on 2026-10-18") {
		t.Errorf("second audit replays: %d runs, %v\n%s", runs, err, out)
	}
}

func TestAChangedLockfileIsAuditedAgain(t *testing.T) {
	f := newAuditFixture(t)
	if _, _, err := f.run(t); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.root, "package-lock.json"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if runs, _, err := f.run(t); runs != 2 || err != nil {
		t.Errorf("a changed lockfile re-audits: %d runs, %v", runs, err)
	}
}

func TestAVerdictIsNotReplayedAfterTheDatabaseDayMovesOn(t *testing.T) {
	f := newAuditFixture(t)
	if _, _, err := f.run(t); err != nil {
		t.Fatal(err)
	}
	f.day = f.day.Add(24 * time.Hour)
	if runs, _, err := f.run(t); runs != 2 || err != nil {
		t.Errorf("a new day re-audits: %d runs, %v", runs, err)
	}
	if runs, _, _ := f.run(t); runs != 2 {
		t.Errorf("and is remembered again: %d runs", runs)
	}
}

func TestAFailingAuditIsNeverRemembered(t *testing.T) {
	f := newAuditFixture(t)
	t.Setenv("AUDIT_HELPER_EXIT", "1")
	if runs, _, err := f.run(t); runs != 1 || err == nil {
		t.Fatalf("a blocking failure fails: %d runs, %v", runs, err)
	}
	if runs, _, err := f.run(t); runs != 2 || err == nil {
		t.Errorf("and runs again: %d runs, %v", runs, err)
	}
	f.step.Blocking = false
	if runs, _, err := f.run(t); runs != 3 || err != nil {
		t.Errorf("a warning is not remembered either: %d runs, %v", runs, err)
	}
	if runs, _, _ := f.run(t); runs != 4 {
		t.Errorf("%d runs", runs)
	}
}

func TestAStepWithoutAnExistingFileIsNeverCached(t *testing.T) {
	f := newAuditFixture(t)
	f.step.CacheFiles = []string{"missing.lock"}
	if _, _, err := f.run(t); err != nil {
		t.Fatal(err)
	}
	if runs, _, _ := f.run(t); runs != 2 {
		t.Errorf("nothing says the verdict holds: %d runs", runs)
	}
	f.step.CacheFiles = nil
	if runs, _, _ := f.run(t); runs != 3 {
		t.Errorf("no declared files: %d runs", runs)
	}
}

func TestAnotherStepOrDirectoryHasItsOwnVerdict(t *testing.T) {
	f := newAuditFixture(t)
	if _, _, err := f.run(t); err != nil {
		t.Fatal(err)
	}
	f.step.Dir = "libs/core"
	if err := os.MkdirAll(filepath.Join(f.root, "libs", "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	if runs, _, _ := f.run(t); runs != 2 {
		t.Errorf("another directory audits: %d runs", runs)
	}
}
