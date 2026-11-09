package dashboard

// Pointer is the cell under the mouse, counted from the top left of the page;
// a negative Col or Row means the pointer is not over the page.
type Pointer struct{ Col, Row int }

// Away is the pointer when it is not over the page.
var Away = Pointer{Col: -1, Row: -1}

// Target is a part of a line that answers a click: the cells [From, To) of the
// line Row.
type Target struct {
	Row, From, To int
	// Command is the id of the command a click runs; it is "" for a run, and
	// then Run is the run to open (0 is the newest).
	Command string
	Run     int
}

// Rendered is the page: its lines, each as wide as the page, and the targets
// of the clicks.
type Rendered struct {
	Lines   []string
	Targets []Target
}
