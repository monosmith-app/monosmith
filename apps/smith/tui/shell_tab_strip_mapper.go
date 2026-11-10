package tui

import (
	"fmt"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/tabs"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// shortTitles are the shorter tab names, used only for as many tabs as it takes
// for the strip to fit.
func shortTitles() map[string]string {
	return map[string]string{
		commands.TabProjects: words.Say("tui.tab.projects_short"), commands.TabRelease: words.Say("tui.tab.release_short"), commands.TabSettings: words.Say("tui.tab.settings_short"),
	}
}

// shortOrder is the order tabs give up their full name in: the ones people
// meet last go first, so "Projects" is the last to be shortened.
var shortOrder = []string{commands.TabSettings, commands.TabRelease, commands.TabProjects}

// badgeRoom is what a status badge (spinner, check, cross and count) may add
// to a tab. It is reserved in every state, so a badge appearing never changes
// which names are spelled out.
const badgeRoom = 3

// tabStrip is the row of views.
func (s Shell) tabStrip() tabs.Strip { return s.renderStrip(s.hoveredTab()) }

// headerReserve is the width the header takes from the tabs when it shares
// their row.
func (s Shell) headerReserve() int {
	if right := s.headerRight(); right != "" {
		return pane.Width(right) + 2
	}
	return 0
}

// renderStrip draws the strip with the tab hover highlighted, in the width the
// header leaves it.
func (s Shell) renderStrip(hover string) tabs.Strip {
	return s.stripAt(s.w-s.headerReserve(), hover)
}

// stripAt draws the strip in a width.
func (s Shell) stripAt(width int, hover string) tabs.Strip {
	list := make([]tabs.Tab, 0, len(s.tabs))
	for _, id := range s.tabs {
		t := tabs.Tab{ID: id, Label: tabTitles()[id], Short: shortTitles()[id], Key: commands.ShortcutOf("view." + id), Badge: s.tabBadge(id)}
		if name := s.guideModalName(); id == commands.TabGuide && name != "" {
			// A form that was answered and waits for its confirmation, or a box
			// that took its place: the tab names it, not "Start".
			t.Label = name
		}
		if id == commands.TabProjects || id == commands.TabDoctor {
			t.Reserve = badgeRoom
		}
		list = append(list, t)
	}
	return tabs.Render(tabs.Props{
		Theme: s.th, Tabs: list, Active: s.tab, Hover: hover, Width: width,
		ShortenOrder: shortOrder,
	})
}

// tabBadge is the mark after a tab's name: how the run is going (Projects, which
// holds the run panel), what doctor found.
func (s Shell) tabBadge(id string) string {
	if id != commands.TabProjects && s.tabBusy(id) {
		// Work for the tab goes on in the background: a spinner says so.
		return s.th.Info(s.th.Spin(s.tick))
	}
	switch id {
	case commands.TabProjects:
		switch {
		case s.running():
			return s.th.Info(s.th.Spin(s.tick))
		case s.run != nil:
			if n := s.run.Counts().Failed; n > 0 {
				return s.th.Failure(fmt.Sprintf("%s%d", s.th.Icons.Failed, n))
			} else if s.run.OK() && slices.Contains(s.run.Command, "--dry-run") {
				// A dry run changed nothing: a tick would say that something shipped.
				return s.th.Muted(s.th.Icons.Skipped)
			} else if s.run.OK() {
				return s.th.Success(s.th.Icons.Passed)
			} else if s.run.Cancelled {
				return s.th.Muted(s.th.Icons.Cancelled)
			}
		}
	case commands.TabDoctor:
		// The same marks as the header: × counts failures only, ▲ warnings.
		if r := s.doctor.Report; r != nil {
			if r.Failed > 0 {
				return s.th.Failure(fmt.Sprintf("%s%d", s.th.Icons.Failed, r.Failed))
			}
			if r.Warnings > 0 {
				return s.th.Warning(fmt.Sprintf("%s%d", s.th.Icons.Warning, r.Warnings))
			}
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

// tabBusy is true while the work that fills a tab goes on in the background:
// the doctor checks, the registries answering, the release plan being read.
func (s Shell) tabBusy(id string) bool {
	switch id {
	case commands.TabDoctor:
		return s.doctor.Loading
	case commands.TabDeps:
		return s.depsLoading
	case commands.TabRelease:
		return s.releaseLoading
	}
	return false
}

// backgroundWork is true while any tab has work in the background, so its
// badge keeps moving.
func (s Shell) backgroundWork() bool {
	return s.tabBusy(commands.TabDoctor) || s.tabBusy(commands.TabDeps) || s.tabBusy(commands.TabRelease)
}
