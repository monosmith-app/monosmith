// Package completioncommand is `smith completion`: install, uninstall and
// status for tab completion, the scripts for each shell, and the one-time offer
// that follows the first interactive run. The offer asks once; the answer
// ("no", "later" or "installed") is kept for the person in tui.json in the smith
// home and never goes into smith.toml. The files and the lines are in
// shellcompletion.
package completioncommand
