package tui

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
)

// openForcedRelease opens the form that releases a package no commit releases.
// It says why when there is no such package: every package has a releasing
// commit, or the plan is not read yet.
func (s *Shell) openForcedRelease() tea.Cmd {
	if !s.release.Loaded || len(s.release.Plan.Unchanged) == 0 {
		s.say("Every package in the plan has a releasing commit, or the plan is not read yet")
		return nil
	}
	plan := s.release.Plan
	if forcible := slices.DeleteFunc(slices.Clone(plan.Unchanged), func(n string) bool { return slices.Contains(plan.NeverReleased, n) }); len(forcible) == 0 {
		// One line of the status bar is too short for the commit to make: a notice
		// shows all of it.
		s.tell("No package to release without a commit",
			"smith has no version to bump for a package that was never released: "+strings.Join(plan.NeverReleased, ", ")+".",
			"To choose its first version, make this commit:\n  git commit --allow-empty -m \"version("+plan.NeverReleased[0]+")[1.0.0]\"")
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
	return len(argv) > 0 && argv[0] == "release" && strings.Contains(" "+strings.Join(argv, " ")+" ", " --force ")
}

// forcedReleaseDetails lists, in the confirmation, the packages released
// without a releasing commit and the level asked for each.
func forcedReleaseDetails(argv []string) []string {
	lines := []string{"Released although no commit releases it:"}
	for i, a := range argv {
		if a == "--bump" && i+1 < len(argv) {
			name, level, _ := strings.Cut(argv[i+1], "=")
			lines = append(lines, fmt.Sprintf("  %s  %s bump of its last release, the same code", name, level))
		}
	}
	return lines
}
