package webenv

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var nameShape = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// ParseList reads a list of environment names typed by a person or given to
// --environments: names separated by commas or spaces. It trims each name and
// drops empty ones; it does not check them.
func ParseList(text string) []string {
	var out []string
	for _, f := range strings.FieldsFunc(text, func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		out = append(out, strings.TrimSpace(f))
	}
	return out
}

// reserved are names a build cannot use: Vite already gives them a meaning.
var reserved = []string{"development", "production", "test", "local", "example"}

// ValidateNames checks the environment names: at least one, each a lowercase
// word (letters, digits and dashes, starting with a letter), none twice, none
// that Vite or the env files already mean (development, production, test,
// local, example).
//
// It returns an error that names the first name that is wrong and why.
func ValidateNames(names []string) error {
	if len(names) == 0 {
		return fmt.Errorf("give at least one environment name, such as %s", strings.Join(DefaultNames, ", "))
	}
	seen := map[string]bool{}
	for _, n := range names {
		switch {
		case !nameShape.MatchString(n):
			return fmt.Errorf("%q is not an environment name: use a lowercase word (letters, digits and dashes, starting with a letter), such as uat", n)
		case slices.Contains(reserved, n):
			return fmt.Errorf("%q cannot be an environment name: Vite or the env files already give it a meaning. Use another word, such as lcl for local", n)
		case seen[n]:
			return fmt.Errorf("the environment %q is listed twice", n)
		}
		seen[n] = true
	}
	return nil
}

// Validate checks the whole settings: the names, a default that is one of
// them, and the two options.
//
// It returns an error that says what is wrong and what is allowed.
func (s Settings) Validate() error {
	if err := ValidateNames(s.Names); err != nil {
		return err
	}
	n := s.Normalized()
	if !slices.Contains(n.Names, n.Default) {
		return fmt.Errorf("the default environment %q is not in the list (%s): pick one of them", n.Default, strings.Join(n.Names, ", "))
	}
	if n.Output != OutputWorkspace && n.Output != OutputProject {
		return fmt.Errorf("the output place %q is not %s or %s", n.Output, OutputWorkspace, OutputProject)
	}
	if n.EnvFiles != FilesCommit && n.EnvFiles != FilesExample {
		return fmt.Errorf("the env files choice %q is not %s or %s", n.EnvFiles, FilesCommit, FilesExample)
	}
	return nil
}
