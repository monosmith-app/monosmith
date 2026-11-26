package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/rail"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// railViews are the rows of the rail: the views of the strip, in its order,
// each with its badge and its key.
func (s Shell) railViews() []rail.View {
	out := make([]rail.View, 0, len(s.tabs))
	for _, id := range s.tabs {
		out = append(out, rail.View{ID: id, Label: s.viewName(id), Key: commands.ShortcutOf("view." + id), Badge: s.tabBadge(id)})
	}
	return out
}

// renderRail draws the rail of views with hover lit; it is empty on a narrow
// screen, which has the strip instead.
func (s Shell) renderRail(hover string) rail.Rail {
	l := s.lay()
	if l.Rail.W == 0 {
		return rail.Rail{}
	}
	return rail.Render(rail.Props{
		Theme: s.th, Title: words.Say("tui.rail.title"), Views: s.railViews(),
		Active: s.tab, Hover: hover, Width: l.Rail.W, Height: l.Rail.H,
	})
}

// railAt is the view on screen row y of the rail, or "".
func (s Shell) railAt(y int) string { return s.renderRail("").At(y - s.lay().Rail.Y) }

// hoveredRail is the view of the rail under the pointer, or "".
func (s Shell) hoveredRail() string {
	if !s.lay().Rail.Contains(s.mouseX, s.mouseY) {
		return ""
	}
	return s.railAt(s.mouseY)
}
