package adoptwizard

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
)

func TestThePlanSummaryLeadsWithNumbersThenTheChecksThenTheChanges(t *testing.T) {
	var changed []string
	for i := 1; i <= 7; i++ {
		changed = append(changed, fmt.Sprintf("change %d", i))
	}
	text := planSummaryText(
		adoptflow.Changes{Created: []string{"smith.toml"}, Edited: []string{"a", "b"}, Removed: []string{"nx.json", "jest.config.ts", "c"}},
		convert.Highlights{Checks: []string{"ESLint config needs a look", "a script needs oxfmt"}, Changed: changed},
	)
	order := []string{"Files: smith creates 1, edits 2 and removes 3.", "Things to check: 2", "Things that behave differently afterwards: 7", "! Check these before you go ahead (2)", "a script needs oxfmt", "Behaves differently afterwards (7)", "change 5", "and 2 more", "once it is converted"}
	at := 0
	for _, want := range order {
		i := strings.Index(text[at:], want)
		if i < 0 {
			t.Fatalf("%q missing or out of order:\n%s", want, text)
		}
		at += i
	}
	if strings.Contains(text, "change 6") {
		t.Errorf("only the first few changes are named:\n%s", text)
	}
}

func TestThePlanSummaryOfAnEasyConversionSaysThereIsNothingToCheck(t *testing.T) {
	text := planSummaryText(adoptflow.Changes{}, convert.Highlights{})
	if !strings.Contains(text, "Things to check: none") || strings.Contains(text, "Check these") {
		t.Errorf("text:\n%s", text)
	}
}

func TestNoLineOfAReviewStartsWithAPlusOrAMinus(t *testing.T) {
	r := convert.Highlights{Checks: []string{"x"}, Changed: []string{"y"}}
	c := adoptflow.Changes{Created: []string{"smith.toml"}, Removed: []string{"nx.json"}}
	f := adoptflow.Found{Kind: adoptflow.KindNx, Source: "Nx workspace", Name: "acme"}
	for _, text := range []string{planSummaryText(c, r), conservativeText(f, "", c, r)} {
		for _, line := range strings.Split(text, "\n") {
			if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
				t.Errorf("the review would paint %q as a diff", line)
			}
		}
	}
}
