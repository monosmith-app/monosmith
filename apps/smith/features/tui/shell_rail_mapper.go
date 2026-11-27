package tui

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectpage"
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
		PinnedTitle: words.Say("tui.rail.pinned"), Pinned: s.railPins(),
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

// pinPrefix starts the id of a pinned project's row of the rail.
const pinPrefix = "pin:"

// railPins are the pinned projects that are in the workspace, in the order
// they were pinned, each with how its tasks last ran and its kind. A pin of a
// project that is gone is kept, but not drawn.
func (s Shell) railPins() []rail.Pin {
	var out []rail.Pin
	for _, name := range s.pins {
		p, ok := s.list.Project(name)
		if !ok {
			continue
		}
		out = append(out, rail.Pin{ID: pinPrefix + name, Mark: projectpage.Glyph(s.th, s.projectMark(name), s.tick), Label: name, Kind: p.Kind})
	}
	return out
}

// pinOf is the project of a pinned row's id, or "" for a view's id.
func pinOf(id string) string {
	name, _ := strings.CutPrefix(id, pinPrefix)
	if name == id {
		return ""
	}
	return name
}
