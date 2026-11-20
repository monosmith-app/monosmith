package pipeline

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/registries"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Doctor checks that Node.js is the version CI installs (nodeVersion), that
// dependencies are installed, and that a built package has every file its
// package.json names. Whether the root package.json workspaces cover each
// project is the join check the host asks for (workspacejoin).
func Doctor(workspaceRoot string, params protocol.DoctorParams, nodeVersion string) protocol.DoctorResult {
	var findings []protocol.Finding
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	findings = append(findings, CheckNodeVersion(workspaceRoot, nodeVersion))
	_, err := os.Stat(filepath.Join(workspaceRoot, "node_modules"))
	installed := protocol.Finding{
		Check:   "Node dependencies are installed",
		Area:    "toolchain",
		OK:      err == nil,
		Problem: toolrun.ProblemUnless(err == nil, "node_modules is missing at the workspace root"),
		Detail:  "node_modules is missing at the workspace root",
		Remedy:  "run `npm install` at the workspace root",
	}
	if err != nil {
		installed.Fix = &protocol.Fix{
			Description: "install the Node dependencies",
			Steps:       []protocol.FixStep{{Description: "install the workspace's dependencies", Argv: []string{"npm", "install", "--no-audit", "--no-fund"}}},
		}
	}
	findings = append(findings, installed)
	if f, ok := registries.CheckNpmrc(workspaceRoot, params.Projects); ok {
		findings = append(findings, f)
	}
	if f, ok := CheckScripts(workspaceRoot, "", "the root package.json", protocol.Project{}); ok {
		findings = append(findings, f)
	}
	if f, ok := CheckConfigImports(workspaceRoot); ok {
		findings = append(findings, f)
	}

	findings = append(findings, CheckInternalRangeFindings(workspaceRoot, params.Projects)...)

	dependents := map[string][]string{}
	for _, p := range params.Projects {
		dependents[p.Name] = dependentsOf(params.Projects, p)
	}
	for _, p := range params.Projects {
		if f, ok := CheckScripts(workspaceRoot, p.Root, p.Name, p); ok {
			findings = append(findings, f)
		}
		if f, ok := CheckBuildDir(workspaceRoot, p, dependents[p.Name]...); ok {
			findings = append(findings, f)
		}
		if f, ok := CheckElectronDir(workspaceRoot, p); ok {
			findings = append(findings, f)
		}
		if p.Kind == protocol.ProjectPackage {
			findings = append(findings, CheckPackageImports(workspaceRoot, p))
		}
		if f, ok := CheckEntryFiles(workspaceRoot, p); ok {
			findings = append(findings, f)
		}
		if f, ok := CheckTsconfigChain(workspaceRoot, p); ok {
			findings = append(findings, f)
		}
	}
	return protocol.DoctorResult{Findings: findings}
}
