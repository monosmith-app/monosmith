package pipeline

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-rust/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-rust/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// probe runs a tool for its version; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Probe(dir, argv) }

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
		Check: "Rust (cargo) is installed", Area: "toolchain", OK: err == nil, Problem: toolrun.ProblemUnless(err == nil, "cargo is not on PATH"),
		Detail: "cargo is not on PATH", Remedy: "install Rust " + Version(workspaceRoot) + " with rustup (https://rustup.rs)",
	})
	if err == nil {
		_, auditErr := probe(workspaceRoot, []string{"cargo", "audit", "--version"})
		findings = append(findings, protocol.Finding{
			Check: "cargo-audit is installed", Area: "toolchain", OK: auditErr == nil, Warning: true, Problem: toolrun.ProblemUnless(auditErr == nil, "cargo-audit is not installed"),
			Detail: "skipped: cargo-audit is not installed, so smith ci audit cannot check the crates for known vulnerabilities", Remedy: "`cargo install --locked cargo-audit`",
		})
	}
	if f, ok := CheckRustVersion(workspaceRoot); ok {
		findings = append(findings, f)
	}
	if found, ok := CheckBuildDir(workspaceRoot, params.Projects); ok {
		findings = append(findings, found...)
	}
	src, err := os.ReadFile(filepath.Join(workspaceRoot, "Cargo.toml"))
	if err != nil {
		return protocol.DoctorResult{Findings: append(findings, protocol.Finding{
			Check: "the root Cargo.toml is a workspace", Area: "config", Problem: "the workspace has crates but no root Cargo.toml", Detail: "the workspace has crates but no root Cargo.toml", Remedy: "create Cargo.toml with [workspace] and members = [...] listing each crate's directory",
		})}
	}
	root, err := discovery.ParseCargo(src)
	if err != nil {
		return protocol.DoctorResult{Findings: append(findings, protocol.Finding{Check: "the root Cargo.toml is valid TOML", Area: "config", Problem: "the root Cargo.toml is not valid TOML", Detail: err.Error(), Remedy: "fix the root Cargo.toml"})}
	}
	for _, m := range root.Members {
		if strings.ContainsAny(m, "*?[") {
			continue
		}
		_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(m), "Cargo.toml"))
		findings = append(findings, protocol.Finding{
			Check: "workspace member " + m + " exists", Area: "config", OK: err == nil, Problem: toolrun.ProblemUnless(err == nil, m+" is a member of the Cargo workspace but has no Cargo.toml"),
			Detail: m + " is a member of the Cargo workspace but has no Cargo.toml", Remedy: "remove \"" + m + "\" from members in the root Cargo.toml",
		})
	}
	for _, p := range params.Projects {
		findings = append(findings, releasing.CheckReleaseTargets(workspaceRoot, p)...)
		listing := workspaceListing(workspaceRoot, string(src), root, p.Root)
		listing.Project = p.Name
		findings = append(findings, listing)
	}
	return protocol.DoctorResult{Findings: findings}
}

// workspaceListing checks that the root workspace knows a crate. A crate that
// has its own [workspace] table is the root of another workspace (a throw-away
// tool, say), and a directory in exclude is left out on purpose: neither is
// "missing from members". A crate that is neither listed nor excluded fails
// with a choice, since only the person knows if it belongs to the workspace.
func workspaceListing(workspaceRoot, rootSrc string, root discovery.Crate, dir string) protocol.Finding {
	matches := func(list []string) bool {
		return slices.ContainsFunc(list, func(m string) bool {
			ok, _ := filepath.Match(m, dir)
			return m == dir || ok
		})
	}
	if matches(root.Members) {
		return protocol.Finding{Check: "the Cargo workspace lists " + dir, OK: true}
	}
	if matches(root.Exclude) {
		return protocol.Finding{Check: dir + " is excluded from the Cargo workspace", OK: true}
	}
	if own, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(dir), "Cargo.toml")); err == nil {
		if crate, err := discovery.ParseCargo(own); err == nil && crate.HasWorkspace {
			return protocol.Finding{Check: dir + " is its own Cargo workspace", OK: true}
		}
	}
	f := protocol.Finding{
		Check:   "the Cargo workspace lists " + dir,
		Problem: "is neither listed nor excluded by the Cargo workspace",
		Detail:  dir + " has a Cargo.toml but the workspace neither lists nor excludes it, so cargo -p cannot find it",
		Remedy:  "add \"" + dir + "\" to members in the root Cargo.toml, or to exclude if it is not part of this workspace",
	}
	add, errAdd := scaffolding.AddMember(rootSrc, dir)
	skip, errSkip := scaffolding.AddExclude(rootSrc, dir)
	if errAdd != nil || errSkip != nil {
		return f // no [workspace] table to edit: the remedy says what to write
	}
	f.Fix = &protocol.Fix{
		Description: "tell the Cargo workspace about " + dir,
		Choices: []protocol.FixChoice{{
			Question: dir + " has a Cargo.toml the workspace does not mention. Is it part of this workspace?",
			Options: []protocol.FixChoiceOption{
				{Label: "add it to members: cargo builds it with the workspace and shares its Cargo.lock", Fix: protocol.Fix{Description: "add " + dir + " to members", Files: []protocol.FileEdit{{Path: "Cargo.toml", Content: add}}}},
				{Label: "not part of this workspace: add it to [workspace] exclude, and cargo leaves it alone", Fix: protocol.Fix{Description: "add " + dir + " to exclude", Files: []protocol.FileEdit{{Path: "Cargo.toml", Content: skip}}}},
			},
		}},
	}
	return f
}
