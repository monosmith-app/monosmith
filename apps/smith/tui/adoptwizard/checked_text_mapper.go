package adoptwizard

import (
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
)

// checkedText says what the closing checks proved, in the order they ran:
// "everything builds (9.8s), every test passes (31s)". Steps that are not a
// check (the installs) are left out; it is "" when no check ran.
func checkedText(steps []checkedStep) string {
	words := map[string]string{
		removalplan.StepBuild:  "everything builds",
		removalplan.StepTest:   "every test passes",
		removalplan.StepLint:   "lint is clean",
		removalplan.StepDoctor: "doctor finds nothing wrong",
	}
	var parts []string
	for _, s := range steps {
		if w, ok := words[s.Name]; ok {
			parts = append(parts, fmt.Sprintf("%s (%s)", w, tookText(s.Took)))
		}
	}
	return strings.Join(parts, ", ")
}

// tookText is a duration a person reads: 9.8s, 31s, 2m05s.
func tookText(d time.Duration) string {
	switch {
	case d < time.Second:
		return "under a second"
	case d < 10*time.Second:
		return fmt.Sprintf("%.1fs", d.Seconds())
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Round(time.Second).Seconds()))
	}
	d = d.Round(time.Second)
	return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
}
