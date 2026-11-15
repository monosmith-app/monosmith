package dependencies

import (
	"strconv"
	"strings"
)

// majorOf is the first number of the version a spec names, or -1.
func majorOf(spec string) int {
	_, v, ok := splitSpec(spec)
	if !ok {
		return -1
	}
	n, err := strconv.Atoi(strings.SplitN(v, ".", 2)[0])
	if err != nil {
		return -1
	}
	return n
}

// crossesMajor reports whether moving a declaration from spec to target
// changes the major version. A spec with no readable major never crosses.
func crossesMajor(spec, target string) bool {
	a, b := majorOf(spec), majorOf(target)
	return a >= 0 && b >= 0 && a != b
}
