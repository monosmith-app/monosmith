package depsview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// PolicyItems are the lines that tell the verdict of the one-version policy:
// a mark and the check on each line, and under a failed check the detail of
// what differs. The detail beside the list shows them under the heading
// PolicyTitle; a narrow screen shows them under the list.
func (v Model) PolicyItems(th theme.Theme) []string {
	if len(v.Policy) == 0 {
		return []string{th.Muted(words.Say("depsview.policy.empty"))}
	}
	var items []string
	for _, p := range v.Policy {
		mark := th.Success(th.Icons.Passed)
		switch {
		case !p.OK && p.Warning:
			mark = th.Warning(th.Icons.Warning)
		case !p.OK:
			mark = th.Failure(th.Icons.Failed)
		}
		items = append(items, mark+" "+p.Check)
		if !p.OK && p.Detail != "" {
			for _, d := range strings.Split(p.Detail, "; ") {
				items = append(items, "    "+th.Muted(d))
			}
		}
	}
	return items
}

// PolicyTitle is the heading over the policy lines.
func PolicyTitle() string { return words.Say("depsview.policy.title") }
