package declaredimports

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	modulePath = regexp.MustCompile(`(?m)^module\s+"?([^\s"]+)"?\s*$`)
	toolLine   = regexp.MustCompile(`^(?:tool\s+)?([^\s()/][^\s()]*)\s*(?://.*)?$`)
)

// modulePathOf is the module line of a go.mod, or "".
func modulePathOf(gomod string) string {
	if m := modulePath.FindStringSubmatch(gomod); m != nil {
		return m[1]
	}
	return ""
}

// toolPackages lists the packages of the go.mod's tool directives (Go 1.24),
// single lines and blocks alike.
func toolPackages(gomod string) []string {
	var out []string
	inBlock := false
	for _, raw := range strings.Split(gomod, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case inBlock && line == ")":
			inBlock = false
		case !inBlock && strings.HasPrefix(line, "tool") && strings.HasSuffix(line, "("):
			inBlock = true
		case inBlock || strings.HasPrefix(line, "tool "):
			if m := toolLine.FindStringSubmatch(line); m != nil {
				out = append(out, m[1])
			}
		}
	}
	return out
}

// supportsTidyDiff reports whether the go command that printed versionOutput
// (go version) has `go mod tidy -diff`, which arrived in Go 1.23.
func supportsTidyDiff(versionOutput string) bool {
	_, rest, ok := strings.Cut(versionOutput, "go version go")
	if !ok {
		return false
	}
	numbers := strings.SplitN(strings.Fields(rest)[0], ".", 3)
	if len(numbers) < 2 {
		return false
	}
	major, err1 := strconv.Atoi(numbers[0])
	digits := numbers[1]
	if end := strings.IndexFunc(digits, func(r rune) bool { return r < '0' || r > '9' }); end >= 0 {
		digits = digits[:end]
	}
	minor, err2 := strconv.Atoi(digits)
	return err1 == nil && err2 == nil && (major > 1 || (major == 1 && minor >= 23))
}
