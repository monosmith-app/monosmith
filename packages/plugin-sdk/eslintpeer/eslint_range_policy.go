package eslintpeer

import (
	"regexp"
	"strconv"
	"strings"
)

// simpleRange is one version with an optional ^, ~, >= or = before it: the
// ranges whose lowest version can be read.
var simpleRange = regexp.MustCompile(`^(?:\^|~|>=|=)?v?(\d+)(?:\.(\d+))?(?:\.(\d+))?$`)

// RangeFloor is the lowest version a simple range accepts ("^9" is 9.0.0).
//
// Parameters: r is a package.json version range.
// Returns the version and true, or false for any range that is not one
// version with an optional ^, ~, >= or = before it.
func RangeFloor(r string) (string, bool) {
	m := simpleRange.FindStringSubmatch(strings.TrimSpace(r))
	if m == nil {
		return "", false
	}
	part := func(s string) string {
		if s == "" {
			return "0"
		}
		return s
	}
	return m[1] + "." + part(m[2]) + "." + part(m[3]), true
}

// Meets says whether an eslint version is one the config runs with: the
// major of Floor, at least Floor.
//
// Parameters: version is a plain version ("9.39.3"), without a range prefix.
// Returns true when it is inside Range.
func Meets(version string) bool {
	got := numbers(version)
	floor := numbers(Floor)
	if got[0] != floor[0] {
		return false
	}
	for i := 1; i < 3; i++ {
		if got[i] != floor[i] {
			return got[i] > floor[i]
		}
	}
	return true
}

// Raised is the eslint range a package.json names in place of current.
//
// Parameters: current is the range the package.json names now.
// Returns Range and true when current is a simple range below Range (or of
// another major), and current and false when it meets Range or is a range
// smith cannot read (a tag, a path, "$eslint").
func Raised(current string) (string, bool) {
	floor, ok := RangeFloor(current)
	if !ok || Meets(floor) {
		return current, false
	}
	return Range, true
}

// numbers are the major, minor and patch of a plain version; a part that is
// absent or not a number is 0. A pre-release or build suffix is ignored.
func numbers(version string) [3]int {
	var out [3]int
	core, _, _ := strings.Cut(strings.TrimPrefix(version, "v"), "-")
	core, _, _ = strings.Cut(core, "+")
	for i, p := range strings.SplitN(core, ".", 3) {
		out[i], _ = strconv.Atoi(p)
	}
	return out
}
