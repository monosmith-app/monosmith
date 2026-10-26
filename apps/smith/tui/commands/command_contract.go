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
	// Guided opens the guided form for the action titled Guide.
	Guided
	// Palette opens the command palette.
	Palette
	// Help opens the shortcuts overlay.
	Help
	// Quit leaves the shell.
	Quit
	// Stop cancels the run in progress.
	Stop
	// Follow toggles following the end of the log.
	Follow
	// OpenDoctor shows the Doctor view (which checks the workspace the first time).
	OpenDoctor
	// CopyFocused copies the focused item to the clipboard.
	CopyFocused
	// ToggleMouse turns mouse reporting off or on.
	ToggleMouse
	// ViewAction runs the action of the open view that Action names.
	ViewAction
	// OpenFixWizard opens the doctor fix wizard on the fixes that need no
	// decision (or, when all of them ask, on the first).
	OpenFixWizard
	// OpenAdoptWizard opens the adopt wizard on the folder: what was found,
	// the level, the plan, the conversion and the closing checks.
	OpenAdoptWizard
	// OpenPackagePrompt opens the guided form that asks which project and which
	// packages to install (Action "install") or remove (Action "uninstall").
	OpenPackagePrompt
	// PaneKey is a key the open view handles itself (arrows, Enter, Space…):
	// it is listed in the footer and the help but never run by the shell.
	PaneKey
)

// The ids of the main views.
const (
	TabHome     = "home"
	TabProjects = "projects"
	TabRun      = "run"
	TabDoctor   = "doctor"
	TabDeps     = "deps"
	TabRelease  = "release"
	TabSettings = "settings"
	TabGuide    = "guide"
)

// Command is one thing the shell can do.
type Command struct {
	ID string
	// Group is the heading it sits under in the help.
	Group string
	Title string
	Hint  string
	// Keys are the Bubble Tea key names that trigger it. Shortcut is how they
	// are shown (derived from Keys unless set), Glyph a shorter form for the
	// footer; a command with a Glyph stands for several keys and is not clickable.
	Keys     []string
	Shortcut string
	Glyph    string
	Kind     Kind
	// Views are the views the command works in; empty means every view. A key
	// in a view's own list wins over the same key in the global one.
	Views []string
	// Footer is the short label in the footer hints; empty keeps the command
	// out of the footer. ShowIn are the views whose footer shows it (Views when
	// empty); Tail puts it at the end of every footer.
	Footer string
	// ScopeInFooter puts the project a run acts on in the footer label
	// ("build web"), so the scope is known before the key is pressed.
	ScopeInFooter bool
	ShowIn        []string
	Tail          bool
	// WhileEditing means the key still works while a text field has the keyboard.
	WhileEditing bool
	// Action is the view action for ViewAction.
	Action string
	// Argv is the smith command line for Execute.
	Argv []string
	// All makes RunTask run on every project, whatever is selected.
	All bool
	// Task is the task for RunTask.
	Task string
	// Tab is the view for ShowTab.
	Tab string
	// Guide is the title of the guided action for Guided.
	Guide string
	// Presents are the ids of the actions (see package actions) the command
	// stands for when its Argv or Task does not already name them: a view
	// that runs several, or a key that reloads one.
	Presents []string
	// Confirm means the command changes the workspace or publishes: a dialog
	// shows the command line before it runs.
	Confirm bool
	// NeedsWorkspace, NeedsIdle (no run in progress) and NeedsRunning say when
	// the command is unavailable.
	NeedsWorkspace, NeedsIdle, NeedsRunning bool
	// NeedsNarrowList and NeedsDetailOpen are for the keys of the narrow
	// layout, where the detail is opened over the list and closed again.
	NeedsNarrowList, NeedsDetailOpen bool
	// OnlyOutside and OnlyInside hide the command in the other situation.
	OnlyOutside, OnlyInside bool
	// NeedsContent hides the command in a folder that holds nothing, where there is nothing for it to work on.
	NeedsContent bool
}

// State is what decides whether a command is available.
type State struct {
	// View is the open view's id (a Tab constant).
	View string
	// Editing is true while a text field has the keyboard; Form while a guided
	// form is open. Plain keys do nothing in either.
	Editing, Form bool
	InWorkspace   bool
	// Narrow is true when the detail is not beside the list; DetailOpen when it
	// fills the pane instead of the list.
	Narrow, DetailOpen bool
	// Running is true while a command is running in the run view.
	Running bool
	// EmptyFolder is true outside a workspace when the folder holds nothing yet.
	EmptyFolder bool
	// Project is the project the run keys act on (the one selected in the list),
	// and Projects how many the workspace has; the footer and the palette name
	// the scope of a run with them.
	Project  string
	Projects int
	// Reasons say, by command id, why the open view cannot do something now
	// (nothing to update, no safe fix): the command is refused with that
	// reason and the footer leaves its hint out.
	Reasons map[string]string
}
