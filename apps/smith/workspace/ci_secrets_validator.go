package workspace

import (
	"fmt"
	"regexp"
	"strings"
)

var secretName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// validateCISecrets refuses a [ci] secrets name that is no environment
// variable name, that repeats, or that starts with GITHUB_, a prefix GitHub
// reserves for its own variables.
func validateCISecrets(names []string) error {
	seen := map[string]bool{}
	for _, n := range names {
		switch {
		case !secretName.MatchString(n):
			return fmt.Errorf("%s: [ci] secrets must be names of environment variables, such as \"API_TOKEN\", not %q", FileName, n)
		case strings.HasPrefix(strings.ToUpper(n), "GITHUB_"):
			return fmt.Errorf("%s: [ci] secrets cannot name %s: GitHub reserves the GITHUB_ prefix", FileName, n)
		case seen[n]:
			return fmt.Errorf("%s: [ci] secrets lists %s twice", FileName, n)
		}
		seen[n] = true
	}
	return nil
}
