package dialog

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// About is the "About" screen: the logo, what smith is for and where its
// documentation is. It closes with Esc, Enter, Space or the OK button.
type About struct {
	version string
	rect    Rect
}

// NewAbout builds the screen for a smith version.
func NewAbout(version string) About { return About{version: version} }

// Place centres the screen on a terminal of the given size.
func (a About) Place(sw, sh int) About {
	w := min(max(sw-4, 10), 72)
	h := len(a.content(theme.Theme{}, w-4)) + 4
	h = min(h, max(sh-2, 3))
	a.rect = Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/2, 0), W: w, H: h}
	return a
}

// Rect is where the screen sits after Place.
func (a About) Rect() Rect { return a.rect }

func centre(s string, w int) string {
	gap := max(w-pane.Width(s), 0)
	return strings.Repeat(" ", gap/2) + s
}

func (a About) content(th theme.Theme, inner int) []string {
	var out []string
	out = append(out, "")
	for _, l := range aboutLogo {
		out = append(out, centre(th.Accent(l), inner))
	}
	title := "MonoSmith"
	if a.version != "" {
		title += " " + a.version
	}
	out = append(out, "", centre(th.Title(title), inner), "")
	for i, p := range aboutText {
		if i > 0 {
			out = append(out, "")
		}
		out = append(out, wrap(p, inner)...)
	}
	out = append(out, "", "Documentation:")
	out = append(out, wrap(aboutDocs, inner)...)
	return out
}

// buttonRow is the row of the OK button inside the box, counted from the top
// of the box.
func (a About) buttonRow() int { return a.rect.H - 2 }

// Lines draws the screen.
func (a About) Lines(th theme.Theme) []string {
	r := a.rect
	if r.W < 8 || r.H < 4 {
		return nil
	}
	body := a.content(th, r.W-4)
	for i := range body {
		body[i] = " " + body[i]
	}
	rows := make([]string, 0, r.H-2)
	rows = append(rows, body[:min(len(body), r.H-3)]...)
	for len(rows) < r.H-3 {
		rows = append(rows, "")
	}
	ok := th.Selected("[ OK ]")
	rows = append(rows, strings.Repeat(" ", max(r.W-2-pane.Width(ok)-2, 0))+ok)
	return pane.Frame(th, "About MonoSmith", rows, r.W, r.H, true)
}

// Key handles a key name; the screen ends on Esc, Enter, Space, q or Ctrl+C.
func (a About) Key(name string) (About, Outcome) {
	switch name {
	case "esc", "enter", " ", "q", "ctrl+c":
		return a, Outcome{Done: true}
	}
	return a, Outcome{}
}

// Click closes the screen on the OK button or a press outside it.
func (a About) Click(x, y int) Outcome {
	if !a.rect.Contains(x, y) {
		return Outcome{Done: true}
	}
	return Outcome{Done: y == a.rect.Y+a.buttonRow() && x >= a.rect.X+a.rect.W-10}
}
