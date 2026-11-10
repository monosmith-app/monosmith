package doctorview

import (
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
)

// Filter is one of the three toggles on the headline: it shows or hides the
// checks of one state in every group.
type Filter int

// The filters, in the order the headline draws them. FilterNone is no filter.
const (
	FilterNone Filter = iota
	FilterFailed
	FilterWarnings
	FilterPassed
)

// filters are the real filters, in the order of the headline.
var filters = []Filter{FilterFailed, FilterWarnings, FilterPassed}

// CommandID is the id of the command (and so the key) that turns the filter on
// or off; "" for FilterNone.
func (f Filter) CommandID() string {
	switch f {
	case FilterFailed:
		return "doctor.failed"
	case FilterWarnings:
		return "doctor.warnings"
	case FilterPassed:
		return "doctor.passed"
	}
	return ""
}

// Key is the key of the filter, as the footer and the button show it.
func (f Filter) Key() string { return commands.ShortcutOf(f.CommandID()) }

// FilterOf is the filter that a command id stands for, FilterNone for any
// other id.
func FilterOf(id string) Filter {
	for _, f := range filters {
		if f.CommandID() == id {
			return f
		}
	}
	return FilterNone
}

// filterOfKey is the filter a key turns on or off, FilterNone for any other key.
func filterOfKey(key string) Filter {
	for _, f := range filters {
		if f.Key() == key {
			return f
		}
	}
	return FilterNone
}

// stateOf is the filter that governs a finding: the state it is in.
func stateOf(f doctor.Finding) Filter {
	switch {
	case f.OK:
		return FilterPassed
	case f.Warning:
		return FilterWarnings
	}
	return FilterFailed
}

// Shows is true while the checks of the filter's state are listed. Failed
// checks and warnings are listed by default; the passed checks are one line
// until the person asks for them.
func (v Model) Shows(f Filter) bool {
	switch f {
	case FilterFailed:
		return !v.HideFailed
	case FilterWarnings:
		return !v.HideWarnings
	case FilterPassed:
		return v.ShowPassed
	}
	return false
}

// visible is true when the filters let a finding be listed.
func (v Model) visible(f doctor.Finding) bool { return v.Shows(stateOf(f)) }

// Toggle turns a filter on or off. The cursor stays on its check when that
// check is still listed, and goes to the first one otherwise.
func (v Model) Toggle(f Filter) Model {
	kept, had := v.cursorFinding()
	switch f {
	case FilterFailed:
		v.HideFailed = !v.HideFailed
	case FilterWarnings:
		v.HideWarnings = !v.HideWarnings
	case FilterPassed:
		v.ShowPassed = !v.ShowPassed
	default:
		return v
	}
	v.Open = false
	v.Cursor = v.firstFinding()
	if !had {
		return v
	}
	for at, r := range v.layout() {
		if r.kind == findingRow && r.finding == kept {
			v.Cursor = at
		}
	}
	return v
}

// cursorFinding is the index of the finding under the cursor.
func (v Model) cursorFinding() (int, bool) {
	rows := v.layout()
	if v.Cursor < 0 || v.Cursor >= len(rows) || rows[v.Cursor].kind != findingRow {
		return 0, false
	}
	return rows[v.Cursor].finding, true
}

// NothingMatches is true when the report holds problems and the filters hide
// every one of them: the page then says so instead of showing an empty list.
func (v Model) NothingMatches() bool {
	return v.Report != nil && v.Report.Failed+v.Report.Warnings > 0 && len(v.groups()) == 0
}
