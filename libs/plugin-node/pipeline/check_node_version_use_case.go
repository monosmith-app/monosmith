package pipeline

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// probe runs a tool for its version; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// CheckNodeVersion compares the installed Node.js with the major CI
// installs. A newer one warns, an older one fails: a dependency's engines
// can need what the machine lacks.
func CheckNodeVersion(workspaceRoot, pin string) protocol.Finding {
	f := protocol.Finding{Check: "Node.js is " + pin + ", as in CI", Warning: true, Remedy: "install Node.js " + pin + " (https://nodejs.org)"}
	got, err := probe(workspaceRoot, []string{"node", "--version"})
	if err != nil {
		f.Warning, f.Detail = false, "node is not on PATH"
		return f
	}
	have := strings.TrimPrefix(strings.TrimSpace(got), "v")
	if found, ok := toolrun.CIPinFinding("Node.js", have, pin); ok {
		found.Remedy = f.Remedy
		return found
	}
	f.Detail = "found Node.js " + have
	return f
}
