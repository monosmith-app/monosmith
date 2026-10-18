package pipeline

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// runtime installs Python in the generated pipelines.
func runtime(version string) protocol.Toolchain {
	return protocol.Toolchain{
		Name: "python", Version: version,
		GitHub: &protocol.ActionStep{Uses: "actions/setup-python@v7", Inputs: []protocol.Input{{Name: "python-version", Value: version}}},
		Azure:  &protocol.ActionStep{Uses: "UsePythonVersion@0", Inputs: []protocol.Input{{Name: "versionSpec", Value: version}}},
		// Elsewhere (GitLab CI, on a Debian image, as root): uv installs the
		// exact version, linked where smith looks for it.
		Install: "curl -LsSf https://astral.sh/uv/install.sh | env UV_INSTALL_DIR=/usr/local/bin sh\n" +
			"uv python install " + version + "\n" +
			"ln -sf \"$(uv python find " + version + ")\" /usr/local/bin/python" + version + "\n" +
			"ln -sf /usr/local/bin/python" + version + " /usr/local/bin/python3",
		Devcontainer: &protocol.DevcontainerFeature{
			ID: "ghcr.io/devcontainers/features/python:1", Options: map[string]string{"version": version},
		},
	}
}
