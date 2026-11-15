package rubygems

import (
	"fmt"
	"regexp"
	"strings"
)

// platforms are the RubyGems platforms a static Go binary for each os/arch
// runs under. RubyGems 3.3.22 and later install an x86_64-linux gem on glibc
// only, so musl gets its own gem with the same binary. Darwin gems name no OS
// version, so they install on every macOS.
var platforms = map[string][]string{
	"linux/amd64":   {"x86_64-linux", "x86_64-linux-musl"},
	"linux/arm64":   {"aarch64-linux", "aarch64-linux-musl"},
	"darwin/amd64":  {"x86_64-darwin"},
	"darwin/arm64":  {"arm64-darwin"},
	"windows/amd64": {"x64-mingw-ucrt"},
	"windows/arm64": {"aarch64-mingw-ucrt"},
}

// Platforms are the gem platforms an os/arch archive is published for.
func Platforms(goos, goarch string) ([]string, bool) {
	p, ok := platforms[goos+"/"+goarch]
	return p, ok
}

// FileName is a platform gem's file name.
func FileName(name, version, platform string) string {
	return fmt.Sprintf("%s-%s-%s.gem", name, version, platform)
}

var gemVersion = regexp.MustCompile(`^[0-9]+(\.[0-9a-zA-Z]+)*$`)

// Version is a semantic version as RubyGems spells it: 1.2.0-rc.1 →
// 1.2.0.pre.rc.1, which RubyGems sorts before 1.2.0. Build metadata (+...)
// has no gem spelling and is dropped.
func Version(semver string) (string, error) {
	v, _, _ := strings.Cut(semver, "+")
	v = strings.Replace(v, "-", ".pre.", 1)
	if !gemVersion.MatchString(v) {
		return "", fmt.Errorf("version %s has no RubyGems spelling", semver)
	}
	return v, nil
}
