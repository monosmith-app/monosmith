package semver

import (
	"strconv"
	"strings"
)

// Compare orders two dotted version strings. It returns a negative number when
// a is older than b, 0 when they are equal and a positive number when a is
// newer. A build suffix (+meta) is ignored, a pre-release (-beta.1) sorts
// before the release it leads to, and when a part is not a number the two
// strings compare as text.
func Compare(a, b string) int {
	ca, pa, _ := strings.Cut(strings.SplitN(a, "+", 2)[0], "-")
	cb, pb, _ := strings.Cut(strings.SplitN(b, "+", 2)[0], "-")
	xa, xb := strings.Split(ca, "."), strings.Split(cb, ".")
	for i := 0; i < len(xa) || i < len(xb); i++ {
		var na, nb int
		var ea, eb error
		if i < len(xa) {
			na, ea = strconv.Atoi(xa[i])
		}
		if i < len(xb) {
			nb, eb = strconv.Atoi(xb[i])
		}
		if ea != nil || eb != nil {
			return strings.Compare(a, b)
		}
		if na != nb {
			if na < nb {
				return -1
			}
			return 1
		}
	}
	switch {
	case pa == pb:
		return 0
	case pa == "":
		return 1
	case pb == "":
		return -1
	}
	return strings.Compare(pa, pb)
}
