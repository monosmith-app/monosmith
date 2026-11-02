package functions

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// lookPath is replaced by tests.
var lookPath = exec.LookPath

// hostOf says which function host a project is written for, by the files that
// host needs; empty for any other project.
func hostOf(workspaceRoot, root string) string {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
	manifest, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	template, _ := os.ReadFile(filepath.Join(dir, "template.yaml"))
	switch {
	case strings.Contains(string(template), "AWS::Serverless-2016-10-31"):
		return functionhost.AWS
	case strings.Contains(string(manifest), `"@azure/functions"`):
		return functionhost.Azure
	case strings.Contains(string(manifest), `"@google-cloud/functions-framework"`):
		return functionhost.GCP
	}
	return ""
}

// Doctor checks, for each host the workspace's function apps use, that the tool
// that runs them locally is installed. A missing tool is a warning: the build and
// the tests need no cloud tool.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	seen := map[string]bool{}
	findings := []protocol.Finding{}
	for _, p := range params.Projects {
		host := hostOf(workspaceRoot, p.Root)
		tool, ok := functionhost.ToolOf(host)
		if host == "" || seen[host] || !ok {
			continue
		}
		seen[host] = true
		installed := false
		if tool.Binary != "" {
			_, err := lookPath(tool.Binary)
			installed = err == nil
		} else {
			_, err := os.Stat(filepath.Join(workspaceRoot, "node_modules", ".bin", "functions-framework"))
			installed = err == nil
		}
		findings = append(findings, functionhost.FindingFor(tool, installed))
	}
	return protocol.DoctorResult{Findings: findings}
}

// Doctor implements serve.DoctorContributor.
func (p *Functions) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return Doctor(p.root, params), nil
}
