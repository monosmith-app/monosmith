package tauri

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/cargoworkspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/extras"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/versionsync"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// lookPath, platform and pkgConfigHas are replaced by tests.
var (
	lookPath = exec.LookPath
	platform = runtime.GOOS
	// pkgConfigHas reports whether pkg-config knows the library.
	pkgConfigHas = func(lib string) bool { return exec.Command("pkg-config", "--exists", lib).Run() == nil }
)

// linuxLibs are the system libraries Tauri 2 links against on Linux.
var linuxLibs = []string{"webkit2gtk-4.1", "gtk+-3.0", "librsvg-2.0"}

// Doctor checks what the tauri tasks need from this machine: the Rust
// toolchain, the tauri CLI and, on Linux, the system libraries.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	findings := []protocol.Finding{}
	var roots []string
	cli, apps, firstApp := exists(workspaceRoot, "node_modules/@tauri-apps/cli/package.json"), 0, ""
	for _, p := range params.Projects {
		if !tauriapp.Detect(workspaceRoot, p.Root) {
			continue
		}
		apps++
		if firstApp == "" {
			firstApp = p.Name
		}
		roots = append(roots, p.Root)
		cli = cli || exists(workspaceRoot, p.Root+"/node_modules/@tauri-apps/cli/package.json")
	}
	if apps == 0 {
		return protocol.DoctorResult{Findings: findings}
	}
	findings = append(findings, cargoworkspace.WorkspaceFinding(workspaceRoot, roots))
	findings = append(findings, frontendDistFindings(workspaceRoot, params.Projects)...)
	findings = append(findings, extras.Findings(workspaceRoot, params.Projects)...)
	findings = append(findings, versionsync.Findings(workspaceRoot, params.Projects)...)
	_, rustc := lookPath("rustc")
	_, cargo := lookPath("cargo")
	var cliFix *protocol.Fix
	if !cli && firstApp != "" {
		cliFix = &protocol.Fix{Description: "Add the tauri CLI to " + firstApp, Steps: []protocol.FixStep{{
			Description: "smith install adds @tauri-apps/cli to the development dependencies of " + firstApp,
			Argv:        []string{"smith", "install", "-w", firstApp, "-D", "@tauri-apps/cli", "--yes"},
		}}}
	}
	findings = append(findings,
		protocol.Finding{
			Check:   "Rust (cargo) is installed",
			Area:    "toolchain",
			OK:      rustc == nil && cargo == nil,
			Problem: toolrun.ProblemUnless(rustc == nil && cargo == nil, "rustc or cargo is not on PATH"),
			Detail:  "rustc or cargo is not on PATH, so tauri-build cannot compile src-tauri",
			Remedy:  "install Rust from https://rustup.rs",
		},
		protocol.Finding{
			Check:   "the tauri CLI is installed",
			Area:    "toolchain",
			OK:      cli,
			Problem: toolrun.ProblemUnless(cli, "node_modules has no @tauri-apps/cli"),
			Detail:  "node_modules has no @tauri-apps/cli",
			Remedy:  "smith install -D @tauri-apps/cli in the project, then npm ci",
			Fix:     cliFix,
		})
	if platform == "linux" {
		missing := linuxLibs
		if _, err := lookPath("pkg-config"); err == nil {
			missing = nil
			for _, lib := range linuxLibs {
				if !pkgConfigHas(lib) {
					missing = append(missing, lib)
				}
			}
		}
		findings = append(findings, protocol.Finding{
			Check:   "the system libraries Tauri needs on Linux",
			Area:    "toolchain",
			OK:      len(missing) == 0,
			Problem: toolrun.ProblemUnless(len(missing) == 0, "pkg-config does not find "+strings.Join(missing, ", ")),
			Detail:  "pkg-config does not find " + strings.Join(missing, ", "),
			Remedy:  "sudo apt-get install -y pkg-config libwebkit2gtk-4.1-dev libgtk-3-dev librsvg2-dev",
		})
	}
	return protocol.DoctorResult{Findings: findings}
}

func exists(workspaceRoot, rel string) bool {
	_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
	return err == nil
}
