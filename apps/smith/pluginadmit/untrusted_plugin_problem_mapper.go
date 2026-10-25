package pluginadmit

import (
	"errors"
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// ErrUntrusted is the cause of the problem that refuses to start an external
// plugin that is not trusted.
var ErrUntrusted = errors.New("an external plugin is not trusted")

// Refusal explains, as a problem with its remedy, why an external plugin was
// not started.
func Refusal(st Status, inCI bool) *problem.Error {
	trust := fmt.Sprintf("smith plugin trust %s", st.Name)
	var p *problem.Error
	switch {
	case st.Problem != "":
		p = problem.New(fmt.Sprintf("the external plugin %q cannot be started: %s", st.Name, st.Problem),
			"check its command in smith.toml or its plugin.json, then run `smith plugin list` to see what smith would start")
	case inCI && st.Reason == plugintrust.ReasonPinMismatch:
		p = problem.New(fmt.Sprintf("the external plugin %q does not match the sha256 pinned in smith.toml, so smith did not start it (%s)", st.Name, st.Detail),
			"if you changed the plugin on purpose, run `"+trust+" --pin` on your machine and commit the new sha256")
	case inCI:
		p = problem.New(fmt.Sprintf("the external plugin %q is not pinned, so smith did not start it: in CI an external plugin runs only when smith.toml pins its sha256", st.Name),
			"run `"+trust+" --pin` on your machine and commit the sha256 it writes to smith.toml")
	case st.Reason == plugintrust.ReasonCommandChanged:
		p = problem.New(fmt.Sprintf("the external plugin %q now runs a different command than the one you trusted, so smith did not start it", st.Name),
			"run `"+trust+"` to review it and trust it again")
	case st.Reason == plugintrust.ReasonFileChanged:
		p = problem.New(fmt.Sprintf("the external plugin %q changed since you trusted it, so smith did not start it (%s)", st.Name, st.Detail),
			"run `"+trust+"` to review it and trust it again")
	default:
		p = problem.New(fmt.Sprintf("the external plugin %q is not trusted yet, so smith did not start it", st.Name),
			"run `"+trust+"` to review it and trust it", "run `smith plugin list` to see every external plugin and what it may see")
	}
	p.Cause = ErrUntrusted
	return p
}
