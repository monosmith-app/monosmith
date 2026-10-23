package menubar

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestDropdownRowsAreAsWideAsItsBorders(t *testing.T) {
	d := dropdown{entries: []Entry{
		{ID: "a", Name: "New workspace…", Shortcut: "Ctrl+N"},
		{Separator: true},
		// The widest entry has no shortcut: its row must not outgrow the box.
		{ID: "b", Name: "Convert an Nx or Turborepo workspace…"},
		{ID: "c", Name: "Quit", Shortcut: "Ctrl+Q"},
	}}
	rows := d.view(theme.Default())
	want := pane.Width(rows[0])
	for i, row := range rows {
		if got := pane.Width(row); got != want {
			t.Errorf("row %d is %d cells wide, the top edge is %d", i, got, want)
		}
	}
}
