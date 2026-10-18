package pipeline

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// RustVersion is installed when rust-toolchain.toml does not pin one; bump it
// deliberately.
const RustVersion = "1.90"

var channel = regexp.MustCompile(`(?m)^[ \t]*channel[ \t]*=[ \t]*"([^"]+)"`)

// Version is the toolchain the workspace pins in rust-toolchain.toml, else
// RustVersion.
func Version(workspaceRoot string) string {
	if src, err := os.ReadFile(filepath.Join(workspaceRoot, "rust-toolchain.toml")); err == nil {
		if m := channel.FindStringSubmatch(string(src)); m != nil {
			return m[1]
		}
	}
	return RustVersion
}

// runtime installs Rust with clippy and rustfmt.
func runtime(version string) protocol.Toolchain {
	return protocol.Toolchain{
		Name: "rust", Version: version,
		GitHub: &protocol.ActionStep{Uses: "dtolnay/rust-toolchain@master", Inputs: []protocol.Input{{Name: "toolchain", Value: version}, {Name: "components", Value: "clippy, rustfmt"}}},
		// Elsewhere (Azure, GitLab): rustup, then cargo on PATH for this job
		// and, on Azure, for the steps after it.
		Install: "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --profile minimal --default-toolchain " + version + " --component clippy,rustfmt\n" +
			"export PATH=\"$HOME/.cargo/bin:$PATH\"\n" +
			"echo \"##vso[task.prependpath]$HOME/.cargo/bin\"",
		Devcontainer: &protocol.DevcontainerFeature{ID: "ghcr.io/devcontainers/features/rust:1", Options: map[string]string{"version": version, "profile": "minimal"}},
	}
}

// CI returns the Rust contribution: the toolchain, cargo fetch, and the
// crates.io token when a crate is published.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	result := protocol.CIResult{
		Toolchains: []protocol.Toolchain{runtime(Version(workspaceRoot))},
		Setup:      []protocol.Step{{Name: "Fetch crates", Command: []string{"cargo", "fetch"}, Blocking: true}},
	}
	for _, p := range params.Projects {
		if p.Publishes() {
			result.Credentials = append(result.Credentials, protocol.Credential{Env: "CARGO_REGISTRY_TOKEN", Description: "crates.io API token that can publish the workspace's crates", Required: true})
			break
		}
	}
	return result
}
