package tauri

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
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
	cli, apps := exists(workspaceRoot, "node_modules/@tauri-apps/cli/package.json"), 0
	for _, p := range params.Projects {
		if !Detect(workspaceRoot, p.Root) {
			continue
		}
		apps++
		roots = append(roots, p.Root)
		cli = cli || exists(workspaceRoot, p.Root+"/node_modules/@tauri-apps/cli/package.json")
	}
	if apps == 0 {
		return protocol.DoctorResult{Findings: findings}
	}
	findings = append(findings, workspaceFinding(workspaceRoot, roots))
	_, rustc := lookPath("rustc")
	_, cargo := lookPath("cargo")
	findings = append(findings,
		protocol.Finding{
			Check:  "the Rust toolchain is installed",
			OK:     rustc == nil && cargo == nil,
			Detail: "rustc or cargo is not on PATH, so tauri-build cannot compile src-tauri",
			Remedy: "install Rust from https://rustup.rs",
		},
		protocol.Finding{
			Check:  "the tauri CLI is installed",
			OK:     cli,
			Detail: "node_modules has no @tauri-apps/cli",
			Remedy: "smith install -D @tauri-apps/cli in the project, then npm ci",
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
			Check:  "the system libraries Tauri needs on Linux",
			OK:     len(missing) == 0,
			Detail: "pkg-config does not find " + strings.Join(missing, ", "),
			Remedy: "sudo apt-get install -y pkg-config libwebkit2gtk-4.1-dev libgtk-3-dev librsvg2-dev",
		})
	}
	return protocol.DoctorResult{Findings: findings}
}

func exists(workspaceRoot, rel string) bool {
	_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
	return err == nil
}
