package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// ask runs a command line the way a menu entry does: unconfirmed.
func (l *lab) ask(argv ...string) {
	sh := l.sh()
	cmd := sh.execute(argv, false)
	l.m = sh
	l.drive(cmd)
}

// A command that changes the workspace and can print its plan shows the plan
// in the dialog, then runs with --yes only after the answer.
func TestAChangingCommandShowsItsPlanBeforeItRuns(t *testing.T) {
	for _, tc := range []struct {
		argv      []string
		plan      string
		planLine  string
		runs      string
		wantTexts []string
	}{
		{[]string{"upgrade"}, "upgrade --dry-run", "change  smith.toml (defaults vintage)", "upgrade --yes", []string{"change smith.toml"}},
		{[]string{"ci", "generate"}, "ci generate --dry-run", "remove  .gitlab-ci.yml\n-stages: [build]", "ci generate --yes", []string{"remove .gitlab-ci.yml", "-stages: [build]"}},
		{[]string{"bootstrap"}, "bootstrap --dry-run", "change  tools/smith/smith", "bootstrap --yes", []string{"tools/smith/smith"}},
		{[]string{"sync"}, "sync --dry-run", "axios (node) will be ^1.6.0", "sync --yes", []string{"axios (node) will be ^1.6.0"}},
	} {
		t.Run(strings.Join(tc.argv, " "), func(t *testing.T) {
			l := newHomeLab(t, 140, 40)
			l.answers(map[string]sources.Captured{tc.plan: {OK: true, Text: tc.planLine}})
			l.ask(tc.argv...)
			l.has(tc.wantTexts...)
			l.has("smith " + tc.runs)
			if len(l.started) != 0 {
				t.Fatalf("nothing may run before the answer: %v", l.started)
			}
			l.keys("enter")
			if got := l.lastStarted(); got != tc.runs {
				t.Fatalf("started %q, want %q", got, tc.runs)
			}
		})
	}
}

// A plan that smith refuses is shown as it is, and nothing runs.
func TestAPlanSmithRefusesIsShownAndNothingRuns(t *testing.T) {
	l := newHomeLab(t, 140, 40)
	l.answers(map[string]sources.Captured{"kind core lib --dry-run": {Text: "✖ core is imported by an app\n  → --force goes ahead"}})
	l.ask("kind", "core", "lib")
	l.has("core is imported by an app", "--force goes ahead")
	l.keys("enter")
	if len(l.started) != 0 {
		t.Fatalf("a refused plan runs nothing: %v", l.started)
	}
}

// A release asked before the Release view was opened reads the plan itself and
// shows the packages and versions in the confirmation.
func TestReleaseReadsItsPlanBeforeAsking(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.ask("release")
	l.has("Packages and versions to publish:", "web 1.0.0 → 1.1.0")
	if len(l.started) != 0 {
		t.Fatalf("nothing may run before the answer: %v", l.started)
	}
}
