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
		Name   string
		Review TrustReview
		Err    error
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

// Result is what the shell does after the tab handled a key, a click or a
// message.
type Result struct {
	// Cmd is the background work to start.
	Cmd tea.Cmd
	// Note is a sentence for a toast; Failed makes it an error.
	Note   string
	Failed bool
	// Theme is set when the interface must redraw in another theme.
	Theme *theme.Choice
	// Form is the title of a guided form to open.
	Form string
	// Run is a smith command line (without the leading smith) the shell
	// shows, asks about and runs in the Run view: the tab never runs it.
	Run []string
	// Search asks the shell to search the plugin marketplace for Term (empty:
	// every plugin) and offer the plugins it finds.
	Search bool
	Term   string
}
