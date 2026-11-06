package tui

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// siblingsFollowedFrom is how many projects a workspace needs before smith add
// follows the ones that are like the new project.
const siblingsFollowedFrom = 3

// addPlanWhy is the sentence above the plan of an add: what the project is
// built from (the choices named with --with, or the stack's defaults) and
// whether it follows the other projects. projects is how many the workspace has;
// offers are the choices smith lists, to say the picked ones in words.
func addPlanWhy(argv []string, projects int, offers []guide.ChoiceOffer) string {
	var b strings.Builder
	b.WriteString(words.Say("tui.add_plan_text.nothing_is_written_yet"))
	if i := slices.Index(argv, "--with"); i >= 0 && i+1 < len(argv) {
		names := strings.Split(argv[i+1], ",")
		choices := strings.Join(names, ", ")
		if l := slices.Index(argv, "--lang"); l >= 0 && l+1 < len(argv) {
			if said := guide.ChoicesInWords(offers, argv[l+1], names); said != "" {
				choices = said
			}
		}
		b.WriteString(words.Say("tui.add_plan_text.it_is_built_from_your_choices", "choices", choices))
	} else {
		b.WriteString(words.Say("tui.add_plan_text.it_takes_the_stack_s_own"))
	}
	switch {
	case slices.Contains(argv, "--fresh"):
		b.WriteString(words.Say("tui.add_plan_text.it_does_not_follow_the_other"))
	case projects >= siblingsFollowedFrom:
		b.WriteString(words.Say("tui.add_plan_text.it_follows_the_other_projects_of"))
	}
	b.WriteString(words.Say("tui.add_plan_text.run_makes_the_project_and_esc"))
	return b.String()
}
