package packageprompt

import (
	"fmt"
	"strings"
	"unicode"
)

// ParsePackages splits what was typed into package names. A name that starts
// with a dash would reach the package manager as a flag, so it is refused and
// the person is told where the option lives instead.
func ParsePackages(text string) ([]string, error) {
	names := strings.Fields(text)
	if len(names) == 0 {
		return nil, fmt.Errorf("type at least one package, such as zod or react@19")
	}
	for _, n := range names {
		if strings.HasPrefix(n, "-") {
			return nil, fmt.Errorf("%q is an option, not a package: the box below sets the development dependency", n)
		}
		if strings.ContainsFunc(n, unicode.IsControl) {
			return nil, fmt.Errorf("%q has a control character", n)
		}
	}
	return names, nil
}
