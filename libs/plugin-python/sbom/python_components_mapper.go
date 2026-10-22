package sbom

import (
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/pelletier/go-toml/v2"
)

// ComponentsFromLock lists the packages a uv.lock or poetry.lock pins. The
// project itself and its workspace members (editable, virtual or directory
// sources) are left out.
func ComponentsFromLock(lock []byte) ([]protocol.Component, error) {
	var doc struct {
		Package []struct {
			Name    string         `toml:"name"`
			Version string         `toml:"version"`
			Source  map[string]any `toml:"source"`
		} `toml:"package"`
	}
	if err := toml.Unmarshal(lock, &doc); err != nil {
		return nil, err
	}
	var out []protocol.Component
	for _, p := range doc.Package {
		if p.Name == "" || local(p.Source) {
			continue
		}
		out = append(out, pypi(p.Name, p.Version))
	}
	return components.Dedupe(out), nil
}

func local(source map[string]any) bool {
	for _, k := range []string{"editable", "virtual", "directory"} {
		if _, ok := source[k]; ok {
			return true
		}
	}
	return false
}

// ComponentsFromRequirements lists the requirements of a requirements.txt:
// a pinned one by its version, the others by their specifier.
func ComponentsFromRequirements(text string) []protocol.Component {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if i := strings.Index(line, " #"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") || strings.Contains(line, "://") {
			continue
		}
		lines = append(lines, line)
	}
	return fromRequirements(lines)
}

// ComponentsFromPyproject lists project.dependencies of a pyproject.toml.
func ComponentsFromPyproject(manifest []byte) ([]protocol.Component, error) {
	var doc struct {
		Project struct {
			Dependencies []string `toml:"dependencies"`
		} `toml:"project"`
	}
	if err := toml.Unmarshal(manifest, &doc); err != nil {
		return nil, err
	}
	return fromRequirements(doc.Project.Dependencies), nil
}

var requirement = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9._-]*)\s*(?:\[[^\]]*\])?\s*(.*)$`)

func fromRequirements(reqs []string) []protocol.Component {
	var out []protocol.Component
	for _, r := range reqs {
		if i := strings.Index(r, ";"); i >= 0 {
			r = r[:i]
		}
		m := requirement.FindStringSubmatch(strings.TrimSpace(r))
		if m == nil {
			continue
		}
		spec := strings.Trim(strings.ReplaceAll(m[2], " ", ""), "()")
		if rest, ok := strings.CutPrefix(spec, "=="); ok && !strings.ContainsAny(rest, ",<>!~") {
			spec = rest
		}
		out = append(out, pypi(m[1], spec))
	}
	return components.Dedupe(out)
}

// pypi normalises the name as PEP 503 does, which is how the package URL
// spells it.
func pypi(name, version string) protocol.Component {
	norm := strings.ToLower(regexp.MustCompile(`[-_.]+`).ReplaceAllString(name, "-"))
	return components.New("pypi", "", norm, version)
}
