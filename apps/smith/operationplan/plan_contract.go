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

// Empty reports whether the plan lists no path.
func (p Plan) Empty() bool {
	for _, g := range p.Groups {
		if len(g.Lines) > 0 {
			return false
		}
	}
	return true
}
