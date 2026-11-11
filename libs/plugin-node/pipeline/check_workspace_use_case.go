package pipeline

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/libs/plugin-node/registries"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Doctor checks that Node.js is the version CI installs (nodeVersion), that
// dependencies are installed, that every Node project is covered by the
// root package.json workspaces, and that a built package has every file
// its package.json names.
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

	var root struct {
		Workspaces []string `json:"workspaces"`
	}
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "package.json"))
	if err == nil {
		err = json.Unmarshal(raw, &root)
	}
	dependents := map[string][]string{}
	for _, p := range params.Projects {
		dependents[p.Name] = dependentsOf(params.Projects, p)
	}
	libraries := manifestfixes.PlanLibraryMoves(workspaceRoot, params.Projects, dependents)
	for _, p := range params.Projects {
		covered := false
		for _, pattern := range root.Workspaces {
			if ok, _ := path.Match(pattern, p.Root); ok || pattern == p.Root {
				covered = true
			}
		}
		findings = append(findings, protocol.Finding{
			Check:   "npm workspaces include " + p.Root,
			Project: p.Name,
			OK:      err == nil && covered,
			Problem: toolrun.ProblemUnless(err == nil && covered, "is not in the root package.json \"workspaces\""),
			Detail:  p.Root + " is not matched by the root package.json \"workspaces\"",
			Remedy:  "add \"" + path.Dir(p.Root) + "/*\" to \"workspaces\" in the root package.json, then run `npm install`\nor, if " + p.Root + " is deliberately standalone: add `exclude = [\"" + p.Root + "\"]` under [workspace] in smith.toml",
		})
		if f, ok := CheckScripts(workspaceRoot, p.Root, p.Name, p); ok {
			findings = append(findings, f)
		}
		if f, ok := CheckLibraryBuildDir(workspaceRoot, p, libraries, dependents[p.Name]...); ok {
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
	}
	return protocol.DoctorResult{Findings: findings}
}
