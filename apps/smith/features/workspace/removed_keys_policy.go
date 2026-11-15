package workspace

import "regexp"

// RemovedKey is a table smith.toml used to have under another name. smith no
// longer reads the old name; `smith doctor --fix` renames it in the text.
type RemovedKey struct {
	// Old and New are the table names as they are written ("task_groups").
	Old, New string
	// Array is true for an array of tables ([[old]]), false for a table ([old]).
	Array  bool
	header *regexp.Regexp
}

// Display writes a table name the way smith.toml writes its header.
func (k RemovedKey) Display(name string) string {
	if k.Array {
		return "[[" + name + "]]"
	}
	return "[" + name + "]"
}

func newRemovedKey(oldName, newName string, array bool) RemovedKey {
	return RemovedKey{
		Old: oldName, New: newName, Array: array,
		header: regexp.MustCompile(`(?m)^([ \t]*\[\[?[ \t]*)` + regexp.QuoteMeta(oldName) + `([ \t]*[.\]])`),
	}
}

// removedKeys lists every table smith renamed, oldest first.
var removedKeys = []RemovedKey{
	newRemovedKey("task_groups", "groups", true),
}

// FindRemovedKeys lists the removed tables smith.toml text still uses. It looks
// at table headers only, so a comment that names one does not count.
func FindRemovedKeys(raw []byte) []RemovedKey {
	var found []RemovedKey
	for _, k := range removedKeys {
		if k.header.Match(raw) {
			found = append(found, k)
		}
	}
	return found
}

// RenameRemovedKeys returns smith.toml text with every removed table header
// renamed ([[task_groups]] becomes [[groups]], [task_groups.tasks.x] becomes
// [groups.tasks.x]). Only the headers change: comments, spacing and the rest of
// the text stay as they are. It returns false when nothing needed renaming.
func RenameRemovedKeys(raw []byte) ([]byte, bool) {
	out := raw
	changed := false
	for _, k := range removedKeys {
		if k.header.Match(out) {
			out = k.header.ReplaceAll(out, []byte("${1}"+k.New+"${2}"))
			changed = true
		}
	}
	return out, changed
}
