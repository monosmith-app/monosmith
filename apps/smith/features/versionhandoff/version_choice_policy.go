package versionhandoff

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// versionPattern is a version smith publishes: 2.50.0, 3.0.0-beta.1. It is the
// pattern of the launchers (packages/launcher-node, choose-version.policy.mjs).
var versionPattern = regexp.MustCompile(`(?i)^\d+\.\d+\.\d+(?:-[0-9A-Z.-]+)?$`)

// The lines of smith.toml that PinIn reads: a table header, and smith = "<version>".
var (
	tableLine = regexp.MustCompile(`^\[([^\]]+)\]$`)
	pinLine   = regexp.MustCompile(`^smith\s*=\s*["']([^"']+)["']$`)
	comment   = regexp.MustCompile(`#.*$`)
)

// PinIn is the [workspace] smith pin in the text of a smith.toml.
//
// Parameters: text is the text of smith.toml.
// Returns: the pinned version, or "" for none, for latest and for a value that
// is not a version.
func PinIn(text string) string {
	table := ""
	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(comment.ReplaceAllString(raw, ""))
		if header := tableLine.FindStringSubmatch(line); header != nil {
			table = strings.TrimSpace(header[1])
			continue
		}
		if pin := pinLine.FindStringSubmatch(line); table == "workspace" && pin != nil {
			if versionPattern.MatchString(pin[1]) {
				return pin[1]
			}
			return ""
		}
	}
	return ""
}

// ChooseVersion is the version a run uses: SMITH_VERSION, else the workspace's
// pin, else none (the newest installed).
//
// Parameters: wanted is SMITH_VERSION; pin reads the workspace's pin, and is
// called only when SMITH_VERSION does not decide.
// Returns: an exact version, or "" for the newest; an error when SMITH_VERSION
// is set to something that is not a version.
func ChooseVersion(wanted string, pin func() string) (string, error) {
	wanted = strings.TrimSpace(wanted)
	switch {
	case wanted == "latest":
		return "", nil
	case wanted != "" && !versionPattern.MatchString(wanted):
		return "", fmt.Errorf("SMITH_VERSION is %q: use a version such as 2.50.0, or latest", wanted)
	case wanted != "":
		return wanted, nil
	}
	return pin(), nil
}

// CompareVersions compares two versions by their numbers; a prerelease sorts
// before its release, and two prereleases compare as text.
//
// Parameters: a and b are versions.
// Returns: a negative number when a is older, a positive one when it is newer,
// and 0 when they are equal.
func CompareVersions(a, b string) int {
	mainA, preA := split(a)
	mainB, preB := split(b)
	numsA, numsB := strings.Split(mainA, "."), strings.Split(mainB, ".")
	for i := range 3 {
		if x, y := number(numsA, i), number(numsB, i); x != y {
			return x - y
		}
	}
	switch {
	case preA == preB:
		return 0
	case preA == "":
		return 1
	case preB == "":
		return -1
	}
	return strings.Compare(preA, preB)
}

// split is a version's numbers and its prerelease, as the launchers read them:
// the text after a second "-" is not part of the prerelease.
func split(v string) (string, string) {
	parts := strings.Split(v, "-")
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// number is the i-th number of a version, 0 when it is missing.
func number(nums []string, i int) int {
	if i >= len(nums) {
		return 0
	}
	n, _ := strconv.Atoi(nums[i])
	return n
}
