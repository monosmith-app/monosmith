package tauri

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// RustVersion is what a Tauri app's pipelines install; bump it deliberately.
const RustVersion = "stable"

// linuxPackages are the apt packages Tauri 2 needs on a Linux runner.
const linuxPackages = "pkg-config libwebkit2gtk-4.1-dev libgtk-3-dev libayatana-appindicator3-dev librsvg2-dev"

// rustToolchain is named apart from the rust plugin's "rust", so a workspace with
// both keeps that plugin's version, components and cross-compile targets.
var rustToolchain = protocol.Toolchain{
	Name: "tauri-rust", Version: RustVersion,
	GitHub: &protocol.ActionStep{Uses: "dtolnay/rust-toolchain@master", Inputs: []protocol.Input{{Name: "toolchain", Value: RustVersion}}},
	Install: "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --profile minimal --default-toolchain " + RustVersion + "\n" +
		"export PATH=\"$HOME/.cargo/bin:$PATH\"\n" +
		"echo \"##vso[task.prependpath]$HOME/.cargo/bin\"",
}

// CI gives a pipeline with a Tauri app the Rust toolchain and, on a Linux
// runner, the system libraries the app links against. `smith ci setup` runs
// the step, so the release's assets job gets it too.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	for _, p := range params.Projects {
		if Detect(workspaceRoot, p.Root) {
			return protocol.CIResult{
				Toolchains: []protocol.Toolchain{rustToolchain},
				Setup: []protocol.Step{{
					Name:     "Install the system libraries Tauri needs",
					Command:  []string{"sh", "-c", "if [ \"$(uname)\" = Linux ]; then sudo apt-get update -qq && sudo apt-get install -y -qq " + linuxPackages + "; fi"},
					Blocking: true,
				}},
			}
		}
	}
	return protocol.CIResult{}
}
