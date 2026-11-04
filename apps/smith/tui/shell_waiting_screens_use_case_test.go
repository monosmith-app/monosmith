package tui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// slowRead makes every captured command of the lab wait until it is stopped or
// released, and reports each stop.
func (l *lab) slowRead(answer sources.Captured) (release chan struct{}, stopped chan struct{}) {
	release, stopped = make(chan struct{}), make(chan struct{}, 4)
	sh := l.sh()
	sh.sources.Capture = func(ctx context.Context, argv []string) sources.Captured {
		select {
		case <-ctx.Done():
			stopped <- struct{}{}
			return sources.Captured{Text: "interrupted"}
		case <-release:
			return answer
		}
	}
	l.m = sh
	return release, stopped
}

// wasStopped waits for the stop of a read.
func wasStopped(t *testing.T, stopped chan struct{}) {
	t.Helper()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("the work was not stopped")
	}
}

// The screens that read first open at once with a plain sentence, and Esc
// stops the read, leaves the person where they were and drops the late answer.
func TestTheScreensThatReadFirstOpenAtOnceAndEscStopsThem(t *testing.T) {
	for _, tc := range []struct {
		name    string
		open    func(*Shell) tea.Cmd
		purpose string
		words   []string
		argv    []string
	}{
		{"add choices", (*Shell).openAddProject, capChoices, []string{"Add a project", "Reading the choices your plugins offer"}, []string{"choices", "--json"}},
		{"swap choices", (*Shell).openSwap, capChoices, []string{"Swap a choice", "Reading the choices your plugins offer"}, []string{"choices", "--json"}},
		{"recipe levels", (*Shell).openRecipeForm, capRecipeLevels, []string{"Write a recipe", "Reading the plugins a recipe can fix"}, []string{"recipe", "new", "--list-levels", "--json"}},
		{"plugin search", func(s *Shell) tea.Cmd { return s.openPluginSearch("") }, capSearch, []string{"Search the plugin marketplace", "Searching the plugin marketplace"}, []string{"plugin", "search", "--json"}},
		{"tools install check", func(s *Shell) tea.Cmd { return s.openTools([]string{"tools", "install"}) }, capTools, []string{"Install the missing tools", "Looking for the tools that are missing"}, []string{"tools", "install", "--dry-run", "--json"}},
		{"compose", (*Shell).openCompose, capCompose, []string{"Compose a stack", "Reading what your plugins offer"}, []string{"compose", "--json"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := newHomeLab(t, 140, 40)
			_, stopped := l.slowRead(sources.Captured{OK: true, Text: "[]"})
			sh := l.sh()
			cmd := tc.open(&sh)
			l.m = sh
			l.drive(cmd)
			l.has(tc.words...)
			l.keys("esc")
			wasStopped(t, stopped)
			l.hasNot(tc.words[1])
			if got := l.sh().tab; got != commands.TabHome {
				t.Fatalf("Esc must leave the person on Home, not %q", got)
			}
			// The answer that was on its way opens nothing.
			l.send(capturedMsg{purpose: tc.purpose, argv: tc.argv, res: sources.Captured{OK: true, Text: "[]"}, ticket: l.sh().planTicket})
			l.hasNot(tc.words[1], "What are you building?")
			if l.sh().choices == choicesAsked || l.sh().formPending != "" {
				t.Fatal("Esc must forget the request for the choices, so the next press asks again")
			}
		})
	}
}

// The choices that arrive replace the working dialog with the form.
func TestTheChoicesFillTheWaitingDialogAndOpenTheForm(t *testing.T) {
	l := newHomeLab(t, 140, 40)
	release, _ := l.slowRead(sources.Captured{OK: true, Text: "[]"})
	sh := l.sh()
	cmd := sh.openAddProject()
	l.m = sh
	l.drive(nil)
	l.has("Reading the choices your plugins offer")
	close(release)
	l.drive(cmd)
	l.hasNot("Reading the choices your plugins offer")
	l.has("What are you building?")
}

// A search opens its dialog at once, and Esc leaves the person on the tab they were on.
func TestThePluginSearchOpensWorkingAndKeepsThePlaceOnCancel(t *testing.T) {
	l := newLab(t, true, 140, 40).openTab(commands.TabSettings)
	_, stopped := l.slowRead(sources.Captured{OK: true, Text: "[]"})
	sh := l.sh()
	cmd := sh.openPluginSearch("lint")
	l.m = sh
	l.drive(cmd)
	l.has("Searching the plugin marketplace", "smith plugin search lint --json")
	l.keys("esc")
	wasStopped(t, stopped)
	if got := l.sh().tab; got != commands.TabSettings {
		t.Fatalf("Esc must leave the person on Settings, not %q", got)
	}
}
