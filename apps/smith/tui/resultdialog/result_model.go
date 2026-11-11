package resultdialog

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/libs/words/words"
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
	rect layout.Rect
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
	m := Model{title: words.Say("resultdialog.done", "operation", op)}
	created, isAdd := ParseCreated(argv, lines)
	if isAdd {
		m.title = words.Say("resultdialog.added", "kind", created.Kind, "name", created.Name)
		where := ""
		if created.Dir != "" {
			where = words.Say("resultdialog.where", "dir", created.Dir)
		}
		count := ""
		if created.Files > 0 {
			count = words.Say("resultdialog.files.note", "files", FileCount(created.Files))
		}
		m.facts = append(m.facts, Fact{Good, words.Say("resultdialog.created", "name", created.Name, "where", where, "files", count)})
		if created.Kind == "lib" {
			m.facts = append(m.facts, Fact{Info, words.Say("resultdialog.lib.tip", "name", created.Name)})
		}
	} else {
		m.facts = append(m.facts, Fact{Good, words.Say("resultdialog.written")})
	}
	m.facts = append(m.facts, Fact{Info, words.Say("resultdialog.uncommitted")})
	m.choices = []choice{
		{words.Say("resultdialog.leave.label"), words.Say("resultdialog.leave.hint"), Outcome{Done: true, Note: words.Say("resultdialog.leave.note")}},
		{words.Say("resultdialog.only.label"), words.Say("resultdialog.only.hint", "operation", op), Outcome{Commit: append(append([]string{"commit"}, strings.Fields(op)...), "--only-operation")}},
		{words.Say("resultdialog.all.label"), words.Say("resultdialog.all.hint"), Outcome{Commit: append(append([]string{"commit"}, strings.Fields(op)...), "--all")}},
	}
	m.foot = words.Say("resultdialog.foot.change")
	return m
}

// NewMissingTool builds the dialog for an add whose first check could not run
// because a tool is missing: the project exists, and the choices say how to go on.
//
// Parameters: t is what the add printed.
// Returns: the dialog, on the first choice.
func NewMissingTool(t MissingTool) Model {
	m := Model{title: words.Say("resultdialog.missing.title", "project", t.Project)}
	where := ""
	if t.Dir != "" {
		where = words.Say("resultdialog.where", "dir", t.Dir)
	}
	count := ""
	if t.Files > 0 {
		count = words.Say("resultdialog.files.note", "files", FileCount(t.Files))
	}
	m.facts = []Fact{
		{Good, words.Say("resultdialog.created", "name", t.Project, "where", where, "files", count)},
		{Bad, words.Say("resultdialog.missing.fact", "tool", t.Title)},
	}
	if t.URL != "" {
		m.choices = append(m.choices, choice{words.Say("resultdialog.open.label"), words.Say("resultdialog.open.hint", "url", strings.TrimPrefix(strings.TrimPrefix(t.URL, "https://"), "http://")), Outcome{Open: t.URL}})
	}
	m.choices = append(m.choices, choice{words.Say("resultdialog.install.label"), words.Say("resultdialog.install.hint"), Outcome{Install: []string{"tools", "install"}}})
	if len(t.Check) > 0 {
		m.choices = append(m.choices, choice{words.Say("resultdialog.check.label"), words.Say("resultdialog.check.hint"), Outcome{Check: t.Check}})
	}
	m.choices = append(m.choices, choice{words.Say("resultdialog.skip.label"), words.Say("resultdialog.skip.hint"), Outcome{Done: true, Note: words.Say("resultdialog.skip.note", "tool", t.Title)}})
	m.foot = words.Say("resultdialog.foot.missing")
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
	m := Model{title: words.Say("resultdialog.version.title")}
	m.facts = []Fact{{Info, words.Say("resultdialog.version.fact")}}
	with := func(rule string) []string { return append(append([]string(nil), argv...), "--prefer", rule) }
	m.choices = []choice{
		{words.Say("resultdialog.version.installed.label"), words.Say("resultdialog.version.installed.hint"), Outcome{Check: with("installed")}},
		{words.Say("resultdialog.version.highest.label"), words.Say("resultdialog.version.highest.hint"), Outcome{Check: with("highest")}},
		{words.Say("resultdialog.version.lowest.label"), words.Say("resultdialog.version.lowest.hint"), Outcome{Check: with("lowest")}},
		{words.Say("resultdialog.version.cancel.label"), words.Say("resultdialog.version.cancel.hint"), Outcome{Done: true, Note: words.Say("resultdialog.version.cancel.note")}},
	}
	m.foot = words.Say("resultdialog.foot.version")
	return m
}

// ChoosesCacheTier is true for a plain `cache clear` in a workspace that has a
// remote cache: which cache goes is the person's choice, asked before the
// confirmation. A line that already names a tier, or a dry run, is the person's own.
//
// Parameters: argv is the command line; remote is true when a remote cache is set.
// Returns: whether NewWhichCache asks first.
func ChoosesCacheTier(argv []string, remote bool) bool {
	if !remote || len(argv) < 2 || argv[0] != "cache" || argv[1] != "clear" {
		return false
	}
	return !slices.ContainsFunc(argv[2:], func(a string) bool { return a == "--remote" || a == "--all" || a == "--dry-run" })
}

// NewWhichCache asks which cache a clear deletes: the local one, the remote
// one or both. Each choice goes on to the usual confirmation, which shows the
// same question as the command line (it runs the line with --dry-run).
//
// Parameters: argv is the `cache clear` command line.
// Returns: the dialog, on the first choice.
func NewWhichCache(argv []string) Model {
	m := Model{title: words.Say("resultdialog.cache.title")}
	m.facts = []Fact{{Info, words.Say("resultdialog.cache.fact")}}
	with := func(flag ...string) []string { return append(append([]string(nil), argv...), flag...) }
	m.choices = []choice{
		{words.Say("resultdialog.cache.local.label"), words.Say("resultdialog.cache.local.hint"), Outcome{Check: with()}},
		{words.Say("resultdialog.cache.remote.label"), words.Say("resultdialog.cache.remote.hint"), Outcome{Check: with("--remote")}},
		{words.Say("resultdialog.cache.all.label"), words.Say("resultdialog.cache.all.hint"), Outcome{Check: with("--all")}},
		{words.Say("resultdialog.cache.cancel.label"), words.Say("resultdialog.cache.cancel.hint"), Outcome{Done: true, Note: words.Say("resultdialog.cache.cancel.note")}},
	}
	m.foot = words.Say("resultdialog.foot.cache")
	return m
}

// NewCommitChoice builds the question that is asked once: should smith commit
// each change it makes, so that it can be undone. Each choice says what it
// writes and where; Esc cancels and nothing is saved.
//
// Parameters: pending is true when a change waits for the answer.
// Returns: the dialog, on the first choice.
func NewCommitChoice(pending bool) Model {
	m := Model{title: words.Say("resultdialog.commit.title")}
	m.facts = []Fact{{Info, words.Say("resultdialog.commit.fact")}}
	cancel := choice{words.Say("resultdialog.commit.close.label"), words.Say("resultdialog.commit.close.hint"), Outcome{Done: true}}
	if pending {
		cancel = choice{words.Say("resultdialog.commit.cancel.label"), words.Say("resultdialog.commit.cancel.hint"), Outcome{Done: true, Note: words.Say("resultdialog.commit.cancel.note")}}
	}
	m.choices = []choice{
		{words.Say("resultdialog.commit.yes.label"), words.Say("resultdialog.commit.yes.hint"), Outcome{CommitChoice: CommitYes}},
		{words.Say("resultdialog.commit.no.label"), words.Say("resultdialog.commit.no.hint"), Outcome{CommitChoice: CommitNo}},
		{words.Say("resultdialog.commit.ask.label"), words.Say("resultdialog.commit.ask.hint"), Outcome{CommitChoice: CommitAsk}},
		cancel,
	}
	m.foot = words.Say("resultdialog.foot.commit")
	return m
}

// NewUndoOff builds the notice for an undo asked while commits are off: it
// says why there is nothing to undo, and a button opens the commit question.
//
// Returns: the dialog, on the first choice.
func NewUndoOff() Model {
	m := Model{title: words.Say("resultdialog.undooff.title")}
	m.facts = []Fact{{Info, words.Say("resultdialog.undooff.fact")}}
	m.choices = []choice{
		{words.Say("resultdialog.undooff.open.label"), words.Say("resultdialog.undooff.open.hint"), Outcome{CommitChoice: CommitOpen}},
		{words.Say("resultdialog.undooff.close.label"), words.Say("resultdialog.undooff.close.hint"), Outcome{Done: true}},
	}
	m.foot = words.Say("resultdialog.foot.undooff")
	return m
}

// Place sizes the dialog for a screen of sw by sh cells.
func (m Model) Place(sw, sh int) Model {
	w := min(max(sw-6, 40), 92)
	h := min(len(m.facts)*2+len(m.choices)*2+6, max(sh-2, 8))
	m.rect = layout.Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/2, 0), W: w, H: max(h, 8)}
	return m
}

// Rect is where the dialog sits after Place.
func (m Model) Rect() layout.Rect { return m.rect }

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
	switch commands.Nav(name) {
	case commands.NavUp, commands.NavPrev:
		m.cursor = max(m.cursor-1, 0)
	case commands.NavDown, commands.NavNext:
		m.cursor = min(m.cursor+1, len(m.choices)-1)
	case commands.NavCancel:
		return m, Outcome{Done: true}
	default:
		if name == "enter" && m.cursor < len(m.choices) {
			return m, m.choices[m.cursor].out
		}
	}
	return m, Outcome{}
}
