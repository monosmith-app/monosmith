package tui

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
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
	b.WriteString("Nothing is written yet. ")
	if i := slices.Index(argv, "--with"); i >= 0 && i+1 < len(argv) {
		names := strings.Split(argv[i+1], ",")
		words := strings.Join(names, ", ")
		if l := slices.Index(argv, "--lang"); l >= 0 && l+1 < len(argv) {
			if said := guide.ChoicesInWords(offers, argv[l+1], names); said != "" {
				words = said
			}
		}
		b.WriteString("It is built from your choices: " + words + ". ")
	} else {
		b.WriteString("It takes the stack's own defaults. ")
	}
	switch {
	case slices.Contains(argv, "--fresh"):
		b.WriteString("It does not follow the other projects. ")
	case projects >= siblingsFollowedFrom:
		b.WriteString("It follows the other projects of its kind: the plan names what it takes from them. ")
	}
	b.WriteString("Run makes the project and Esc leaves everything as it is. Tick the box to commit it, so it can be undone.")
	return b.String()
}
