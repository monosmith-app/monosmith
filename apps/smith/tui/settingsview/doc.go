// Package settingsview is the Settings tab: a list of sections on the left and
// the editor or detail of the open section on the right. Plugins switches the
// built-in plugins on and off and trusts the external ones, Architecture
// chooses how the workspace's code is laid out, Style picks the
// theme and the symbols with a live preview, Workspace edits the settings
// `smith config` knows with an editor for each type, Recipes lists the
// workspace's recipes, and Forms opens the guided forms.
//
// The view reads and writes through a Backend: functions that call what the
// CLI commands call (the same code, in-process, never a child process), so a
// change made here leaves smith.toml byte for byte as `smith plugin disable`
// or `smith config set` would. Backend functions run in the background and
// come back as messages (see Msg).
package settingsview
