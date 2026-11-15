package adoptsplit

import (
	"regexp"
	"strings"
)

// pythonToolNames are the development dependencies that are tools, by name
// (lower case, with - for _ and .).
var pythonToolNames = map[string]bool{
	"pytest": true, "ruff": true, "mypy": true, "black": true, "isort": true, "flake8": true, "pylint": true,
	"pyright": true, "coverage": true, "tox": true, "nox": true, "pre-commit": true, "build": true, "twine": true,
	"pip-audit": true, "bandit": true, "codespell": true, "autopep8": true, "yapf": true,
}

// pythonToolPrefixes are the families of tools: pytest plugins, flake8
// plugins, and type stubs.
var pythonToolPrefixes = []string{"pytest-", "flake8-", "types-", "mypy-"}

var requirementName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*`)

// RequirementName is the normalised name of a requirement such as
// "pytest[cov]>=8; python_version>'3.9'": lower case, with - for _ and .
// It is "" when the text names nothing.
//
// Parameters: requirement is one entry of a dependency list.
func RequirementName(requirement string) string {
	name := requirementName.FindString(strings.TrimSpace(requirement))
	return strings.NewReplacer("_", "-", ".", "-").Replace(strings.ToLower(name))
}

// IsPythonTool says whether a development dependency is a tool of the
// workspace: a test runner, linter, formatter, type checker or build tool.
// A name not listed is not a tool, and the person is asked about it.
//
// Parameters: name is a requirement name, as RequirementName returns it.
func IsPythonTool(name string) bool {
	if pythonToolNames[name] {
		return true
	}
	for _, prefix := range pythonToolPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
