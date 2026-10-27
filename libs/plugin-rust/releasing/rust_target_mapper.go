package releasing

import (
	"fmt"
	"slices"
	"strings"
)

// targets are the Rust target triples of the platforms an app is released
// for, by Go's "os/arch" spelling (the one smith.toml uses). Windows is the
// MSVC toolchain's, Linux glibc's: what each ecosystem's users run.
var targets = map[string]string{
	"windows/amd64": "x86_64-pc-windows-msvc",
	"windows/arm64": "aarch64-pc-windows-msvc",
	"linux/amd64":   "x86_64-unknown-linux-gnu",
	"linux/arm64":   "aarch64-unknown-linux-gnu",
	"darwin/amd64":  "x86_64-apple-darwin",
	"darwin/arm64":  "aarch64-apple-darwin",
}

// Target is the Rust target triple of a platform ("os/arch").
func Target(platform string) (string, error) {
	if t, ok := targets[platform]; ok {
		return t, nil
	}
	known := make([]string, 0, len(targets))
	for p := range targets {
		known = append(known, p)
	}
	slices.Sort(known)
	return "", fmt.Errorf("%q has no Rust target: release a Rust app for %s", platform, strings.Join(known, ", "))
}
