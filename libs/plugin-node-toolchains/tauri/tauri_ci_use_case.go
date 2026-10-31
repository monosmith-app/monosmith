package tauri

import (
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// RustVersion is what a Tauri app's pipelines install when the workspace pins no
// Rust toolchain; bump it deliberately.
const RustVersion = "stable"

// linuxPackages are the apt packages Tauri 2 needs on a Linux runner.
const linuxPackages = "pkg-config libwebkit2gtk-4.1-dev libgtk-3-dev libayatana-appindicator3-dev librsvg2-dev"

// rustToolchain is named like the rust plugin's, so the pipeline keeps the first of
// the two (the rust plugin's, with its components and cross-compile targets) and
// never installs a second Rust that would replace the workspace's pin as the
// default. The version is the workspace's pin, else RustVersion.
func rustToolchain(workspaceRoot string) protocol.Toolchain {
	version := FindRustPin(workspaceRoot)
	if version == "" {
		version = RustVersion
	}
	return protocol.Toolchain{
		Name: "rust", Version: version,
		GitHub: &protocol.ActionStep{Uses: "dtolnay/rust-toolchain@master", Inputs: []protocol.Input{{Name: "toolchain", Value: version}}},
		Install: "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --profile minimal --default-toolchain " + version + "\n" +
			"export PATH=\"$HOME/.cargo/bin:$PATH\"\n" +
			"echo \"##vso[task.prependpath]$HOME/.cargo/bin\"",
	}
}

// CI gives a pipeline with a Tauri app the Rust toolchain (the workspace's pinned
// one when it has one) and, on a Linux runner, the system libraries the app links
// against. `smith ci setup` runs the step and logs it, so the release's assets job
// gets it too.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	for _, p := range params.Projects {
		if tauriapp.Detect(workspaceRoot, p.Root) {
			return protocol.CIResult{
				Toolchains: []protocol.Toolchain{rustToolchain(workspaceRoot)},
				Setup: []protocol.Step{{
					Name:     "Install the system libraries Tauri needs",
					Command:  []string{"sh", "-c", "if [ \"$(uname)\" = Linux ]; then echo 'installing the system libraries Tauri needs (apt-get)'; sudo apt-get update -qq && sudo apt-get install -y -qq " + linuxPackages + "; fi"},
					Blocking: true,
				}},
			}
		}
	}
	return protocol.CIResult{}
}
