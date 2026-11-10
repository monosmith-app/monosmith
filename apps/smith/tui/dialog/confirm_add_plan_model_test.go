package dialog

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// addPlan is the plan of the classic Add of a React app, as the beta.22 review
// showed it.
var addPlan = []string{
	"Plan: add app web",
	"Project web",
	"create apps/web/ (the stack decides the files: it comes from the recipe, the choices or the language)",
	"Workspace",
	"change smith.toml (the stack of the project is saved)",
	"Generated files smith removes",
	"remove tools/smith/install.sh (removed only if the project uses node, python or .NET: the pipelines then install smith with its package manager; the empty folder goes with it)",
	"Generated files smith may rewrite",
	"change .github/workflows/ci.yml (regenerated after the add; edits outside smith:slot markers are lost (smith ci generate --dry-run shows the diff))",
	"change .devcontainer/devcontainer.json (regenerated)",
	"Dry run: nothing was changed.",
}

// TestAShortCommandWithLongPathsMakesAWideDialogThatNeverBreaksAPath keeps the
// undo dialog readable: its command is short, its paths are not.
func TestAShortCommandWithLongPathsMakesAWideDialogThatNeverBreaksAPath(t *testing.T) {
	path := "libs/second/greeting/greet_use_case_test.go"
	c := NewConfirm("Undo the last change?", "smith reverts it with a new commit.", []string{"undo", "--yes"}).
		WithDetails([]string{"remove " + path}).Place(110, 32)
	var lines []string
	for _, l := range c.Lines(theme.Default()) {
		lines = append(lines, ansi.Strip(l))
	}
	if !strings.Contains(strings.Join(lines, "\n"), path) {
		t.Fatalf("the path is broken across lines:\n%s", strings.Join(lines, "\n"))
	}
}

// TestALongCommandIsNeverCutAndHasOnePrompt checks that every word of a command
// that wraps stays on screen, and that only its first line starts with "$".
func TestALongCommandIsNeverCutAndHasOnePrompt(t *testing.T) {
	argv := []string{"add", "app", "web", "--lang", "node", "--with", "react", "--environments", "lcl,dev,uat,prd",
		"--default-environment", "prd", "--environment-output", "workspace", "--env-files", "commit"}
	for _, size := range [][2]int{{110, 32}, {80, 24}, {72, 20}} {
		c := NewConfirm("Add web?", "Nothing is written yet.", argv).WithDetails(addPlan).Place(size[0], size[1])
		var lines []string
		for _, l := range c.Lines(theme.Default()) {
			lines = append(lines, strings.TrimSpace(strings.Trim(strings.TrimSpace(ansi.Strip(l)), "│")))
		}
		text := strings.Join(lines, "\n")
		for _, word := range argv {
			if !strings.Contains(text, word) {
				t.Fatalf("%v: the word %q is cut from the command:\n%s", size, word, text)
			}
		}
		if n := strings.Count(text, "$ "); n != 1 {
			t.Fatalf("%v: %d lines start with a prompt, want 1:\n%s", size, n, text)
		}
	}
}

func TestTheAddConfirmKeepsTheCommandAndTheButtonsAndTheCommitBoxInFull(t *testing.T) {
	why := "Nothing is written yet. It is built from your choices: react. Run makes the project and Esc leaves everything as it is. Tick the box to commit it, so it can be undone."
	argv := []string{"add", "app", "web", "--lang", "node", "--with", "react,eslint,vitest,tailwind,zod,something-long"}
	for _, size := range [][2]int{{110, 32}, {80, 24}} {
		c := NewConfirm("Add web?", why, argv).WithDetails(addPlan).Place(size[0], size[1])
		if r := c.Rect(); r.H > size[1] || r.W > size[0] {
			t.Fatalf("%v: the dialog is %dx%d", size, r.W, r.H)
		}
		lines := make([]string, 0)
		for _, l := range c.Lines(theme.Default()) {
			lines = append(lines, ansi.Strip(l))
		}
		text := strings.Join(lines, "\n")
		flat := strings.Join(strings.Fields(strings.ReplaceAll(text, "│", " ")), " ")
		for _, want := range []string{"Add web?", CommitLabel, "The command that will run:", "something-long", "Run (enter)", "Cancel (esc)"} {
			if !strings.Contains(flat, want) {
				t.Fatalf("%v: the dialog lost %q:\n%s", size, want, text)
			}
		}
		if size[1] == 24 && !strings.Contains(flat, "more lines") {
			t.Fatalf("%v: a plan that does not fit says how much is left: %s", size, text)
		}
	}
}
