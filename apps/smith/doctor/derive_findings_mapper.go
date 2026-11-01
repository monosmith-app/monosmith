package doctor

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// deriveChecks reports the derived files that are out of date or edited by
// hand, with their owner. It reports nothing when the workspace did not turn
// derived files on.
func deriveChecks(s *session.Session) []protocol.Finding {
	if !s.Config.Derive.Enabled {
		return nil
	}
	plan, _, err := derive.BuildPlan(s)
	if err != nil {
		return []protocol.Finding{{Check: "derived files can be checked", Detail: err.Error(), Remedy: "fix the problem above, then run smith derive --check"}}
	}
	return deriveFindings(plan)
}

// deriveFindings maps a plan to findings: one per file that needs action, or
// one passing finding when none does.
func deriveFindings(plan derive.Plan) []protocol.Finding {
	var out []protocol.Finding
	for _, i := range plan.Pending() {
		f := protocol.Finding{
			Check:  "derived file " + i.File.Path + " (" + i.File.Renderer + ") is up to date",
			Detail: fmt.Sprintf("%s: %s (owner: %s)", i.Status, i.Detail, i.Owner),
			Remedy: "run smith derive",
		}
		switch i.Status {
		case derive.StatusDrift, derive.StatusDiverged:
			f.Remedy = "see the diff with smith derive --explain " + i.File.Path + ", then choose smith derive --replace, --adopt or --eject " + i.File.Path
		case derive.StatusBroken:
			f.Remedy = "fix the file, or give it to yourself with smith derive --eject " + i.File.Path
		}
		out = append(out, f)
	}
	if len(out) == 0 {
		out = append(out, protocol.Finding{Check: "derived files are up to date", OK: true})
	}
	if n := len(plan.Diverging); n > 0 {
		// A file that differs from its siblings is the project's own, so this
		// passes: it tells, it does not ask for an action.
		out = append(out, protocol.Finding{
			Check:  "config files that differ from their sibling projects are yours",
			Detail: fmt.Sprintf("%d files diverge from siblings (first: %s); smith leaves them alone", n, plan.Diverging[0].Path),
			Remedy: "smith derive --explain <file> says how a file differs",
			OK:     true,
		})
	}
	return out
}
