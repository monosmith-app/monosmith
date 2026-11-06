package tui

import (
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// openForcedRelease opens the form that releases a package no commit releases.
// It says why when there is no such package: every package has a releasing
// commit, or the plan is not read yet.
func (s *Shell) openForcedRelease() tea.Cmd {
	if !s.release.Loaded || len(s.release.Plan.Unchanged) == 0 {
		s.say(words.Say("tui.forced_release.every_package_in_the_plan_has"))
		return nil
	}
	plan := s.release.Plan
	if forcible := slices.DeleteFunc(slices.Clone(plan.Unchanged), func(n string) bool { return slices.Contains(plan.NeverReleased, n) }); len(forcible) == 0 {
		// One line of the status bar is too short for the commit to make: a notice
		// shows all of it.
		s.tell(words.Say("tui.forced_release.no_package_to_release_without_a"),
			words.Say("tui.forced_release.smith_has_no_version_to_bump", "list", strings.Join(plan.NeverReleased, ", ")),
			words.Say("tui.forced_release.to_choose_its_first_version_make", "neverReleased", plan.NeverReleased[0]))
		return nil
	}
	g := s.guide.WithAction(guide.ForcedReleaseAction(plan.Unchanged, plan.NeverReleased))
	if g, ok := g.Pick(guide.ForcedReleaseTitle); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// isForcedRelease says whether a command line is a release without a releasing commit.
func isForcedRelease(argv []string) bool {
	return len(argv) > 0 && argv[0] == "release" && strings.Contains(" "+strings.Join(argv, " ")+" ", " --force ") // words:literal
}

// forcedReleaseDetails lists, in the confirmation, the packages released
// without a releasing commit and the level asked for each.
func forcedReleaseDetails(argv []string) []string {
	lines := []string{words.Say("tui.forced_release.released")}
	for i, a := range argv {
		if a == "--bump" && i+1 < len(argv) {
			name, level, _ := strings.Cut(argv[i+1], "=")
			lines = append(lines, words.Say("tui.forced_release.bump_of_its_last_release_the", "name", name, "level", level))
		}
	}
	return lines
}
