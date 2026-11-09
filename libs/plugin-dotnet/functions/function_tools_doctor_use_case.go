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
// build and the tests need no cloud tool. Google's framework is a NuGet
// package that dotnet restores, so it has no finding.
func Findings(workspaceRoot string, projects []protocol.Project) []protocol.Finding {
	seen := map[string]bool{}
	var findings []protocol.Finding
	for _, p := range projects {
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
		host := DetectHost(dir, ProjectFile(dir))
		tool, ok := functionhost.ToolOf(host)
		if host == "" || tool.Binary == "" || seen[host] || !ok {
			continue
		}
		seen[host] = true
		_, err := lookPath(tool.Binary)
		findings = append(findings, functionhost.FindingFor(tool, err == nil))
	}
	return findings
}
