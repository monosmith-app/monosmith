package commands

// Kind is what a command does when chosen.
type Kind int

// The kinds of command.
const (
	// ShowTab switches the main view to Tab.
	ShowTab Kind = iota
	// Execute runs Argv as a smith command line in a pane.
	Execute
	// RunTask runs Task on the selected project, or on every project.
	RunTask
	// RunSelection runs the task selected in the navigator.
	RunSelection
	// Guided opens the guided form for the action titled Guide.
	Guided
	// Palette opens the command palette.
	Palette
	// Help opens the shortcuts overlay.
	Help
	// About shows the version.
	About
	// Quit leaves the shell.
	Quit
	// Stop cancels the run in progress.
	Stop
	// Follow toggles following the end of the log.
	Follow
	// NextPane and PrevPane move the focus between panes.
	NextPane
	PrevPane
	// ReloadDoctor runs doctor again.
	ReloadDoctor
	// ContextMenu opens the context menu of the selection.
	ContextMenu
)

// The ids of the main views.
const (
	TabProjects = "projects"
	TabRun      = "run"
	TabDoctor   = "doctor"
	TabRelease  = "release"
	TabSettings = "settings"
	TabGuide    = "guide"
)

// Command is one thing the shell can do.
type Command struct {
	ID string
	// Menu is the menu it sits in; empty means it is only on keys and in the palette.
	Menu  string
	Title string
	Hint  string
	// Keys are the Bubble Tea key names that trigger it; Shortcut is how the
	// first of them is shown beside the control.
	Keys     []string
	Shortcut string
	Kind     Kind
	// Argv is the smith command line for Execute.
	Argv []string
	// Task is the task for RunTask.
	Task string
	// Tab is the view for ShowTab.
	Tab string
	// Guide is the title of the guided action for Guided.
	Guide string
	// Confirm means the command changes the workspace or publishes: a dialog
	// shows the command line before it runs.
	Confirm bool
	// NeedsWorkspace, NeedsIdle (no run in progress), NeedsRunning and
	// NeedsSelection say when the command is unavailable.
	NeedsWorkspace, NeedsIdle, NeedsRunning, NeedsSelection bool
	// OnlyOutside and OnlyInside hide the command in the other situation.
	OnlyOutside, OnlyInside bool
}

// State is what decides whether a command is available.
type State struct {
	InWorkspace bool
	// Running is true while a command is running in the run view.
	Running bool
	// HasSelection is true when a task is selected in the navigator.
	HasSelection bool
}
