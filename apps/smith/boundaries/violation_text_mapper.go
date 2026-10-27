package boundaries

import (
	"fmt"
	"strings"
)

// Text is the violations as the lines a person reads: the edge, the rule and
// the hint, errors and warnings marked apart. No violation reads as one OK line.
func Text(vs []Violation) string {
	if len(vs) == 0 {
		return "boundaries OK: every dependency between projects is allowed\n"
	}
	var b strings.Builder
	for _, v := range vs {
		mark := "✖"
		if v.Warning {
			mark = "!"
		}
		how := "imports"
		if v.Via == "schema" {
			how = "consumes a schema of"
		}
		fmt.Fprintf(&b, "%s %s: %s %s %s\n    rule: %s\n    → %s\n", mark, v.Edge(), v.From, how, v.To, v.Rule, v.Hint)
	}
	errs := len(Errors(vs))
	fmt.Fprintf(&b, "\n%d error(s), %d warning(s)\n", errs, len(vs)-errs)
	return b.String()
}
