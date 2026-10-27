package dependencies

import (
	"encoding/json"
	"regexp"
	"strings"
)

// The sections of a composer.json that name Packagist packages.
const (
	SectionRequire = "require"
	SectionDev     = "require-dev"
)

type composerJSON struct {
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
}

// ParseComposer reads the packages a composer.json requires, by section.
// Platform requirements (php, ext-*, lib-*, composer-plugin-api) say what
// the package runs on, not what is installed from Packagist, and are left
// out.
func ParseComposer(src []byte) (require, requireDev map[string]string, err error) {
	var c composerJSON
	if err := json.Unmarshal(src, &c); err != nil {
		return nil, nil, err
	}
	keep := func(in map[string]string) map[string]string {
		out := map[string]string{}
		for name, constraint := range in {
			if strings.Contains(name, "/") {
				out[name] = strings.TrimSpace(constraint)
			}
		}
		return out
	}
	return keep(c.Require), keep(c.RequireDev), nil
}

// plainConstraint is one constraint sync and up understand: a version,
// optionally with ^, ~, >=, <=, > or <. Alternatives (^1.0 || ^2.0),
// ranges, wildcards (1.2.*) and branches (dev-main) are left alone.
var plainConstraint = regexp.MustCompile(`^(\^|~|>=|<=|>|<|=)?\s*\d+(\.\d+){0,2}$`)

// Rewritable reports whether sync and up may change a constraint.
func Rewritable(constraint string) bool { return plainConstraint.MatchString(constraint) }
