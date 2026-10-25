package pluginadmit

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Findings turns the status of each external plugin the workspace needs
// into a finding: what it is, where it comes from, whether it is pinned and
// the variables it may see; a plugin smith would refuse to start fails, with
// the command that fixes it.
func Findings(statuses []Status, inCI bool) []protocol.Finding {
	var out []protocol.Finding
	for _, st := range statuses {
		check := fmt.Sprintf("external plugin %s is trusted (from %s, pinned %s, sees %s)", st.Name, st.Source, yesNo(st.Pinned), seen(st.Env))
		f := protocol.Finding{Check: check, OK: st.Trusted && st.Problem == ""}
		if !f.OK {
			refusal := Refusal(st, inCI)
			f.Check = fmt.Sprintf("external plugin %s is trusted", st.Name)
			f.Detail = refusal.Problem
			f.Remedy = strings.Join(refusal.Remedy, "\n")
			if st.Reason == plugintrust.ReasonNew && st.Problem == "" {
				f.Detail += fmt.Sprintf(" (from %s, pinned %s, sees %s)", st.Source, yesNo(st.Pinned), seen(st.Env))
			}
		}
		out = append(out, f)
	}
	return out
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func seen(env []string) string {
	if len(env) == 0 {
		return "the baseline environment only"
	}
	return "the baseline plus " + strings.Join(env, ", ")
}
