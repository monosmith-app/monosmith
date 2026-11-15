package pluginhost

import (
	"fmt"
	"strconv"
	"strings"
)

// Satisfies reports whether a plugin version meets a constraint from
// smith.toml: empty or "*" (any), "1.2.3" (exactly), "^1.2" (same major, at
// least 1.2), "~1.2" (same major and minor, at least 1.2), or ">=1.2".
func Satisfies(version, constraint string) (bool, error) {
	constraint = strings.TrimSpace(constraint)
	if constraint == "" || constraint == "*" {
		return true, nil
	}
	v, err := parseVersion(version)
	if err != nil {
		return false, fmt.Errorf("plugin version %q: %w", version, err)
	}
	op, rest := "=", constraint
	for _, p := range []string{">=", "^", "~", "="} {
		if strings.HasPrefix(constraint, p) {
			op, rest = p, strings.TrimSpace(constraint[len(p):])
			break
		}
	}
	want, err := parseVersion(rest)
	if err != nil {
		return false, fmt.Errorf("version constraint %q: %w", constraint, err)
	}
	atLeast := compare(v, want) >= 0
	switch op {
	case ">=":
		return atLeast, nil
	case "^":
		return atLeast && v[0] == want[0], nil
	case "~":
		return atLeast && v[0] == want[0] && v[1] == want[1], nil
	}
	return compare(v, want) == 0, nil
}

// parseVersion reads major[.minor[.patch]], ignoring a leading v and any
// pre-release or build suffix.
func parseVersion(s string) ([3]int, error) {
	var v [3]int
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if s == "" || len(parts) > 3 {
		return v, fmt.Errorf("not a version")
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return v, fmt.Errorf("not a version")
		}
		v[i] = n
	}
	return v, nil
}

func compare(a, b [3]int) int {
	for i := range a {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}
