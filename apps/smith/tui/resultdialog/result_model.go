package resultdialog

import (
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
)

// Model is the dialog: a title, what happened, and a list of choices. It runs
// nothing: a key that decides returns an Outcome for the host to carry out.
type Model struct {
	title   string
	facts   []Fact
	choices []choice
	cursor  int
	// foot is the sentence under the list that tells what is safe.
	foot string
	rect dialog.Rect
	// working, tick and elapsed draw the moving line of a command that runs.
	working string
	tick    int
	elapsed time.Duration
}

// NewAfterChange builds the dialog for a change that finished and was not
// committed. The first choice, the default, leaves the files as they are;
// committing is the second, so smith never commits unasked.
//
// Parameters: argv is the smith command line that ran; lines is its output.
// Returns: the dialog.
func NewAfterChange(argv, lines []string) Model {
	op := OperationText(argv)
	m := Model{title: "Done: " + op}
	created, isAdd := ParseCreated(argv, lines)
	if isAdd {
		m.title = "Added " + created.Kind + " " + created.Name
		where := ""
		if created.Dir != "" {
			where = " in " + created.Dir
		}
		count := ""
		if created.Files > 0 {
			count = " (" + Plural(created.Files, "file") + ")"
		}
		m.facts = append(m.facts, Fact{Good, created.Name + " was created" + where + count + "."})
		if created.Kind == "lib" {
			m.facts = append(m.facts, Fact{Info, "Use it from an app: press + on the app and pick " + created.Name + ", or run smith install -w <app> " + created.Name + "."})
		}
	} else {
		m.facts = append(m.facts, Fact{Good, "The change is written to your files."})
	}
	m.facts = append(m.facts, Fact{Info, "It is not committed: git shows these files as new changes."})
	m.choices = []choice{
		{"Leave them uncommitted", "review with git status and commit when you are ready (default)", Outcome{Done: true, Note: "Not committed: review with git status"}},
		{"Commit only the files this operation changed", "one commit named chore: " + op + "; your own changes stay out; smith undo can revert it", Outcome{Commit: append(append([]string{"commit"}, strings.Fields(op)...), "--only-operation")}},
		{"Commit everything git shows as changed", "one commit with every changed file, your own work included; you see the list first", Outcome{Commit: append(append([]string{"commit"}, strings.Fields(op)...), "--all")}},
	}
	m.foot = "Nothing is committed unless you choose it. You see the list of files before it is."
	return m
}

// NewMissingTool builds the dialog for an add whose first check could not run
// because a tool is missing: the project exists, and the choices say how to go on.
//
// Parameters: t is what the add printed.
// Returns: the dialog, on the first choice.
func NewMissingTool(t MissingTool) Model {
	m := Model{title: "Add " + t.Project + ": it could not be checked"}
	where := ""
	if t.Dir != "" {
		where = " in " + t.Dir
	}
	count := ""
	if t.Files > 0 {
		count = " (" + Plural(t.Files, "file") + ")"
	}
	m.facts = []Fact{
		{Good, t.Project + " was created" + where + count + "."},
		{Bad, "It could not be checked yet: " + t.Title + " is not installed."},
	}
	if t.URL != "" {
		m.choices = append(m.choices, choice{"Open the install page", strings.TrimPrefix(strings.TrimPrefix(t.URL, "https://"), "http://") + " (copies the link)", Outcome{Open: t.URL}})
	}
	m.choices = append(m.choices, choice{"Install it with smith", "smith asks how, and nothing runs before you choose", Outcome{Install: []string{"tools", "install"}}})
	if len(t.Check) > 0 {
		m.choices = append(m.choices, choice{"Check again", "after you installed it", Outcome{Check: t.Check}})
	}
	m.choices = append(m.choices, choice{"Continue without checking", "Doctor keeps reminding you", Outcome{Done: true, Note: "Not checked: " + t.Title + " is missing. Doctor will keep reminding you."}})
	m.foot = "Nothing is lost: your project stays on disk."
	return m
}

// ChoosesVersion reports whether a command line is a sync (smith sync, or
// smith deps up --align) that has not said which version wins yet: no --prefer,
// no --to, and it is not a check or a dry run.
//
// Parameters: argv is the smith command line, without the leading "smith".
// Returns: true when the person should be asked which version wins.
func ChoosesVersion(argv []string) bool {
	sync := len(argv) > 0 && argv[0] == "sync"
	align := len(argv) > 1 && (argv[0] == "up" || (argv[0] == "deps" && argv[1] == "up")) && slices.Contains(argv, "--align")
	if !sync && !align {
		return false
	}
	for _, a := range argv {
		if a == "--prefer" || a == "--to" || a == "--check" || a == "--dry-run" || strings.HasPrefix(a, "--prefer=") || strings.HasPrefix(a, "--to=") {
			return false
		}
	}
	return true
}

// NewWhichVersion builds the dialog that asks which version wins when projects
// declare a dependency at different versions. Each choice goes on to the usual
// plan and confirmation with the answer as --prefer; nothing is written here.
//
// Parameters: argv is the sync command line that has no --prefer yet.
// Returns: the dialog, on the first choice.
func NewWhichVersion(argv []string) Model {
	m := Model{title: "Which version wins?"}
	m.facts = []Fact{{Info, "Some projects declare the same dependency at different versions. You choose the one every project gets."}}
	with := func(rule string) []string { return append(append([]string(nil), argv...), "--prefer", rule) }
	m.choices = []choice{
		{"The installed version (smith's pick)", "the version the package manager has installed (default)", Outcome{Check: with("installed")}},
		{"The highest declared", "projects on an older version move up", Outcome{Check: with("highest")}},
		{"The lowest declared", "projects on a newer version move down", Outcome{Check: with("lowest")}},
		{"Cancel", "nothing is changed", Outcome{Done: true, Note: "Cancelled: no version was changed"}},
	}
	m.foot = "You see the plan before anything is written. One dependency at a time: smith sync --to <name>=<version>."
	return m
}

// Place sizes the dialog for a screen of sw by sh cells.
func (m Model) Place(sw, sh int) Model {
	w := min(max(sw-6, 40), 92)
	h := min(len(m.facts)*2+len(m.choices)*2+6, max(sh-2, 8))
	m.rect = dialog.Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/2, 0), W: w, H: max(h, 8)}
	return m
}

// Rect is where the dialog sits after Place.
func (m Model) Rect() dialog.Rect { return m.rect }

// Cursor is the index of the choice under the cursor.
func (m Model) Cursor() int { return m.cursor }

// Labels are the choices, in order.
func (m Model) Labels() []string {
	out := make([]string, len(m.choices))
	for i, c := range m.choices {
		out[i] = c.label
	}
	return out
}

// Key handles one key, by the name Bubble Tea gives it.
//
// Parameters: name is the key name ("up", "enter", "esc").
// Returns: the new dialog and what the host must do; Esc closes with a note
// that nothing was changed by the dialog.
func (m Model) Key(name string) (Model, Outcome) {
	switch name {
	case "up", "k", "shift+tab":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j", "tab":
		m.cursor = min(m.cursor+1, len(m.choices)-1)
	case "enter":
		if m.cursor < len(m.choices) {
			return m, m.choices[m.cursor].out
		}
	case "esc", "q", "ctrl+c":
		return m, Outcome{Done: true}
	}
	return m, Outcome{}
}
