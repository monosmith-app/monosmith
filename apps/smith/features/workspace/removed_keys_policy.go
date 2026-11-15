package workspace

import (
	"regexp"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tomltext"
)

// RemovedKey is a table smith.toml used to have under another name, or a key
// it used to have in another table. smith no longer reads the old one;
// `smith doctor --fix` moves it in the text.
type RemovedKey struct {
	// Old and New are the table names as they are written ("task_groups").
	Old, New string
	// Array is true for an array of tables ([[old]]), false for a table ([old]).
	Array bool
	// OldKey and NewKey are set for a key that moved between tables: OldKey
	// in table Old is now NewKey in table New.
	OldKey, NewKey string
	header         *regexp.Regexp
}

// Display writes a table name the way smith.toml writes its header, with the
// key of a moved key after it.
func (k RemovedKey) Display(name string) string {
	if k.OldKey != "" {
		key := k.NewKey
		if name == k.Old {
			key = k.OldKey
		}
		return "[" + name + "] " + key
	}
	if k.Array {
		return "[[" + name + "]]"
	}
	return "[" + name + "]"
}

// present reports whether smith.toml text still uses the old name.
func (k RemovedKey) present(raw []byte) bool {
	if k.OldKey != "" {
		_, ok := tomltext.Raw(string(raw), k.Old, k.OldKey)
		return ok
	}
	return k.header.Match(raw)
}

// rename moves the old name to the new one in smith.toml text. A moved key
// that is already set in its new table keeps the new value.
func (k RemovedKey) rename(raw []byte) []byte {
	if k.OldKey == "" {
		return k.header.ReplaceAll(raw, []byte("${1}"+k.New+"${2}"))
	}
	value, _ := tomltext.Raw(string(raw), k.Old, k.OldKey)
	text, _ := tomltext.Remove(string(raw), k.Old, k.OldKey)
	if _, set := tomltext.Raw(text, k.New, k.NewKey); !set {
		text = tomltext.SetRaw(text, k.New, k.NewKey, value)
	}
	return []byte(text)
}

func newRemovedKey(oldName, newName string, array bool) RemovedKey {
	return RemovedKey{
		Old: oldName, New: newName, Array: array,
		header: regexp.MustCompile(`(?m)^([ \t]*\[\[?[ \t]*)` + regexp.QuoteMeta(oldName) + `([ \t]*[.\]])`),
	}
}

// removedKeys lists every table smith renamed and every key it moved, oldest
// first.
var removedKeys = []RemovedKey{
	newRemovedKey("task_groups", "groups", true),
	// ADR 0101: the pin is the workspace's, read by the launchers as well as CI.
	{Old: "ci", OldKey: "smith_version", New: "workspace", NewKey: "smith"},
}

// FindRemovedKeys lists the removed tables smith.toml text still uses. It looks
// at table headers only, so a comment that names one does not count.
func FindRemovedKeys(raw []byte) []RemovedKey {
	var found []RemovedKey
	for _, k := range removedKeys {
		if k.present(raw) {
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
		if k.present(out) {
			out = k.rename(out)
			changed = true
		}
	}
	return out, changed
}
