package functions

import (
	"os/exec"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// lookPath is replaced by tests.
var lookPath = exec.LookPath

// Findings checks, for each host the workspace's function apps use, that the
// tool that runs them locally is installed. A missing tool is a warning: the
// build and the tests need no cloud tool.
func Findings(workspaceRoot string, projects []protocol.Project) []protocol.Finding {
	seen := map[string]bool{}
	var findings []protocol.Finding
	for _, p := range projects {
		host := DetectHost(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)))
		tool, ok := functionhost.ToolOf(host)
		if host == "" || seen[host] || !ok {
			continue
		}
		seen[host] = true
		installed := frameworkInstalled(workspaceRoot)
		if tool.Binary != "" {
			_, err := lookPath(tool.Binary)
			installed = err == nil
		}
		findings = append(findings, functionhost.FindingFor(tool, installed))
	}
	return findings
}

// frameworkInstalled reports whether the workspace venv holds the Google
// Functions Framework, the local tool of that host.
func frameworkInstalled(workspaceRoot string) bool {
	for _, pattern := range []string{".venv/Lib/site-packages/functions_framework", ".venv/lib/python3*/site-packages/functions_framework"} {
		if matches, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(pattern))); len(matches) > 0 {
			return true
		}
	}
	return false
}
