package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
)

// Page is what the shell needs of a view to show it in the one content pane:
// the title of its frame, the keys of its footer, a way to give it the keys
// and messages meant for it, and its body drawn for a width and a height.
// The shell asks the open tab for its Page and draws nothing else of the view,
// so a new view (the Run page, the later pages) plugs in by implementing it.
type Page interface {
	// Title is the heading set into the top edge of the pane.
	Title() string
	// Footer is the hints of the view's keys, then the tail every view ends with.
	Footer() (view, tail []commands.Hint)
	// Update gives the page a message (a key press); it returns what to run next.
	Update(msg tea.Msg) tea.Cmd
	// View draws the body in width by height cells, with what a click on each line means.
	View(width, height int) content
}
