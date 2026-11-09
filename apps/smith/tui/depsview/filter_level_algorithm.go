package depsview

import "github.com/monosmith-app/monosmith/apps/smith/tui/commands"

// Filter is one chip of the row at the top of the view: it limits the list to
// the releases of one level. One chip is on at a time; FilterNone stands for no
// chip, and a view that has none chosen shows every release.
type Filter int

// The chips, in the order the row draws them.
const (
	FilterNone Filter = iota
	FilterAll
	FilterPatch
	FilterMinor
	FilterMajor
)

// filters are the real chips, in the order of the row.
var filters = []Filter{FilterAll, FilterPatch, FilterMinor, FilterMajor}

// CommandID is the id of the command (and so the key) that turns the chip on;
// "" for FilterNone.
func (f Filter) CommandID() string {
	switch f {
	case FilterAll:
		return "deps.filter-all"
	case FilterPatch:
		return "deps.filter-patch"
	case FilterMinor:
		return "deps.filter-minor"
	case FilterMajor:
		return "deps.filter-major"
	}
	return ""
}

// Key is the key of the chip, as the button shows it.
func (f Filter) Key() string { return commands.ShortcutOf(f.CommandID()) }

// level is the level of release the chip lets through; "" lets every level through.
func (f Filter) level() string {
	switch f {
	case FilterPatch:
		return "patch"
	case FilterMinor:
		return "minor"
	case FilterMajor:
		return "major"
	}
	return ""
}

// FilterOf is the chip that a command id stands for, FilterNone for any other id.
func FilterOf(id string) Filter {
	for _, f := range filters {
		if f.CommandID() == id {
			return f
		}
	}
	return FilterNone
}

// filterOfKey is the chip a key turns on, FilterNone for any other key.
func filterOfKey(key string) Filter {
	for _, f := range filters {
		if f.Key() == key {
			return f
		}
	}
	return FilterNone
}

// rowsAt are the rows of one level, in their order; every row for the All chip.
func rowsAt(rows []Row, f Filter) []Row {
	level := f.level()
	if level == "" {
		return rows
	}
	var out []Row
	for _, r := range rows {
		if levelOf(r) == level {
			out = append(out, r)
		}
	}
	return out
}

// levelOf is the level of a row as the view counts it: anything that is not a
// patch or a minor release is a major one.
func levelOf(r Row) string {
	switch r.Level {
	case "patch", "minor":
		return r.Level
	}
	return "major"
}
