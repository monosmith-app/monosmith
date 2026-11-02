// Package resultdialog is the dialog that closes a run that changed the
// workspace, or that changed it and could not be checked. It states what
// happened in plain words (the project exists, N files; what could not be
// checked and why) and offers the next steps as choices with their
// consequences: commit the change now, or open the install page, install the
// tool, check again, continue without checking. smith commits nothing unasked,
// so the default choice is always the one that does nothing. The model runs
// nothing: a key returns an Outcome for the host to carry out.
package resultdialog
