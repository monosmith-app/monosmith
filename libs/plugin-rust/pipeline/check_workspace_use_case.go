package pipeline

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-rust/releasing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// probe runs a tool for its version; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// Doctor checks that cargo is installed and that the root Cargo.toml's
// workspace lists every crate and nothing that is gone: cargo refuses a
// workspace whose members are missing, and ignores a crate it does not list.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	var findings []protocol.Finding
	_, err := probe(workspaceRoot, []string{"cargo", "--version"})
	findings = append(findings, protocol.Finding{
		Check: "cargo is installed", OK: err == nil,
		Detail: "cargo is not on PATH", Remedy: "install Rust " + Version(workspaceRoot) + " with rustup (https://rustup.rs)",
	})
	if f, ok := CheckRustVersion(workspaceRoot); ok {
		findings = append(findings, f)
	}
	src, err := os.ReadFile(filepath.Join(workspaceRoot, "Cargo.toml"))
	if err != nil {
		return protocol.DoctorResult{Findings: append(findings, protocol.Finding{
			Check: "the root Cargo.toml is a workspace", Detail: "the workspace has crates but no root Cargo.toml", Remedy: "create Cargo.toml with [workspace] and members = [...] listing each crate's directory",
		})}
	}
	root, err := discovery.ParseCargo(src)
	if err != nil {
		return protocol.DoctorResult{Findings: append(findings, protocol.Finding{Check: "the root Cargo.toml is valid TOML", Detail: err.Error(), Remedy: "fix the root Cargo.toml"})}
	}
	for _, m := range root.Members {
		if strings.ContainsAny(m, "*?[") {
			continue
		}
		_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(m), "Cargo.toml"))
		findings = append(findings, protocol.Finding{
			Check: "workspace member " + m + " exists", OK: err == nil,
			Detail: m + " is a member of the Cargo workspace but has no Cargo.toml", Remedy: "remove \"" + m + "\" from members in the root Cargo.toml",
		})
	}
	for _, p := range params.Projects {
		findings = append(findings, releasing.CheckReleaseTargets(workspaceRoot, p)...)
		listed := slices.Contains(root.Members, p.Root) || slices.ContainsFunc(root.Members, func(m string) bool {
			ok, _ := filepath.Match(m, p.Root)
			return ok
		})
		findings = append(findings, protocol.Finding{
			Check: "the Cargo workspace lists " + p.Root, OK: listed,
			Detail: p.Root + " has a Cargo.toml but the workspace does not list it, so cargo -p cannot find it", Remedy: "add \"" + p.Root + "\" to members in the root Cargo.toml",
		})
	}
	return protocol.DoctorResult{Findings: findings}
}
