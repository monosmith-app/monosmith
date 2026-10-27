package discovery

import (
	"regexp"
	"strings"
)

// PyProject is the subset of pyproject.toml the plugin reads.
type PyProject struct {
	Project struct {
		Name         string            `toml:"name"`
		Version      string            `toml:"version"`
		Description  string            `toml:"description"`
		Dependencies []string          `toml:"dependencies"`
		Scripts      map[string]string `toml:"scripts"`
	} `toml:"project"`
	Tool struct {
		Smith struct {
			Kind  string `toml:"kind"`
			Stack string `toml:"stack"`
		} `toml:"smith"`
	} `toml:"tool"`
}

var requirementName = regexp.MustCompile(`^\s*([A-Za-z0-9][A-Za-z0-9._-]*)`)

// ImportName normalises a distribution name the way Python packages are
// conventionally imported: lowercase, with - and . turned into _.
func ImportName(distribution string) string {
	return strings.NewReplacer("-", "_", ".", "_").Replace(strings.ToLower(distribution))
}

// RequirementImportName returns the import name of a PEP 508 requirement
// string, or "" if it cannot be parsed.
func RequirementImportName(requirement string) string {
	m := requirementName.FindStringSubmatch(requirement)
	if m == nil {
		return ""
	}
	return ImportName(m[1])
}
