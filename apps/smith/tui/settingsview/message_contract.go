package settingsview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Msg is a message the tab's background work sends back; the shell passes
// every one to Model.Handle.
type Msg interface{ settingsMsg() }

type (
	// Loaded carries the workspace as read.
	Loaded struct{ Snapshot Snapshot }
	// Applied says how a change ended; Fresh is the workspace read again after it.
	Applied struct {
		Note  string
		Err   error
		Fresh *Snapshot
	}
	// Reviewed carries the review of a plugin the person asked to trust.
	Reviewed struct {
		// Ticket is the number of the request: an answer to a stopped one is dropped.
		Ticket int
		Name   string
		Review TrustReview
		Err    error
	}
	// UpdateChecked carries what the registry said when the person asked.
	UpdateChecked struct {
		Info UpdateInfo
		Err  error
	}
	// StyleSaved says the choice of theme and symbols was kept (or why not);
	// Choice is what the interface should draw with now.
	StyleSaved struct {
		Theme, Icons string
		Choice       theme.Choice
		Err          error
	}
)

func (Loaded) settingsMsg()     {}
func (Applied) settingsMsg()    {}
func (Reviewed) settingsMsg()   {}
func (StyleSaved) settingsMsg() {}

func (UpdateChecked) settingsMsg() {}

// PluginToggle is a switch of a built-in plugin that just ended: the plugin
// and the state it was switched to. The shell offers to switch it back.
type PluginToggle struct {
	Name    string
	Enabled bool
}

// Result is what the shell does after the tab handled a key, a click or a
// message.
type Result struct {
	// Cmd is the background work to start.
	Cmd tea.Cmd
	// Note is a sentence for a toast; Failed makes it an error.
	Note   string
	Failed bool
	// Toggled is set when the change was a plugin switch that went through: the
	// shell says it in the status line and offers to undo it.
	Toggled *PluginToggle
	// Theme is set when the interface must redraw in another theme.
	Theme *theme.Choice
	// Form is the title of a guided form to open.
	Form string
	// Run is a smith command line (without the leading smith) the shell
	// shows, asks about and runs in the run panel: the tab never runs it.
	Run []string
	// Search asks the shell to search the plugin marketplace for Term (empty:
	// every plugin) and offer the plugins it finds.
	Search bool
	Term   string
	// CacheSetup asks the shell to open the remote cache setup.
	CacheSetup bool
	// Update asks the shell to offer the update (the same dialog as the status bar).
	Update bool
	// Checked is set when the registry was asked: what it found, for the status bar.
	Checked *UpdateInfo
	// MutedChanged is true when the reminder was switched; Muted is its new state.
	MutedChanged, Muted bool
}
