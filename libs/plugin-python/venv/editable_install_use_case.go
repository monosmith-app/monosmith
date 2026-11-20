package venv

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// DevLists are the names of the lists in a project's pyproject.toml that hold
// its own development dependencies: the extras of
// [project.optional-dependencies] (smith's scaffolds use "dev") and the PEP 735
// [dependency-groups] of the same names.
var DevLists = []string{"dev", "lint", "test", "testing", "tests", "typing"}

// pyprojectLists is the part of pyproject.toml EditableTarget reads.
type pyprojectLists struct {
	Project struct {
		OptionalDependencies map[string]any `toml:"optional-dependencies"`
	} `toml:"project"`
	DependencyGroups map[string][]any `toml:"dependency-groups"`
}

// EditableTarget is what pip installs for one project: the project in
// editable mode with its development extras (-e ./<root>[dev,test]), then the
// requirements of its development dependency groups. pip installs a group
// itself only from 25.1 on, so smith reads the groups and their
// include-group entries and passes their requirements.
//
// Parameters: workspaceRoot is the workspace root; projectRoot is the
// project's folder from it, with forward slashes.
// Returns: the arguments that follow "pip install"; a project whose
// pyproject.toml cannot be read gets -e ./<root> only.
func EditableTarget(workspaceRoot, projectRoot string) []string {
	target := "./" + projectRoot
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(projectRoot), "pyproject.toml"))
	if err != nil {
		return []string{"-e", target}
	}
	extras, groups := DevRequirements(raw)
	if len(extras) > 0 {
		target += "[" + strings.Join(extras, ",") + "]"
	}
	return append([]string{"-e", target}, groups...)
}

// DevRequirements reads the development lists of a pyproject.toml.
//
// Parameters: pyproject is the file's text.
// Returns: the names of the development extras the project declares, sorted;
// and the requirements of its development dependency groups, include-group
// entries resolved, in order and without repeats. Both are empty when the
// text is not TOML.
func DevRequirements(pyproject []byte) (extras, groups []string) {
	var p pyprojectLists
	if toml.Unmarshal(pyproject, &p) != nil {
		return nil, nil
	}
	seen := map[string]bool{}
	visited := map[string]bool{}
	var walk func(name string)
	walk = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		for _, entry := range p.DependencyGroups[name] {
			switch e := entry.(type) {
			case string:
				if !seen[e] {
					seen[e] = true
					groups = append(groups, e)
				}
			case map[string]any:
				if inc, ok := e["include-group"].(string); ok {
					walk(inc)
				}
			}
		}
	}
	for _, name := range DevLists {
		if _, ok := p.Project.OptionalDependencies[name]; ok {
			extras = append(extras, name)
		}
		walk(name)
	}
	sort.Strings(extras)
	return extras, groups
}
