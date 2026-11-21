package trustedpublishing

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

// Pending lists the options smith has to offer: the workspace's pipelines ask
// for the option's secret, smith.toml does not decide the option, and every CI
// provider is GitHub, the only one with these sign-ins.
//
// Parameters: cfg is the workspace configuration; secrets are the credential
// names the generated pipelines pass (before any keyless sign-in drops one).
// Returns: the undecided options, in the order of Options.
func Pending(cfg workspace.Config, secrets []string) []Option {
	if len(cfg.CI.Providers) == 0 || slices.ContainsFunc(cfg.CI.Providers, func(p string) bool { return p != "github" }) {
		return nil
	}
	var out []Option
	for _, o := range Options {
		if !o.Decided(cfg.Release) && slices.Contains(secrets, o.Secret) {
			out = append(out, o)
		}
	}
	return out
}
