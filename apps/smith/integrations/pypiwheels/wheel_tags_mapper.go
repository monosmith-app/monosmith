package pypiwheels

import (
	"fmt"
	"regexp"
	"strings"
)

// platformTags are the wheel platform tags a static Go binary for each
// os/arch runs under. Linux binaries need no libc, so one wheel serves glibc
// (manylinux) and musl (musllinux). Go 1.25 and later need macOS 12.
var platformTags = map[string][]string{
	"linux/amd64":   {"manylinux_2_17_x86_64", "manylinux2014_x86_64", "musllinux_1_1_x86_64"},
	"linux/arm64":   {"manylinux_2_17_aarch64", "manylinux2014_aarch64", "musllinux_1_1_aarch64"},
	"darwin/amd64":  {"macosx_12_0_x86_64"},
	"darwin/arm64":  {"macosx_12_0_arm64"},
	"windows/amd64": {"win_amd64"},
	"windows/arm64": {"win_arm64"},
}

// Tags are a platform's wheel tags; the binary needs no Python, so any
// Python 3 and no ABI.
func Tags(goos, goarch string) ([]string, bool) {
	platforms, ok := platformTags[goos+"/"+goarch]
	if !ok {
		return nil, false
	}
	tags := make([]string, len(platforms))
	for i, p := range platforms {
		tags[i] = "py3-none-" + p
	}
	return tags, true
}

var separators = regexp.MustCompile(`[-_.]+`)

// Distribution is a project name as wheel file names spell it (PEP 427):
// my-tool → my_tool.
func Distribution(name string) string {
	return separators.ReplaceAllString(strings.ToLower(name), "_")
}

var prerelease = regexp.MustCompile(`^(\d+\.\d+\.\d+)-(alpha|beta|rc|a|b)\.?(\d+)$`)

// Version is a semantic version as Python spells it (PEP 440):
// 1.2.0-rc.1 → 1.2.0rc1. A prerelease other than alpha, beta or rc has no
// Python spelling.
func Version(semver string) (string, error) {
	if !strings.ContainsAny(semver, "-+") {
		return semver, nil
	}
	m := prerelease.FindStringSubmatch(semver)
	if m == nil {
		return "", fmt.Errorf("version %s has no Python spelling (use -alpha.N, -beta.N or -rc.N)", semver)
	}
	kind := map[string]string{"alpha": "a", "a": "a", "beta": "b", "b": "b", "rc": "rc"}[m[2]]
	return m[1] + kind + m[3], nil
}
