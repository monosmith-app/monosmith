package pluginadmit

import (
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/pluginenv"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
)

// Describe works out the status of each plugin in a workspace. A command that
// cannot be measured is reported in Problem, not as an error.
func Describe(store *plugintrust.Store, workspace string, plugins []Plugin) []Status {
	inCI := plugintrust.InCI(os.Getenv)
	out := make([]Status, 0, len(plugins))
	for _, p := range plugins {
		st := Status{Plugin: p, Env: pluginenv.Visible(p.ManifestEnv, p.WorkspaceEnv), Pinned: p.Pin != ""}
		d, err := plugintrust.Check(store, plugintrust.Request{Workspace: workspace, Plugin: p.Name, Command: p.Command, Pin: p.Pin, InCI: inCI})
		if err != nil {
			st.Problem = err.Error()
			st.Reason, st.Detail = plugintrust.ReasonNew, err.Error()
			out = append(out, st)
			continue
		}
		st.Trusted, st.Reason, st.Detail, st.Fingerprint = d.Trusted, d.Reason, d.Detail, d.Fingerprint
		out = append(out, st)
	}
	return out
}
