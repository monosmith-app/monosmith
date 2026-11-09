package shellcompletion

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// shortBody is the most lines of a body that the plan prints one by one. A
// longer body is a whole completion script, which the plan only counts.
const shortBody = 3

// RenderPlan says what install will write: each file that changes, why, and the exact lines
// that go in. It is the text a person reads before the question.
//
// Parameters: p is the plan.
// Returns: the text, one line per entry, with no end of line at the end.
func RenderPlan(p Plan) string {
	var b strings.Builder
	fmt.Fprintln(&b, words.Say("completion.plan.intro", "shell", p.Shell.Title()))
	for _, t := range p.Targets {
		if t.State == Current {
			// Already as it should be: install leaves it alone, so the plan does not list it.
			continue
		}
		fmt.Fprintf(&b, "\n  %s  (%s)\n", t.Path, targetNote(t))
		for _, line := range excerpt(t) {
			fmt.Fprintf(&b, "    %s\n", line)
		}
	}
	fmt.Fprintf(&b, "\n%s", words.Say("completion.plan.undo"))
	return b.String()
}

// targetNote says in words what happens to the file.
func targetNote(t Target) string {
	switch {
	case t.State == Current || t.State == Outdated:
		return words.Say("completion.plan.replace")
	case t.Whole:
		return words.Say("completion.plan.new")
	}
	return words.Say("completion.plan.add")
}

// excerpt is what goes into the file, with a long script cut to a count.
func excerpt(t Target) []string {
	var out []string
	out = append(out, beginMarker, managedNote)
	body := strings.Split(normalise(t.Body), "\n")
	if len(body) <= shortBody {
		out = append(out, body...)
	} else {
		out = append(out, words.Say("completion.plan.script", "lines", len(body)))
	}
	return append(out, endMarker)
}
