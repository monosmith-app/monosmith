package tauri

import (
	"os"
	"path/filepath"
	"regexp"
)

var (
	toolchainChannel = regexp.MustCompile(`(?m)^[ \t]*channel[ \t]*=[ \t]*"([^"]+)"`)
	manifestRustVer  = regexp.MustCompile(`(?m)^[ \t]*rust-version[ \t]*=[ \t]*"([^"]+)"`)
)

// FindRustPin is the Rust toolchain the workspace pins: the channel of its
// rust-toolchain.toml, else the rust-version of its root Cargo.toml. It is empty
// when the workspace pins none.
func FindRustPin(workspaceRoot string) string {
	for _, f := range []struct {
		file string
		re   *regexp.Regexp
	}{{"rust-toolchain.toml", toolchainChannel}, {"Cargo.toml", manifestRustVer}} {
		src, err := os.ReadFile(filepath.Join(workspaceRoot, f.file))
		if err != nil {
			continue
		}
		if m := f.re.FindStringSubmatch(string(src)); m != nil {
			return m[1]
		}
	}
	return ""
}
