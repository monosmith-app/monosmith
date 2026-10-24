package tabs

import "github.com/monosmith-app/monosmith/apps/smith/tui/theme"

// Tab is one view in the strip.
type Tab struct {
	ID string
	// Label is the full name; Short is a shorter one used when the full names
	// do not fit ("" means the name cannot be shortened).
	Label, Short string
	// Badge is a small styled mark after the name (a failure count, a spinner),
	// "" for none. Reserve is the room, in cells, kept for it in the fit
	// calculation even while it is empty, so a badge appearing never changes
	// which names are spelled out.
	Badge   string
	Reserve int
}

// Props is everything a strip is drawn from.
type Props struct {
	Theme theme.Theme
	Tabs  []Tab
	// Active and Hover are tab ids ("" for none).
	Active, Hover string
	// Width is the cells the strip may use.
	Width int
	// Tall draws the three-row strip that opens into the pane below; otherwise
	// the strip is one row.
	Tall bool
	// ShortenOrder lists the tab ids that give up their full name first.
	ShortenOrder []string
	// Divider is the column where the pane below is divided in two (a ┬ is set
	// in its top edge there), 0 for no division.
	Divider int
	// Title is set at the right end of the middle row of a tall strip, when it
	// fits (the pane below has no top edge of its own to hold it).
	Title string
}

// The ids of the scroll markers in Spans: a click on them moves to the
// previous or the next tab.
const (
	Prev = "‹"
	Next = "›"
)

// Span is where an id sits in the strip, in columns: x0 inclusive, x1 exclusive.
type Span struct {
	ID     string
	X0, X1 int
}

// Strip is a drawn strip. Lines has three lines when tall and one otherwise;
// the last line of a tall strip is the top edge of the pane below it.
type Strip struct {
	Lines []string
	Spans []Span
}

// At is the id under column x, or "".
func (s Strip) At(x int) string {
	for _, sp := range s.Spans {
		if x >= sp.X0 && x < sp.X1 {
			return sp.ID
		}
	}
	return ""
}
