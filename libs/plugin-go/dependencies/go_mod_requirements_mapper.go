package dependencies

import (
	"regexp"
	"strings"
)

// Requirement is one `require` entry of a go.mod.
type Requirement struct {
	Path, Version string
	// Indirect entries are go.mod's bookkeeping for what dependencies need.
	Indirect bool
}

var (
	requireLine = regexp.MustCompile(`^(?:require\s+)?(\S+)\s+(v\S+?)\s*(//.*)?$`)
	// release is a tagged version, with an optional pre-release; a
	// pseudo-version (a commit between tags) is not one.
	release = regexp.MustCompile(`^v\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+incompatible)?$`)
	pseudo  = regexp.MustCompile(`-\d{14}-[0-9a-f]{12}$`)
)

// ParseRequirements reads the require directives of a go.mod, single lines
// and blocks alike. Everything else (replace, exclude, tool) is skipped.
func ParseRequirements(gomod string) []Requirement {
	var out []Requirement
	inBlock, skipBlock := false, false
	for _, raw := range strings.Split(gomod, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case inBlock && line == ")":
			inBlock, skipBlock = false, false
			continue
		case !inBlock && strings.HasSuffix(line, "(") && !strings.HasPrefix(line, "//"):
			inBlock = true
			skipBlock = strings.TrimSpace(strings.TrimSuffix(line, "(")) != "require"
			continue
		case skipBlock, line == "", strings.HasPrefix(line, "//"):
			continue
		case !inBlock && !strings.HasPrefix(line, "require "):
			continue
		}
		m := requireLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		out = append(out, Requirement{Path: m[1], Version: m[2], Indirect: strings.Contains(m[3], "indirect")})
	}
	return out
}

// Rewritable reports whether sync and up may change the version: a tagged
// release, not a pseudo-version.
func Rewritable(version string) bool {
	return release.MatchString(version) && !pseudo.MatchString(version)
}
