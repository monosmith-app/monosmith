package operationplan

// Action is what an operation does to a path.
type Action string

// The actions a plan line can carry.
const (
	// Create adds a file or a directory.
	Create Action = "create"
	// Change edits a file that exists.
	Change Action = "change"
	// Remove deletes a file or a directory.
	Remove Action = "remove"
)

// Line is one path an operation touches.
type Line struct {
	Action Action
	// Path is workspace-relative, with forward slashes.
	Path string
	// Note says why, in a few words. It may be empty.
	Note string
	// Owner is "smith" or "you" when smith.lock records the path, else empty.
	Owner string
	// Diff is the unified diff of a file that exists and may be hand-edited.
	// It may be empty; Render shows at most MaxDiffLines of it.
	Diff string
}

// Group is the lines that one feature adds, under one heading.
type Group struct {
	Heading string
	Lines   []Line
}

// Plan is what an operation will do, before it does it.
type Plan struct {
	// Title is one sentence: "add package billing".
	Title  string
	Groups []Group
}

// Add appends a line to the group with the given heading and creates the
// group when it does not exist yet. Groups keep the order of first use.
func (p *Plan) Add(heading string, l Line) {
	for i := range p.Groups {
		if p.Groups[i].Heading == heading {
			p.Groups[i].Lines = append(p.Groups[i].Lines, l)
			return
		}
	}
	p.Groups = append(p.Groups, Group{Heading: heading, Lines: []Line{l}})
}

// Risky reports whether the plan can lose something: it deletes a path, or it
// edits a file that exists and shows a diff of it, or one the person owns.
// A plan that only creates files, or edits files smith owns, is not risky.
func (p Plan) Risky() bool {
	for _, g := range p.Groups {
		for _, l := range g.Lines {
			if l.Action == Remove || (l.Action == Change && (l.Diff != "" || l.Owner == "you")) {
				return true
			}
		}
	}
	return false
}

// Empty reports whether the plan lists no path.
func (p Plan) Empty() bool {
	for _, g := range p.Groups {
		if len(g.Lines) > 0 {
			return false
		}
	}
	return true
}
