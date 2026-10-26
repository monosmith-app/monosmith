package pipeline

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// ToolchainFile is where a workspace pins Rust for rustup, local and CI alike.
const ToolchainFile = "rust-toolchain.toml"

var rustcVersion = regexp.MustCompile(`rustc (\d+\.\d+(?:\.\d+)?)`)

// CheckRustVersion compares the rustc rustup resolves in the workspace (so a
// rust-toolchain.toml counts) with the Rust the generated CI and dev
// container install. A newer local Rust warns, an older one fails: a crate
// can need a newer Cargo than the machine has, and the error then reads like
// a dependency problem ("feature edition2024 is required"). A machine
// without rustc gives no finding; "cargo is installed" already says so.
//
// The fix pins Rust in rust-toolchain.toml, which the generated CI, the dev
// container and rustup all follow: an older local installs the pinned
// toolchain, a newer one moves the pin to what this machine runs. It never
// rewrites a rust-toolchain.toml that exists.
func CheckRustVersion(workspaceRoot string) (protocol.Finding, bool) {
	out, err := probe(workspaceRoot, []string{"rustc", "--version"})
	m := rustcVersion.FindStringSubmatch(out)
	if err != nil || m == nil {
		return protocol.Finding{}, false
	}
	ci := Version(workspaceRoot)
	f, ok := toolrun.CIPinFinding("Rust", m[1], ci)
	if !ok {
		return protocol.Finding{}, false
	}
	if f.OK {
		return f, true
	}
	_, statErr := os.Stat(filepath.Join(workspaceRoot, ToolchainFile))
	pinned := statErr == nil
	if f.Warning {
		local := majorMinor(m[1])
		f.Remedy = "pin Rust once in " + ToolchainFile + " (channel = \"" + local + "\") and run `smith ci generate`, so CI and the dev container follow this machine"
		if !pinned {
			f.Fix = &protocol.Fix{
				Description: "pin Rust " + local + " in " + ToolchainFile,
				Files:       []protocol.FileEdit{{Path: ToolchainFile, Content: toolchainFile(local)}},
			}
		}
		return f, true
	}
	f.Remedy = "`rustup toolchain install " + ci + "` (or `rustup update`), and pin it in " + ToolchainFile + " so local, CI and the dev container agree"
	fix := &protocol.Fix{
		Description: "install Rust " + ci,
		Steps: []protocol.FixStep{{
			Description: "install the toolchain CI uses",
			Argv:        []string{"rustup", "toolchain", "install", ci, "--profile", "minimal", "--component", "clippy,rustfmt"},
		}},
	}
	if !pinned {
		fix.Description = "pin Rust " + ci + " in " + ToolchainFile + " and install it"
		fix.Files = []protocol.FileEdit{{Path: ToolchainFile, Content: toolchainFile(ci)}}
	}
	f.Fix = fix
	return f, true
}

// toolchainFile is a rust-toolchain.toml pinning the channel.
func toolchainFile(channel string) string {
	return "[toolchain]\nchannel = \"" + channel + "\"\ncomponents = [\"clippy\", \"rustfmt\"]\n"
}

// majorMinor trims 1.90.2 to 1.90.
func majorMinor(v string) string {
	if m := regexp.MustCompile(`^\d+\.\d+`).FindString(v); m != "" {
		return m
	}
	return v
}
