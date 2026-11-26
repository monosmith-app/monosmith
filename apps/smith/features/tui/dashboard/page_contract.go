package dashboard

// Pointer is the cell under the mouse, counted from the top left of the page;
// a negative Col or Row means the pointer is not over the page.
type Pointer struct{ Col, Row int }

// Away is the pointer when it is not over the page.
var Away = Pointer{Col: -1, Row: -1}

// TargetKind says what a click on a target does.
type TargetKind int

// The kinds of target.
const (
	// TargetButton is a button: a click runs its Command.
	TargetButton TargetKind = iota
	// TargetRun is the Last run box: a click opens the run Run.
	TargetRun
	// TargetItem is the text of an item of the list: a click selects the item.
	TargetItem
	// TargetKey is the key a narrow item shows in place of its buttons: a click
	// runs its Command, as a button does.
	TargetKey
)

// Target is a part of a line that answers a click: the cells [From, To) of the
// line Row.
type Target struct {
	Row, From, To int
	// Command is the id of the command a click runs; it is "" for a run, and
	// then Run is the run to open (0 is the newest).
	Command string
	Run     int
	// Kind says what a click does.
	Kind TargetKind
	// Item is the item of the list the target belongs to (-1 outside the
	// list), and Button the place of the button among the item's buttons.
	Item, Button int
}

// Choice is where the keyboard is on the list: the item, and the button of it
// that Enter presses. A negative Item is no item.
type Choice struct{ Item, Button int }

// Rows are the lines [From, To) of the page an item of the list takes.
type Rows struct{ From, To int }

// Rendered is the page: its lines, each as wide as the page, the targets of
// the clicks and the lines of each item of the list.
type Rendered struct {
	Lines   []string
	Targets []Target
	Items   []Rows
}
