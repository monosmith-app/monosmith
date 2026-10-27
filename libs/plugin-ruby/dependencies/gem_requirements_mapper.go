package dependencies

import (
	"regexp"
	"strings"
)

// Where a gem is declared.
const (
	SectionDependencies = "dependencies"
	SectionDev          = "development_dependencies"
)

// GemRequirement is one gem a gemspec or Gemfile declares, and the version
// constraint it gives.
type GemRequirement struct {
	Section string
	Name    string
	// Constraint is the constraint as written: "~> 1.2", or several joined
	// with ", " (">= 1.0, < 2"). Gems declared without one are not listed.
	Constraint string
}

var (
	gemspecCall = regexp.MustCompile(`^\s*\w+\.add_(runtime_|development_)?dependency\s*\(?\s*(['"])([^'"]+)['"]((?:\s*,\s*['"][^'"]*['"])*)`)
	gemfileCall = regexp.MustCompile(`^\s*gem\s+\(?\s*(['"])([^'"]+)['"]((?:\s*,\s*['"][^'"]*['"])*)`)
	quoted      = regexp.MustCompile(`['"]([^'"]*)['"]`)
	groupOpen   = regexp.MustCompile(`^(\s*)group\s+(.*)\s+do\b`)
)

func constraintOf(rest string) string {
	var parts []string
	for _, m := range quoted.FindAllStringSubmatch(rest, -1) {
		parts = append(parts, strings.TrimSpace(m[1]))
	}
	return strings.Join(parts, ", ")
}

// ParseGemspec reads the dependencies a gemspec declares with a version
// constraint: add_dependency (and add_runtime_dependency) as production,
// add_development_dependency as development.
func ParseGemspec(src string) []GemRequirement {
	var out []GemRequirement
	for _, line := range strings.Split(src, "\n") {
		m := gemspecCall.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		constraint := constraintOf(m[4])
		if constraint == "" {
			continue
		}
		section := SectionDependencies
		if m[1] == "development_" {
			section = SectionDev
		}
		out = append(out, GemRequirement{Section: section, Name: m[3], Constraint: constraint})
	}
	return out
}

// ParseGemfile reads the gems a Gemfile declares with a version constraint;
// those inside a development or test group are development dependencies.
func ParseGemfile(src string) []GemRequirement {
	var out []GemRequirement
	devIndent := -1
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimRight(line, "\r")
		if devIndent >= 0 && strings.TrimSpace(trimmed) == "end" && len(trimmed)-len(strings.TrimLeft(trimmed, " \t")) == devIndent {
			devIndent = -1
			continue
		}
		if m := groupOpen.FindStringSubmatch(trimmed); m != nil && devIndent < 0 &&
			(strings.Contains(m[2], "development") || strings.Contains(m[2], "test")) {
			devIndent = len(m[1])
			continue
		}
		m := gemfileCall.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		constraint := constraintOf(m[3])
		if constraint == "" {
			continue
		}
		section := SectionDependencies
		if devIndent >= 0 {
			section = SectionDev
		}
		out = append(out, GemRequirement{Section: section, Name: m[2], Constraint: constraint})
	}
	return out
}

// plainConstraint is one constraint sync and up understand: a version,
// optionally with ~>, >=, <=, > or <. Several constraints, != and
// pre-releases are left alone.
var plainConstraint = regexp.MustCompile(`^(~>|>=|<=|>|<|=)?\s*\d+(\.\d+)*$`)

// Rewritable reports whether sync and up may change a constraint.
func Rewritable(constraint string) bool { return plainConstraint.MatchString(constraint) }
