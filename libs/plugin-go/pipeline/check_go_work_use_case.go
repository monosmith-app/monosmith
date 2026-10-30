package pipeline

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Doctor checks that go.work lists every Go project and that every entry it
// lists still exists: a stale entry breaks every go command in the workspace.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	src, err := os.ReadFile(filepath.Join(workspaceRoot, "go.work"))
	if err != nil {
		return protocol.DoctorResult{Findings: []protocol.Finding{{
			Check: "go.work exists", Detail: "the workspace has Go projects but no go.work", Remedy: "run `go work init` then `go work use` for each Go project",
		}}}
	}
	work := ParseGoWork(string(src))
	listed := map[string]bool{}
	var findings []protocol.Finding
	fixed := string(src)
	for _, entry := range work.Use {
		listed[entry] = true
		_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(entry), "go.mod"))
		if err != nil {
			fixed, _ = DropUse(fixed, entry)
		}
		findings = append(findings, protocol.Finding{
			Check: "go.work entry " + entry + " exists", OK: err == nil,
			Detail: entry + " is listed in go.work but has no go.mod", Remedy: "remove `" + entry + "` from go.work (`go work edit -dropuse=./" + entry + "`)",
		})
	}
	lints := false
	for _, p := range params.Projects {
		if !listed[p.Root] {
			fixed = AddUse(fixed, p.Root)
		}
		findings = append(findings, protocol.Finding{
			Check: "go.work lists " + p.Root, OK: listed[p.Root],
			Detail: p.Root + " has a go.mod but go.work does not use it", Remedy: "run `go work use ./" + p.Root + "`",
		})
		if lint, ok := p.Tasks["lint"]; ok && len(lint.Command) > 0 && lint.Command[0] == "golangci-lint" {
			lints = true
		}
	}
	// One fix makes go.work match the modules, for every finding it clears.
	if fixed != string(src) {
		fix := &protocol.Fix{Description: "make go.work use exactly the workspace's Go modules", Safe: true, Files: []protocol.FileEdit{{Path: "go.work", Content: fixed}}}
		for i := range findings {
			if !findings[i].OK {
				findings[i].Fix = fix
			}
		}
	}
	for _, p := range params.Projects {
		if p.Kind == protocol.ProjectPackage {
			findings = append(findings, CheckPackageImports(workspaceRoot, p))
		}
	}
	required := work.GoVersion
	if required == "" {
		required = DefaultGoVersion
	}
	findings = append(findings, CheckToolVersions(workspaceRoot, required, lints)...)
	return protocol.DoctorResult{Findings: findings}
}
