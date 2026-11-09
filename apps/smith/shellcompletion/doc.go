// Package shellcompletion installs tab completion for smith in the shell of
// the person, only when the person asks. It writes the completion script that
// cobra generates to the standard user location of the shell, between marker
// lines, so that Uninstall removes exactly what Install added.
//
// Nothing here writes without a plan: PlanFor lists each file and RenderPlan
// shows each line first, and the caller asks. A file that a package of smith
// ships is left alone, and so is a file that smith did not write. The machine
// comes through the shellenv.Environment port. The one-time offer, the answer
// it remembers and the `smith completion` command are in completioncommand.
package shellcompletion
