package adoptmove

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// ClaimsOf turns the handshakes of the started plugins into the claims Detect
// reads: one for each plugin that discovers projects by a manifest, named by
// its first language. Plugins that discover nothing (conventions, deploy
// targets) or name no manifest are left out. The order of the plugins is kept.
//
// Parameters: plugins are the plugins' initialize results.
// Returns: the claims, empty when no plugin discovers projects.
func ClaimsOf(plugins []protocol.InitializeResult) []Claim {
	var out []Claim
	for _, p := range plugins {
		if !p.Capabilities.Discover || len(p.Manifests) == 0 || len(p.Languages) == 0 {
			continue
		}
		out = append(out, Claim{Language: p.Languages[0], Manifests: p.Manifests})
	}
	return out
}
