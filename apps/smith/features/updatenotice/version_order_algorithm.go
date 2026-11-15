package updatenotice

import (
	"strconv"
	"strings"
)

// version is a parsed "1.73.0" or "2.0.0-beta.22".
type version struct {
	n   [3]int
	pre string
}

// parse reads "1.73.0", "v1.73.0-rc.1" or "2.0.0-beta.22"; ok is false for
// anything else (a "dev" build, an empty string).
func parse(v string) (out version, ok bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.Index(v, "+"); i >= 0 {
		v = v[:i]
	}
	if i := strings.Index(v, "-"); i >= 0 {
		v, out.pre = v[:i], v[i+1:]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		x, err := strconv.Atoi(p)
		if err != nil || x < 0 {
			return out, false
		}
		out.n[i] = x
	}
	return out, true
}

// Newer is true when version a is higher than version b, by the semver rules:
// a release is higher than its own prereleases, and prerelease parts compare
// one by one, numbers by value (beta.9 before beta.10). A version that is not
// numbers (a development build) is never newer and never older.
func Newer(a, b string) bool {
	x, okA := parse(a)
	y, okB := parse(b)
	if !okA || !okB {
		return false
	}
	for i := range x.n {
		if x.n[i] != y.n[i] {
			return x.n[i] > y.n[i]
		}
	}
	switch {
	case x.pre == y.pre:
		return false
	case x.pre == "":
		return true
	case y.pre == "":
		return false
	}
	return comparePre(x.pre, y.pre) > 0
}

// comparePre orders two prerelease strings by the semver rules.
func comparePre(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		an, aerr := strconv.Atoi(as[i])
		bn, berr := strconv.Atoi(bs[i])
		switch {
		case aerr == nil && berr == nil && an != bn:
			return cmpInt(an, bn)
		case aerr == nil && berr != nil:
			return -1
		case aerr != nil && berr == nil:
			return 1
		case aerr != nil && berr != nil && as[i] != bs[i]:
			return strings.Compare(as[i], bs[i])
		}
	}
	return cmpInt(len(as), len(bs))
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// ChannelOf is the npm dist-tag smith checks for a running version: "next"
// for a prerelease (2.0.0-beta.22), "latest" for a stable release. A person
// who installed the next channel hears about the next channel only.
func ChannelOf(v string) string {
	if p, ok := parse(v); ok && p.pre != "" {
		return "next"
	}
	return "latest"
}
