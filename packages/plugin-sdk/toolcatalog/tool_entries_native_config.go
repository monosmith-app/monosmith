package toolcatalog

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// The native entries are the system tools of C and C++ and of desktop apps
// (Tauri). They are limited to the systems that have them.
func init() {
	entries = append(entries,
		protocol.ToolNeed{
			Name: "xcode-clt", Title: "the Xcode Command Line Tools", Systems: []string{"darwin"},
			Why:         "the C, C++ and Rust compilers and the linker of macOS come with the Xcode Command Line Tools",
			Consequence: "smith cannot build C, C++ or Tauri projects on this Mac",
			Program:     "xcode-select", VersionArgs: []string{"-p"},
			Download: "https://v2.tauri.app/start/prerequisites/#macos", Source: "https://v2.tauri.app/start/prerequisites/", Checked: checked,
		},
		protocol.ToolNeed{
			Name: "msvc-build-tools", Title: "the Microsoft C++ Build Tools", Systems: []string{"windows"},
			Why:         "C, C++ and Rust programs on Windows are compiled and linked with the Microsoft C++ Build Tools (the C++ workload of Visual Studio)",
			Consequence: "smith cannot build C, C++ or Tauri projects on this computer",
			// vswhere is not on PATH; Microsoft installs it in this fixed folder. It
			// shows the Visual Studio Installer is there, not that the C++ workload is.
			Program:     `C:\Program Files (x86)\Microsoft Visual Studio\Installer\vswhere.exe`,
			VersionArgs: []string{"-latest", "-products", "*", "-requires", "Microsoft.VisualStudio.Component.VC.Tools.x86.x64", "-property", "installationVersion"},
			Install: []protocol.ToolInstall{
				{Manager: "winget", Package: "Microsoft.VisualStudio.2022.BuildTools", Note: "the installer opens with no workload chosen: tick \"Desktop development with C++\" there"},
			},
			Download: "https://visualstudio.microsoft.com/visual-cpp-build-tools/", Source: "https://v2.tauri.app/start/prerequisites/", Checked: checked,
		},
		protocol.ToolNeed{
			Name: "webview2", Title: "the Microsoft Edge WebView2 Runtime", Systems: []string{"windows"},
			Why:         "a Tauri app shows its pages in the WebView2 Runtime, which Windows 10 (version 1803 and later) and Windows 11 already have",
			Consequence: "smith cannot run a Tauri app on this computer",
			// Microsoft documents this registry value as the way to detect the runtime (machine-wide install).
			Program:     "reg",
			VersionArgs: []string{"query", `HKLM\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`, "/v", "pv"},
			Install:     []protocol.ToolInstall{install("winget", "Microsoft.EdgeWebView2Runtime")},
			Download:    "https://developer.microsoft.com/microsoft-edge/webview2/", Source: "https://learn.microsoft.com/microsoft-edge/webview2/concepts/distribution", Checked: checked,
		},
		protocol.ToolNeed{
			Name: "tauri-linux-libs", Title: "the Tauri system libraries (WebKitGTK and build tools)", Systems: []string{"linux"},
			Why:         "a Tauri app on Linux is built against WebKitGTK and a few system libraries",
			Consequence: "smith cannot build or run a Tauri app on this computer",
			Program:     "pkg-config", VersionArgs: []string{"--modversion", "webkit2gtk-4.1"},
			Install: []protocol.ToolInstall{
				install("apt", "libwebkit2gtk-4.1-dev build-essential pkg-config libxdo-dev libssl-dev libayatana-appindicator3-dev librsvg2-dev"),
				install("dnf", "webkit2gtk4.1-devel gcc gcc-c++ make pkgconf-pkg-config openssl-devel libappindicator-gtk3-devel librsvg2-devel libxdo-devel"),
				{Manager: "pacman", Package: "webkit2gtk-4.1 base-devel pkgconf openssl xdotool librsvg", Note: "tray icons also need libappindicator-gtk3, which Arch no longer has in its own repositories"},
			},
			Download: "https://v2.tauri.app/start/prerequisites/#linux", Source: "https://v2.tauri.app/start/prerequisites/", Checked: checked,
		},
		protocol.ToolNeed{
			Name: "cpp-compiler", Title: "a C and C++ compiler", Systems: []string{"linux"},
			Why:         "the C and C++ projects compile with gcc or clang",
			Consequence: "smith build and test fail for the C and C++ projects",
			Program:     "g++", Alternatives: []string{"clang++", "c++"},
			Install: []protocol.ToolInstall{
				install("apt", "build-essential"), install("dnf", "gcc gcc-c++ make"), install("pacman", "base-devel"),
			},
			Download: "https://gcc.gnu.org/install/", Source: "https://gcc.gnu.org/install/", Checked: checked,
		},
	)
}

// ForChoice returns the entries with the given names, each limited to the
// stacks that have the choice. A plugin whose choice needs a tool that other
// projects of its language do not (Tauri needs Rust) lists it this way.
//
// Parameters: choice is the choice name; names are entry names.
// Returns: copies of the entries with Choice set.
// Panics when a name is not in the catalogue.
func ForChoice(choice string, names ...string) []protocol.ToolNeed {
	out := For(names...)
	for i := range out {
		out[i].Choice = choice
	}
	return out
}
