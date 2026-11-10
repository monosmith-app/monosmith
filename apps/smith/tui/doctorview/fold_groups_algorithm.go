package doctorview

// Codes a drawn line carries in the map View returns, besides a finding's
// index (0 and up): the line of the collapsed passes, the headline (a click on
// it folds or unfolds every group) and, from HeaderCode(0) down, the heading
// of each group.
const (
	// PassedRow is the line that stands for the collapsed passes.
	PassedRow = -1
	// ToggleAllRow is the headline, which folds or unfolds every group.
	ToggleAllRow = -2
	// extraBase is the code of the first extra check (see Extra); later ones count down.
	extraBase = -50
	// headerBase is the code of the first group's heading; later ones count down.
	headerBase = -100
)

// ExtraCode is the code of the row of extra check i (its place in Extras).
func ExtraCode(i int) int { return extraBase - i }

// HeaderCode is the code of the heading of group g (its place in the groups as
// drawn).
func HeaderCode(g int) int { return headerBase - g }

// rowKind tells what the cursor can rest on.
type rowKind int

const (
	headerRow rowKind = iota
	findingRow
	passedRow
	extraRow
)

// row is one place the cursor can rest: a group's heading, a finding, the
// line of collapsed passes, or a check doctor does not run (extra, its place
// in Extras).
type row struct {
	kind    rowKind
	group   int
	finding int
	extra   int
}

// layout is every row in the order it is drawn: with groups, each heading
// followed by its findings while the group is open; flat, only the findings;
// the collapsed passes, then the checks doctor does not run.
func (v Model) layout() []row {
	var rows []row
	for gi, g := range v.groups() {
		if v.Grouped {
			rows = append(rows, row{kind: headerRow, group: gi})
		}
		if !v.groupOpen(g) {
			continue
		}
		for _, i := range g.Indexes {
			rows = append(rows, row{kind: findingRow, group: gi, finding: i})
		}
	}
	if v.collapsed() {
		rows = append(rows, row{kind: passedRow})
	}
	if v.Report != nil {
		for i := range v.Extras {
			rows = append(rows, row{kind: extraRow, extra: i})
		}
	}
	return rows
}

// groupOpen is whether a group shows its findings: what the person chose for
// it, else what they chose for all of them, else open. Without groups there is
// nothing to fold.
func (v Model) groupOpen(g Group) bool {
	if !v.Grouped {
		return true
	}
	if open, set := v.Folds[g.Project]; set {
		return open
	}
	return v.All >= 0
}

// AllOpen is whether every group shows its findings.
func (v Model) AllOpen() bool {
	for _, g := range v.groups() {
		if !v.groupOpen(g) {
			return false
		}
	}
	return true
}

// ToggleAll opens every group, or closes them all when all are open.
func (v Model) ToggleAll() Model {
	close := v.AllOpen()
	v.Folds, v.All = nil, 1
	if close {
		v.All = -1
	}
	return v
}

// toggleGroup opens group gi when closed and closes it when open.
func (v Model) toggleGroup(gi int) Model {
	groups := v.groups()
	if gi < 0 || gi >= len(groups) {
		return v
	}
	return v.setGroup(groups[gi], !v.groupOpen(groups[gi]))
}

// setGroup records that a group is open or closed, without touching the
// map another copy of the model holds.
func (v Model) setGroup(g Group, open bool) Model {
	folds := make(map[string]bool, len(v.Folds)+1)
	for k, o := range v.Folds {
		folds[k] = o
	}
	folds[g.Project] = open
	v.Folds = folds
	return v
}

// firstFinding is the row of the first finding, which is where the cursor
// starts so the detail has something to tell; 0 when there is none.
func (v Model) firstFinding() int {
	for at, r := range v.layout() {
		if r.kind == findingRow {
			return at
		}
	}
	return 0
}

// OnHeader is true when the cursor is on a group's heading.
func (v Model) OnHeader() bool {
	rows := v.layout()
	return v.Cursor >= 0 && v.Cursor < len(rows) && rows[v.Cursor].kind == headerRow
}

// IsCursorRow is true when a code from the map View returns is the row the
// cursor is on.
func (v Model) IsCursorRow(code int) bool {
	rows := v.layout()
	if v.Cursor < 0 || v.Cursor >= len(rows) {
		return false
	}
	switch r := rows[v.Cursor]; r.kind {
	case headerRow:
		return code == HeaderCode(r.group)
	case findingRow:
		return code == r.finding
	case extraRow:
		return code == ExtraCode(r.extra)
	}
	return code == PassedRow
}
