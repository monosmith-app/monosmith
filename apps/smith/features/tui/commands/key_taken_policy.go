package commands

import "slices"

// KeyTaken is true when key already does something in view: a command of that
// view, or one that works in every view. A button whose key is chosen at run time
// (the task of a project) takes only a key for which this is false.
//
// Parameters: key is a Bubble Tea key name; view is a Tab constant.
func KeyTaken(key, view string) bool {
	for _, c := range all {
		if slices.Contains(c.Keys, key) && (len(c.Views) == 0 || slices.Contains(c.Views, view)) {
			return true
		}
	}
	return false
}
