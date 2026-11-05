package pathfinder

import (
	"slices"
	"strings"
)

// rowKind says what a row of the list stands for.
type rowKind int

const (
	rowUp rowKind = iota
	rowFolder
	rowFile
	rowSetting
)

// row is one line of the list.
type row struct {
	kind    rowKind
	name    string
	ignored bool
	hint    string
	// value is the text the field takes when the row is chosen.
	value string
}

// Setting is a name that is not a file but that the field also accepts, such
// as a smith.toml setting (ci.devcontainer). Hint says what it is.
type Setting struct{ Path, Hint string }

// buildRows makes the rows for the typed text.
//
// Parameters: dir and prefix are the typed text cut at its last slash; raw are
// the names of dir; settings are offered at the top level; ignored says whether
// a path is git-ignored.
// Returns: an "up" row (when dir is not the start folder and nothing is typed
// after it), then the folders, the files and the settings that match prefix,
// the ones that start with it first. A name that starts with a dot is listed
// only when prefix starts with a dot.
func buildRows(dir, prefix string, raw []rawEntry, settings []Setting, ignored func(rel string) bool) []row {
	var rows []row
	if dir != "" && prefix == "" {
		rows = append(rows, row{kind: rowUp, name: "..", hint: "up one folder", value: parentOf(dir)})
	}
	want := strings.ToLower(prefix)
	var folders, files, found []row
	for _, e := range raw {
		if strings.HasPrefix(e.name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}
		if !strings.Contains(strings.ToLower(e.name), want) {
			continue
		}
		r := row{kind: rowFile, name: e.name, ignored: ignored(dir + e.name), value: dir + e.name}
		if e.dir {
			r.kind, r.value = rowFolder, dir+e.name+"/"
		}
		if r.ignored {
			r.hint = "git-ignored"
		}
		if e.dir {
			folders = append(folders, r)
		} else {
			files = append(files, r)
		}
	}
	byName := func(a, b row) int { return strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name)) }
	slices.SortFunc(folders, byName)
	slices.SortFunc(files, byName)
	found = append(folders, files...)
	if dir == "" && prefix != "" {
		for _, s := range settings {
			if strings.Contains(strings.ToLower(s.Path), want) {
				found = append(found, row{kind: rowSetting, name: s.Path, hint: s.Hint, value: s.Path})
			}
		}
	}
	// A name that starts with what was typed comes before one that only holds it.
	slices.SortStableFunc(found, func(a, b row) int {
		return startsRank(a, want) - startsRank(b, want)
	})
	return append(rows, found...)
}

func startsRank(r row, want string) int {
	if strings.HasPrefix(strings.ToLower(r.name), want) {
		return 0
	}
	return 1
}
