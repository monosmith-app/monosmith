package pipeline

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// runtime installs Node.js in the generated pipelines.
func runtime(version string) protocol.Toolchain {
	return protocol.Toolchain{
		Name: "node", Version: version,
		GitHub: &protocol.ActionStep{Uses: "actions/setup-node@v7", Inputs: []protocol.Input{{Name: "node-version", Value: version}}},
		Azure:  &protocol.ActionStep{Uses: "UseNode@1", Inputs: []protocol.Input{{Name: "version", Value: version + ".x"}}},
		// Elsewhere (GitLab CI, on a Debian image, as root): NodeSource.
		Install: "curl -fsSL https://deb.nodesource.com/setup_" + version + ".x | bash -\napt-get install -y -qq nodejs",
		Devcontainer: &protocol.DevcontainerFeature{
			ID: "ghcr.io/devcontainers/features/node:1", Options: map[string]string{"version": version},
		},
	}
}
