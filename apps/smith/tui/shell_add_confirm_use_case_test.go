package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// realAddPlan is the output of `smith add app web --lang node --with react
// --dry-run` of the beta.23 binary, line for line (the indentation too).
const realAddPlan = `Plan: add app web
  Project web
    create  apps/web/  (the stack decides the files: it comes from the recipe, the choices or the language)
  Workspace
    change  smith.toml  (the stack of the project is saved)
  Generated files smith removes
    remove  tools/smith/install.sh  (removed only if the project uses node, python or .NET: the pipelines then install smith with its package manager; the empty folder goes with it)
  Generated files smith may rewrite
    change  .github/workflows/ci.yml  (regenerated after the add; edits outside smith:slot markers are lost (smith ci generate --dry-run shows the diff))
    change  .devcontainer/devcontainer.json  (regenerated)
Dry run: nothing was changed.
`

// bigAddPlan is a plan with owners and notes, as a workspace that already holds
// projects of its own kind shows it.
const bigAddPlan = realAddPlan + `  Follows the other projects of its kind
    owner   web follows api (node): the same test runner, linter and formatter
    note    the stack is taken from the sibling, not from the defaults of smith.toml
    note    the CI pipelines are regenerated for the new project
    note    run smith install after the add to fetch the packages of the new project
`

// TestTheAddConfirmShowsTheCommandInTheRealShell feeds the real plan into the
// confirmation of the shell and checks the frame at the sizes of the review.
func TestTheAddConfirmShowsTheCommandInTheRealShell(t *testing.T) {
	argv := []string{"add", "app", "web", "--lang", "node", "--with", "react"}
	for _, plan := range []string{realAddPlan, bigAddPlan} {
		for _, size := range [][2]int{{110, 32}, {100, 30}, {120, 30}, {80, 24}} {
			l := newLab(t, true, size[0], size[1])
			sh := l.sh()
			sh.addPlanRead(capturedMsg{argv: append(append([]string(nil), argv...), "--dry-run"), res: sources.Captured{OK: true, Text: plan}})
			l.m = sh
			l.has("The command that will run:", "$ smith add app web --lang node --with react", "Run (enter)", "Commit this change")
			r := l.sh().confirm.Rect()
			lay := l.sh().lay()
			if r.Y < lay.Main.Y || r.Y+r.H > lay.Footer.Y {
				t.Fatalf("%v: the dialog (rows %d-%d) covers the tab bar or the status bar (the frame is rows %d-%d)", size, r.Y, r.Y+r.H, lay.Main.Y, lay.Footer.Y)
			}
			if strings.Count(plan, "\n") > 11 && size[1] == 24 {
				l.has("more lines")
			}
		}
	}
}

// TestTheAddPlanDoesNotEndWithTheDryRunLine checks that the dialog leaves out
// the words "dry run": it already says that nothing is written yet.
func TestTheAddPlanDoesNotEndWithTheDryRunLine(t *testing.T) {
	l := newLab(t, true, 110, 32)
	sh := l.sh()
	sh.addPlanRead(capturedMsg{argv: []string{"add", "app", "web", "--lang", "node", "--dry-run"}, res: sources.Captured{OK: true, Text: realAddPlan}})
	l.m = sh
	l.has("The command that will run:")
	l.hasNot("Dry run")
}

// TestTheAddConfirmFollowsAResize keeps the command on screen when the window
// shrinks while the dialog is open.
func TestTheAddConfirmFollowsAResize(t *testing.T) {
	l := newLab(t, true, 120, 40)
	sh := l.sh()
	sh.addPlanRead(capturedMsg{argv: []string{"add", "app", "web", "--lang", "node", "--with", "react", "--dry-run"}, res: sources.Captured{OK: true, Text: bigAddPlan}})
	l.m = sh
	l.has("The command that will run:")
	l.send(tea.WindowSizeMsg{Width: 100, Height: 24})
	l.has("The command that will run:", "$ smith add app web", "Run (enter)")
}
