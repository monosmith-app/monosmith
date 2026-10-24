package tui

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/tabs"
)

// shortTitles are the shorter tab names, used only for as many tabs as it takes
// for the strip to fit.
var shortTitles = map[string]string{
	commands.TabProjects: "Proj", commands.TabRelease: "Rel", commands.TabPlugins: "Plug", commands.TabSettings: "Sett",
}

// shortOrder is the order tabs give up their full name in: the ones people
// meet last go first, so "Projects" is the last to be shortened.
var shortOrder = []string{commands.TabSettings, commands.TabPlugins, commands.TabRelease, commands.TabProjects}

// badgeRoom is what a status badge (spinner, check, cross and count) may add
// to a tab. It is reserved in every state, so a badge appearing never changes
// which names are spelled out.
const badgeRoom = 3

// tabStrip is the row of views, three rows high on a tall terminal.
func (s Shell) tabStrip() tabs.Strip { return s.renderStrip(s.hoveredTab()) }

// renderStrip draws the strip with the tab hover highlighted.
func (s Shell) renderStrip(hover string) tabs.Strip {
	l := s.lay()
	list := make([]tabs.Tab, 0, len(s.tabs))
	for _, id := range s.tabs {
		t := tabs.Tab{ID: id, Label: tabTitles[id], Short: shortTitles[id], Badge: s.tabBadge(id)}
		if id == commands.TabRun || id == commands.TabDoctor {
			t.Reserve = badgeRoom
		}
		list = append(list, t)
	}
	divider := 0
	if l.HasDetail() {
		divider = l.Main.W - 1
	}
	title := ""
	if l.Tall {
		title = s.page().Title()
	}
	return tabs.Render(tabs.Props{
		Theme: s.th, Tabs: list, Active: s.tab, Hover: hover, Width: l.Tabs.W, Tall: l.Tall,
		ShortenOrder: shortOrder, Title: title, Divider: divider,
	})
}

// tabsInBody is how many rows of the strip belong to the content below it: the
// last one of a tall strip is the content's top edge.
func (s Shell) tabsInBody() int {
	if s.lay().Tall {
		return 1
	}
	return 0
}

// tabBadge is the mark after a tab's name: how the run is going, what doctor found.
func (s Shell) tabBadge(id string) string {
	switch id {
	case commands.TabRun:
		switch {
		case s.running():
			return s.th.Info(s.th.Spin(s.tick))
		case s.run != nil:
			if n := s.run.Counts().Failed; n > 0 {
				return s.th.Failure(fmt.Sprintf("%s%d", s.th.Icons.Failed, n))
			} else if s.run.OK() {
				return s.th.Success(s.th.Icons.Passed)
			}
		}
	case commands.TabDoctor:
		if s.doctor.Report != nil && s.doctor.Report.Failed > 0 {
			return s.th.Failure(fmt.Sprintf("%s%d", s.th.Icons.Failed, s.doctor.Report.Failed))
		}
	}
	return ""
}

// hoveredTab is the tab under the pointer, or "".
func (s Shell) hoveredTab() string {
	l := s.lay()
	if !l.Tabs.Contains(s.mouseX, s.mouseY) {
		return ""
	}
	id := s.renderStrip("").At(s.mouseX)
	if id == tabs.Prev || id == tabs.Next {
		return ""
	}
	return id
}
