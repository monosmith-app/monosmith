package pipeline

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/registries"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
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
	findings = append(findings, protocol.Finding{
		Check:  "Node dependencies are installed",
		OK:     err == nil,
		Detail: "node_modules is missing at the workspace root",
		Remedy: "run `npm install` at the workspace root",
	})
	if f, ok := registries.CheckNpmrc(workspaceRoot, params.Projects); ok {
		findings = append(findings, f)
	}
	if f, ok := CheckScripts(workspaceRoot, "", "the root package.json", protocol.Project{}); ok {
		findings = append(findings, f)
	}
	if f, ok := CheckConfigImports(workspaceRoot); ok {
		findings = append(findings, f)
	}

	var root struct {
		Workspaces []string `json:"workspaces"`
	}
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "package.json"))
	if err == nil {
		err = json.Unmarshal(raw, &root)
	}
	for _, p := range params.Projects {
		covered := false
		for _, pattern := range root.Workspaces {
			if ok, _ := path.Match(pattern, p.Root); ok || pattern == p.Root {
				covered = true
			}
		}
		findings = append(findings, protocol.Finding{
			Check:  "npm workspaces include " + p.Root,
			OK:     err == nil && covered,
			Detail: p.Root + " is not matched by the root package.json \"workspaces\"",
			Remedy: "add \"" + path.Dir(p.Root) + "/*\" to \"workspaces\" in the root package.json, then run `npm install`\nor, if " + p.Root + " is deliberately standalone: add `exclude = [\"" + p.Root + "\"]` under [workspace] in smith.toml",
		})
		if f, ok := CheckScripts(workspaceRoot, p.Root, p.Name, p); ok {
			findings = append(findings, f)
		}
		if f, ok := CheckBuildDir(workspaceRoot, p); ok {
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
