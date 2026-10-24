// Package commands is the keymap and the list of everything the shell can do:
// each command with its keys, the view it belongs to, the label it has in the
// footer, whether it changes the workspace (so it needs a confirmation that
// shows the command line first) and when it is unavailable and why. The keys,
// the footer hints, the help overlay and the command palette are all drawn
// from this one list, so a label cannot name a key that does nothing.
package commands
