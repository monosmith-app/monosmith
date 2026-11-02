package toolinstall

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/toolchainpin"
)

// Requirement is the version range a workspace wants of a tool. The zero value
// accepts every version.
type Requirement struct {
	// Prefix is a version prefix on dot boundaries: "22" accepts 22.3.0, not 220.0.0.
	Prefix string `json:"prefix,omitempty"`
	// Min is the lowest version accepted.
	Min string `json:"min,omitempty"`
}

// Any reports whether the requirement accepts every version.
func (r Requirement) Any() bool { return r.Prefix == "" && r.Min == "" }

// String words the requirement for a person: "22.x", "1.23.4 or newer".
func (r Requirement) String() string {
	switch {
	case r.Prefix != "" && r.Min != "":
		return r.Min + " or newer in " + r.Prefix + ".x"
	case r.Prefix != "":
		return r.Prefix + ".x"
	case r.Min != "":
		return r.Min + " or newer"
	}
	return "any version"
}

// Satisfied reports whether an installed version meets the requirement. An
// empty installed version (the program printed none) is accepted: smith cannot
// tell, and does not offer to replace what runs.
func (r Requirement) Satisfied(installed string) bool {
	if installed == "" || r.Any() {
		return true
	}
	if r.Prefix != "" && !toolchainpin.Satisfies(r.Prefix, installed) {
		return false
	}
	return r.Min == "" || CompareVersions(installed, r.Min) >= 0
}

// Version is the version an install should follow: the pinned prefix, else the
// minimum, else the fallback.
func (r Requirement) Version(fallback string) string {
	switch {
	case r.Prefix != "":
		return r.Prefix
	case r.Min != "":
		return r.Min
	}
	return fallback
}

var versionNumber = regexp.MustCompile(`\d+(?:\.\d+)+`)

// FindVersion returns the first version number in the text a program printed
// ("go version go1.25.1 windows/amd64" gives "1.25.1"), or "".
func FindVersion(output string) string { return versionNumber.FindString(output) }

// CompareVersions compares dotted numbers component by component: it returns
// a negative number when a is older than b, 0 when equal and a positive number
// when a is newer. A missing component counts as 0.
func CompareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < max(len(as), len(bs)); i++ {
		x, y := component(as, i), component(bs, i)
		if x != y {
			return x - y
		}
	}
	return 0
}

func component(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, _ := strconv.Atoi(parts[i])
	return n
}

// Fill replaces {major} and {major.minor} in a package name with the parts of
// a version. A version with fewer parts fills the missing ones with 0.
func Fill(pkg, version string) string {
	if !strings.Contains(pkg, "{") {
		return pkg
	}
	parts := strings.Split(version, ".")
	major, minor := parts[0], "0"
	if len(parts) > 1 {
		minor = parts[1]
	}
	pkg = strings.ReplaceAll(pkg, "{major.minor}", major+"."+minor)
	return strings.ReplaceAll(pkg, "{major}", major)
}
