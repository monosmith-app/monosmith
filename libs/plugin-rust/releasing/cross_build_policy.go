package releasing

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Reach says what building a platform's program takes on this machine.
type Reach int

const (
	// Here: the machine builds it with the Rust target alone.
	Here Reach = iota
	// WithLinker: the machine builds it once a linker for the other
	// processor is installed (Linux for another Linux processor).
	WithLinker
	// OnItsOwnSystem: no linker or SDK for the target can be had here; the
	// program is built on the system it is for.
	OnItsOwnSystem
)

// CrossReach says whether a program for platform can be built on host (both
// "os/arch"). Windows builds for either Windows processor and macOS for
// either Mac processor with the Rust target alone; Linux needs a linker for
// the other processor; nothing builds for another operating system, whose
// system libraries and linker are not on the machine.
func CrossReach(host, platform string) Reach {
	hostOS, hostArch, _ := strings.Cut(host, "/")
	targetOS, arch, _ := strings.Cut(platform, "/")
	switch {
	case host == platform:
		return Here
	case hostOS != targetOS:
		return OnItsOwnSystem
	case targetOS == "linux" && hostArch != arch:
		return WithLinker
	}
	return Here
}

// linkers are the C compilers that link a Linux program for the other
// processor, by the platform's processor.
var linkers = map[string]string{"arm64": "aarch64-linux-gnu-gcc", "amd64": "x86_64-linux-gnu-gcc"}

// Linker is the program that links platform's Linux programs on another
// processor's machine, and the Debian package that installs it.
func Linker(platform string) (program, pkg string) {
	_, arch, _ := strings.Cut(platform, "/")
	program = linkers[arch]
	return program, "gcc-" + strings.ReplaceAll(strings.TrimSuffix(program, "-gcc"), "_", "-")
}

// LinkerEnv is the variable cargo reads the linker of a target from.
func LinkerEnv(triple string) string {
	return "CARGO_TARGET_" + strings.ToUpper(strings.ReplaceAll(triple, "-", "_")) + "_LINKER"
}

// LinkerAvailable reports whether the platform's linker is set up: the
// environment names one, the workspace's .cargo/config.toml configures one
// for the target, or the usual cross compiler is on PATH. Tests replace it.
var LinkerAvailable = func(workspaceRoot, platform, triple string) bool {
	if os.Getenv(LinkerEnv(triple)) != "" {
		return true
	}
	if configured(workspaceRoot, triple) {
		return true
	}
	program, _ := Linker(platform)
	return program != "" && toolrun.Available(program)
}

// configured reports whether .cargo/config.toml has a [target.<triple>]
// table with a linker.
func configured(workspaceRoot, triple string) bool {
	src, err := os.ReadFile(filepath.Join(workspaceRoot, ".cargo", "config.toml"))
	if err != nil {
		return false
	}
	table := regexp.MustCompile(`(?s)\[target\.` + regexp.QuoteMeta(triple) + `\](.*?)(\n\[|$)`)
	m := table.FindStringSubmatch(string(src))
	return m != nil && strings.Contains(m[1], "linker")
}
