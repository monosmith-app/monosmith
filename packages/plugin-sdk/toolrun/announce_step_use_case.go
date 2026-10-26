package toolrun

import "fmt"

// Step tells the user, in one line on Log, what a plugin is about to do before
// a slow command such as creating a virtual environment, so a wait of seconds
// does not look like a hang.
func Step(message string) {
	_, _ = fmt.Fprintln(Log, message+" ...")
}
