package scaffold

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

// ValidateName accepts kebab-case names that are valid as a directory, an npm
// package name, a Go path element and (with - → _) a Python import name.
func ValidateName(kind, name string) error {
	if len(name) > 50 || !namePattern.MatchString(name) {
		return fmt.Errorf("%s name %q must be lowercase letters and digits with words joined by single hyphens, starting with a letter (e.g. web-app)", kind, name)
	}
	return nil
}

// SplitKind reads a combined kind ("node-app", "python-package") as a kind
// and a language.
func SplitKind(kind string) (string, string, bool) {
	i := strings.LastIndex(kind, "-")
	if i <= 0 {
		return kind, "", false
	}
	switch k := protocol.ProjectKind(kind[i+1:]); k {
	case protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage:
		return string(k), kind[:i], true
	}
	return kind, "", false
}
