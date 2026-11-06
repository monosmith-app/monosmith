package settingsview

import (
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Model is the state of the Settings tab. It is a value: every method returns
// the next one.
type Model struct {
	b     Backend
	forms []Form
	snap  Snapshot
	// loaded is true once the workspace was read; loading while it is being read.
	loaded, loading bool
	section         Section
	// onItems is true while the keyboard is in the right pane, false while it
	// is in the list of sections.
	onItems bool
	cursor  [6]int
	// scroll is where the window of each section's list starts, apart from
	// the cursor, so the wheel scrolls without selecting (see pane.Scroll).
	scroll [6]*pane.Scroll
	// listW and detailRows are the width of the list of sections and the rows of
	// the detail the person gave by moving a divider; 0 is automatic.
	listW, detailRows int
	// ed is the editor over the right pane, nil when none is open.
	ed *editor
	// note is the last thing that happened, shown under the list; failed makes it an error.
	note   string
	failed bool
	// confirm is the architecture whose warning was shown: choosing it again,
	// with nothing between, applies it.
	confirm string
	// savedTheme and savedIcons are the saved choice of the Style section.
	savedTheme, savedIcons string
	tipSeen                bool
	// busy says what a change in the background is doing ("" when none); stoppable
	// is true when Esc may drop its result; ticket tells the answer of the work
	// that is wanted from one the person stopped.
	busy      string
	stoppable bool
	ticket    int
	// tick and waited move the spinner line (see Pulse).
	tick   int
	waited time.Duration
}

// New builds the tab on a backend; forms are the guided forms its last section opens.
func New(b Backend, forms []Form) Model {
	return Model{
		b: b, forms: forms, onItems: true,
		scroll:     [6]*pane.Scroll{{}, {}, {}, {}, {}, {}},
		savedTheme: normalise(b.Style.Theme), savedIcons: normalise(b.Style.Icons), tipSeen: b.TipSeen,
	}
}

// normalise turns "auto" into "": nothing saved.
func normalise(v string) string {
	if v == theme.Auto {
		return ""
	}
	return v
}

// Available is false when the tab has nothing to read (no backend).
func (m Model) Available() bool { return m.b.Load != nil }

// Show opens a section, with the keyboard in its rows.
func (m Model) Show(s Section) Model {
	m.section, m.onItems = s, true
	return m
}

// Section is the open section.
func (m Model) Section() Section { return m.section }

// Open is called each time the tab is shown: it reads the workspace the first
// time and remembers that the person has seen the tab, so Home stops
// pointing at it.
func (m Model) Open() (Model, tea.Cmd) {
	var cmds []tea.Cmd
	if !m.tipSeen {
		m.tipSeen = true
		if mark := m.b.MarkTipSeen; mark != nil {
			cmds = append(cmds, func() tea.Msg { mark(); return nil })
		}
	}
	if !m.loaded && !m.loading && m.b.Load != nil {
		m.loading = true
		cmds = append(cmds, m.load())
	}
	return m, tea.Batch(cmds...)
}

// TipWanted is true until the person has opened the tab once: Home points at it.
func (m Model) TipWanted() bool { return !m.tipSeen && m.b.Load != nil }

// Busy is true while a change or a review runs in the background.
func (m Model) Busy() bool { return m.busy != "" }

// Pulse moves the spinner line of a tab that waits for a change.
//
// Parameters: tick is the animation frame; waited is how long the work lasts.
// Returns: the tab with the moving line.
func (m Model) Pulse(tick int, waited time.Duration) Model {
	m.tick, m.waited = tick, waited
	return m
}

// begin marks the start of background work: the tab says what it does at
// once. stoppable is true for work that only reads, so Esc may drop it.
func (m Model) begin(doing string, stoppable bool) Model {
	m.busy, m.stoppable, m.ticket = doing, stoppable, m.ticket+1
	return m
}

// Modal is true while an editor has the keyboard: plain keys are its own.
func (m Model) Modal() bool { return m.ed != nil }

func (m Model) load() tea.Cmd {
	load := m.b.Load
	return func() tea.Msg { return Loaded{Snapshot: load()} }
}

// Reload reads the workspace again.
func (m Model) Reload() (Model, tea.Cmd) {
	if m.b.Load == nil || m.loading {
		return m, nil
	}
	m.loading = true
	return m, m.load()
}

// Handle takes what the background work reports.
func (m Model) Handle(msg Msg) (Model, Result) {
	switch msg := msg.(type) {
	case Loaded:
		m.snap, m.loaded, m.loading = msg.Snapshot, true, false
		m = m.clampCursors()
	case Applied:
		m.busy = ""
		return m.applied(msg)
	case Reviewed:
		if msg.Ticket != m.ticket {
			// The person stopped the review: its answer is not wanted.
			return m, Result{}
		}
		m.busy = ""
		if msg.Err != nil {
			m.note, m.failed = msg.Err.Error(), true
			return m, Result{Note: m.note, Failed: true}
		}
		m.ed = newReview(msg.Name, msg.Review)
	case StyleSaved:
		m.busy = ""
		if msg.Err != nil {
			m.note, m.failed = words.Say("settingsview.style.failed", "reason", msg.Err.Error()), true
			return m, Result{Note: m.note, Failed: true}
		}
		m.savedTheme, m.savedIcons = msg.Theme, msg.Icons
		choice := msg.Choice
		m.note, m.failed = "", false
		return m, Result{Note: words.Say("settingsview.style.saved", "style", describeStyle(msg.Theme, msg.Icons)), Theme: &choice}
	}
	return m, Result{}
}

// applied is the end of a change: the editor closes on success and shows the
// reason on failure; the fresh snapshot replaces the old.
func (m Model) applied(msg Applied) (Model, Result) {
	if msg.Err != nil {
		if m.ed != nil {
			e := *m.ed
			e.err = msg.Err.Error()
			m.ed = &e
			return m, Result{}
		}
		m.note, m.failed = msg.Err.Error(), true
		return m, Result{Note: m.note, Failed: true}
	}
	m.ed = nil
	// The shell shows a success as a toast: keep it out of the pane, or it prints twice.
	m.note, m.failed = "", false
	if msg.Fresh != nil {
		m.snap = *msg.Fresh
		m = m.clampCursors()
	}
	return m, Result{Note: msg.Note}
}

func describeStyle(themeName, icons string) string {
	return words.Say("settingsview.style.describe", "theme", orAuto(themeName), "symbols", orAuto(icons))
}

func orAuto(v string) string {
	if v == "" {
		return theme.Auto
	}
	return v
}

// Reasons say, by command id, why a key of the tab does nothing now, so the
// footer leaves it out and the key is refused with the reason.
func (m Model) Reasons() map[string]string {
	r := map[string]string{}
	if m.ed != nil {
		for _, id := range []string{"settings.move", "settings.change", "settings.sections", "settings.trust", "settings.forget", "settings.unset"} {
			r[id] = words.Say("settingsview.reason.editor")
		}
		return r
	}
	it, ok := m.current()
	if !m.onItems {
		r["settings.back"] = words.Say("settingsview.reason.sections")
		r["settings.trust"] = words.Say("settingsview.reason.rows")
		r["settings.forget"], r["settings.unset"] = r["settings.trust"], r["settings.trust"]
		return r
	}
	switch {
	case !ok:
		r["settings.change"] = words.Say("settingsview.reason.nothing")
	case it.kind == kindInstalled || it.kind == kindRecipe:
		r["settings.change"] = words.Say("settingsview.reason.listed")
	}
	if !ok || it.kind != kindExternal {
		r["settings.trust"] = words.Say("settingsview.reason.select.trust")
		r["settings.forget"] = words.Say("settingsview.reason.select.forget")
	} else if !m.snap.External[it.idx].Trusted {
		r["settings.forget"] = words.Say("settingsview.reason.untrusted")
	}
	if !ok || it.kind != kindSetting {
		r["settings.unset"] = words.Say("settingsview.reason.select.unset")
	} else if !m.snap.Settings[it.idx].Set {
		r["settings.unset"] = words.Say("settingsview.reason.default")
	}
	return r
}
