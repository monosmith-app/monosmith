package boundaries

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Check returns every dependency in the graph that breaks a rule of cfg, by
// the project that depends and then the one it depends on. A dependency an
// exception names breaks nothing. With no rules and the built-in one met, it
// returns nothing.
func Check(g *projectgraph.Graph, cfg workspace.BoundariesSection) []Violation {
	var out []Violation
	for _, from := range g.Projects {
		for _, name := range from.Deps {
			to := g.Get(name)
			if to == nil || excepted(from, to, cfg.Exceptions) {
				continue
			}
			via := "import"
			if slices.Contains(from.SchemaDeps, name) {
				via = "schema"
			}
			for _, rule := range cfg.Rules {
				if why, broken := breaks(rule, from, to); broken {
					out = append(out, newViolation(from, to, why, via, false))
				}
			}
			if to.Kind == protocol.ProjectApp && (cfg.Kinds == nil || *cfg.Kinds) {
				out = append(out, newViolation(from, to, "an app is deployed, so no project depends on it", via, cfg.Kinds == nil))
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

// Errors are the violations that fail a check; the rest are warnings.
func Errors(vs []Violation) []Violation {
	var out []Violation
	for _, v := range vs {
		if !v.Warning {
			out = append(out, v)
		}
	}
	return out
}

func excepted(from, to *projectgraph.Project, exceptions []workspace.BoundaryException) bool {
	return slices.ContainsFunc(exceptions, func(e workspace.BoundaryException) bool {
		return from.Matches(e.From) && to.Matches(e.To)
	})
}

// breaks says whether the dependency breaks the rule, and how.
func breaks(rule workspace.BoundaryRule, from, to *projectgraph.Project) (string, bool) {
	if !from.Matches(rule.From) {
		return "", false
	}
	why := ""
	if rule.Reason != "" {
		why = " (" + rule.Reason + ")"
	}
	if i := slices.IndexFunc(rule.Deny, to.Matches); i >= 0 {
		return fmt.Sprintf("projects matching %q may not depend on %q%s", rule.From, rule.Deny[i], why), true
	}
	if len(rule.Allow) > 0 && !slices.ContainsFunc(rule.Allow, to.Matches) {
		return fmt.Sprintf("projects matching %q may only depend on %s%s", rule.From, quoted(rule.Allow), why), true
	}
	return "", false
}

func quoted(selectors []string) string {
	q := make([]string, len(selectors))
	for i, s := range selectors {
		q[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(q, ", ")
}

func newViolation(from, to *projectgraph.Project, rule, via string, warning bool) Violation {
	v := Violation{From: from.Name, To: to.Name, Rule: rule, Via: via, Warning: warning}
	remove := fmt.Sprintf("remove %s's import of %s", from.Name, to.Name)
	if via == "schema" {
		remove = fmt.Sprintf("stop %s consuming a schema that %s owns", from.Name, to.Name)
	}
	v.Hint = remove + ", or move what both need into a lib that neither depends back on; if the dependency is on purpose, add [[boundaries.exceptions]] with from = \"" +
		from.Name + "\", to = \"" + to.Name + "\" and a reason to smith.toml"
	return v
}
