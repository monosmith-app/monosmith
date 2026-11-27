package rail

import "github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"

// View is one row of the rail.
type View struct {
	ID string
	// Label is the view's name.
	Label string
	// Key is the key that opens the view, drawn at the right end of the row
	// ("" for none).
	Key string
	// Badge is a short styled state mark before the key (a failure count, a
	// spinner), "" for none.
	Badge string
}

// Pin is one project pinned to the rail, under the views.
type Pin struct {
	// ID is what a click on the row gives back; it never equals a view's ID.
	ID string
	// Mark is the styled mark of how the project last ran; Label is its name
	// and Kind its kind, drawn muted at the right end.
	Mark, Label, Kind string
}

// Props is everything a rail is drawn from.
type Props struct {
	Theme theme.Theme
	// Title is set into the top border of the frame.
	Title string
	Views []View
	// PinnedTitle is the heading of the pinned projects, and Pinned the
	// projects, in order. A rail without pins draws no heading.
	PinnedTitle string
	Pinned      []Pin
	// Active is the id of the open view, Hover the id of the view under the
	// pointer ("" for none).
	Active, Hover string
	// Width and Height are the cells of the rail, its frame included.
	Width, Height int
}

// Rail is a drawn rail: its lines, and the view each line stands for.
type Rail struct {
	Lines []string
	// Rows maps a line of Lines to the id of the view drawn on it.
	Rows map[int]string
}

// At is the id of the view on line y of the rail, or "".
//
// Parameters: y is the line, counted from the rail's top border.
// Returns: the view's id, "" for a line that holds no view.
func (r Rail) At(y int) string { return r.Rows[y] }
