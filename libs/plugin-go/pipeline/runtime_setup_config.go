package pipeline

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// runtime installs Go in the generated pipelines. The Go module cache is not
// cached by the action: smith's own cache covers the tasks.
func runtime(version string) protocol.Toolchain {
	azure := version
	if strings.Count(azure, ".") == 1 {
		azure += ".0" // GoTool wants a full version
	}
	return protocol.Toolchain{
		Name: "go", Version: version,
		GitHub: &protocol.ActionStep{Uses: "actions/setup-go@v7", Inputs: []protocol.Input{{Name: "go-version", Value: version}, {Name: "cache", Value: "false"}}},
		Azure:  &protocol.ActionStep{Uses: "GoTool@0", Inputs: []protocol.Input{{Name: "version", Value: azure}}},
		// Elsewhere (GitLab CI, on a Debian image, as root): the release
		// tarball, on PATH for the rest of the job.
		Install: "curl -fsSL https://go.dev/dl/go" + azure + ".linux-amd64.tar.gz | tar -C /usr/local -xz\n" +
			"export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH",
		// The feature would otherwise install the newest golangci-lint.
		Devcontainer: &protocol.DevcontainerFeature{
			ID:      "ghcr.io/devcontainers/features/go:1",
			Options: map[string]string{"version": azure, "golangciLintVersion": strings.TrimPrefix(GolangciLintVersion, "v")},
		},
	}
}
