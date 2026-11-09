package adoptwizard

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// planSummaryText is the top of the review of a conversion, before the files:
// what it comes to in numbers, then the things to check (all of them), then
// what behaves differently afterwards (the first few; the rest are in the
// report that is saved once it is converted). Lines start with a bullet, never
// with + or -.
func planSummaryText(c adoptflow.Changes, r convert.Highlights) string {
	var b strings.Builder
	b.WriteString(words.Say("adopt.summary.short"))
	if c.Total() > 0 {
		b.WriteString(words.Say("adopt.summary.files", "created", len(c.Created), "edited", len(c.Edited), "removed", len(c.Removed)))
	}
	b.WriteString(words.Say("adopt.summary.checks", "count", countOrNone(len(r.Checks), words.Say("adopt.summary.below"))))
	b.WriteString(words.Say("adopt.summary.changed", "count", countOrNone(len(r.Changed), words.Say("adopt.summary.below"))))
	if len(r.Checks) > 0 {
		b.WriteString(words.Say("adopt.summary.checkhead", "count", len(r.Checks)))
		for _, x := range r.Checks {
			b.WriteString("  • " + x + "\n")
		}
	}
	if len(r.Changed) > 0 {
		b.WriteString(words.Say("adopt.summary.changedhead", "count", len(r.Changed)))
		for i, x := range r.Changed {
			if i == shownChanged {
				b.WriteString(words.Say("adopt.summary.more", "count", len(r.Changed)-shownChanged, "file", convert.ReportFile))
				break
			}
			b.WriteString("  • " + x + "\n")
		}
	}
	return b.String()
}

func countOrNone(n int, where string) string {
	if n == 0 {
		return words.Say("adopt.summary.none")
	}
	return words.Say("adopt.summary.count", "count", n, "where", where)
}
