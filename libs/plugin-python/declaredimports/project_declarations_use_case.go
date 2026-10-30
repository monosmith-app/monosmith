package declaredimports

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// declarations are the distributions a project declares, by normalised name.
type declarations struct {
	// runtime are installed with the package: dependencies and extras.
	runtime map[string]bool
	// dev are not: dependency groups, development requirements, build
	// requirements, and the workspace's requirements-dev.txt.
	dev map[string]bool
	// known is false when the declarations cannot be read statically
	// (dynamic dependencies, no [project] table).
	known bool
}

var (
	requirementName = regexp.MustCompile(`^\s*([A-Za-z0-9][A-Za-z0-9._-]*)`)
	separators      = regexp.MustCompile(`[-_.]+`)
)

// normalise is PEP 503's name normalisation.
func normalise(name string) string {
	return separators.ReplaceAllString(strings.ToLower(name), "-")
}

func (d *declarations) add(into map[string]bool, requirement string) {
	if strings.HasPrefix(strings.TrimSpace(requirement), "-") { // -r other.txt, --index-url
		return
	}
	if m := requirementName.FindStringSubmatch(requirement); m != nil {
		into[normalise(m[1])] = true
	}
}

func (d *declarations) addAll(into map[string]bool, list any) {
	switch v := list.(type) {
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				d.add(into, s)
			}
		}
	case map[string]any: // [tool.poetry.dependencies]: name = version
		for name := range v {
			if name != "python" {
				into[normalise(name)] = true
			}
		}
	}
}

func table(m map[string]any, keys ...string) map[string]any {
	for _, k := range keys {
		next, ok := m[k].(map[string]any)
		if !ok {
			return nil
		}
		m = next
	}
	return m
}

// readDeclarations reads what the project at dir declares, and what the
// workspace's requirements-dev.txt adds for its development files.
func readDeclarations(workspaceRoot, dir string) (declarations, error) {
	d := declarations{runtime: map[string]bool{}, dev: map[string]bool{}}
	raw, err := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
	if err != nil {
		return d, err
	}
	var py map[string]any
	if err := toml.Unmarshal(raw, &py); err != nil {
		return d, err
	}
	project := table(py, "project")
	if project == nil {
		if table(py, "tool", "poetry") == nil {
			return d, nil
		}
		project = map[string]any{}
	}
	d.known = true
	dynamic, _ := project["dynamic"].([]any)
	for _, f := range dynamic {
		if f == "dependencies" {
			d.known = false
		}
	}
	d.addAll(d.runtime, project["dependencies"])
	for _, extra := range table(project, "optional-dependencies") {
		d.addAll(d.runtime, extra)
	}
	for _, group := range table(py, "dependency-groups") {
		d.addAll(d.dev, group)
	}
	d.addAll(d.dev, table(py, "build-system")["requires"])
	d.addAll(d.dev, table(py, "tool", "uv")["dev-dependencies"])
	d.addAll(d.runtime, table(py, "tool", "poetry", "dependencies"))
	for _, group := range table(py, "tool", "poetry", "group") {
		if g, ok := group.(map[string]any); ok {
			d.addAll(d.dev, g["dependencies"])
		}
	}
	for _, name := range requirementFiles(dir) {
		into := d.dev
		if name == "requirements.txt" {
			into = d.runtime
			d.known = true // the project keeps its dependencies there
		}
		d.addFile(into, filepath.Join(dir, name))
	}
	d.addFile(d.dev, filepath.Join(workspaceRoot, "requirements-dev.txt"))
	return d, nil
}

func requirementFiles(dir string) []string {
	matches, _ := filepath.Glob(filepath.Join(dir, "*requirements*.txt"))
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, filepath.Base(m))
	}
	return names
}

func (d *declarations) addFile(into map[string]bool, file string) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line, _, _ = strings.Cut(line, "#")
		d.add(into, line)
	}
}
