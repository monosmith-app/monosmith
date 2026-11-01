package pipeline

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// GovulncheckVersion is pinned; bump it deliberately.
const GovulncheckVersion = "v1.9.0"

// GolangciLintVersion is installed when a project lints with golangci-lint;
// bump it deliberately.
const GolangciLintVersion = "v2.14.0"

// DefaultGoVersion is used when go.work does not say.
const DefaultGoVersion = "1.24"

// toolchainVersion is the Go version the pipelines install: go.work's, and
// without one (or without a go directive in it) the newest a project's go.mod
// asks for, so a missing go.work does not downgrade the pipelines to the default.
func toolchainVersion(workspaceRoot string, projects []protocol.Project) string {
	if src, err := os.ReadFile(filepath.Join(workspaceRoot, "go.work")); err == nil {
		if v := ParseGoWork(string(src)).GoVersion; v != "" {
			return v
		}
	}
	newest := ""
	for _, p := range projects {
		src, err := os.ReadFile(ModuleFile(workspaceRoot, p.Root))
		if err != nil {
			continue
		}
		if m := goDirective.FindStringSubmatch(string(src)); m != nil && versionLess(newest, m[1]) {
			newest = m[1]
		}
	}
	if newest == "" {
		return DefaultGoVersion
	}
	return newest
}

// versionLess compares two major.minor versions numerically ("1.9" < "1.24").
func versionLess(a, b string) bool {
	if a == "" {
		return true
	}
	var am, an, bm, bn int
	_, _ = fmt.Sscanf(a, "%d.%d", &am, &an)
	_, _ = fmt.Sscanf(b, "%d.%d", &bm, &bn)
	return am < bm || (am == bm && an < bn)
}

// CI returns the Go contribution: the toolchain at go.work's version and a
// non-blocking vulnerability scan of every module.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	version := toolchainVersion(workspaceRoot, params.Projects)
	result := protocol.CIResult{
		Toolchains: []protocol.Toolchain{runtime(version)},
		Setup:      []protocol.Step{{Name: "Download Go modules", Command: []string{"go", "mod", "download"}, Blocking: true}},
	}
	for _, p := range params.Projects {
		if lint, ok := p.Tasks["lint"]; ok && len(lint.Command) > 0 && lint.Command[0] == "golangci-lint" {
			result.Setup = append(result.Setup, protocol.Step{Name: "Install golangci-lint", Command: []string{"go", "install", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@" + GolangciLintVersion}, Blocking: true})
			break
		}
	}
	store := protocol.PackageStore{Name: "go", Path: "~/go/pkg/mod", Env: "GOMODCACHE"}
	for _, p := range params.Projects {
		store.Files = append(store.Files, path.Join(p.Root, "go.sum"))
	}
	result.Stores = []protocol.PackageStore{store}
	// govulncheck needs a module, so it runs once in each one.
	for _, p := range params.Projects {
		result.Audit = append(result.Audit, protocol.Step{Name: "govulncheck " + p.Name, Command: []string{"go", "run", "golang.org/x/vuln/cmd/govulncheck@" + GovulncheckVersion, "./..."}, Dir: p.Root, Blocking: false, CacheFiles: []string{path.Join(p.Root, "go.mod"), path.Join(p.Root, "go.sum")}})
	}
	if params.Audit.Enabled("go") {
		// Opt-in ([ci] audit_policies): the module cache still matches go.sum.
		for _, p := range params.Projects {
			result.Audit = append(result.Audit, protocol.Step{Name: "go mod verify " + p.Name, Command: []string{"go", "mod", "verify"}, Dir: p.Root, Blocking: params.Audit.Blocks("go"), CacheFiles: []string{path.Join(p.Root, "go.mod"), path.Join(p.Root, "go.sum")}})
		}
	}
	return result
}
