package taskcache

import (
	"path"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// WorkspacePrefix marks an input glob as workspace-relative.
const WorkspacePrefix = "{workspaceRoot}/"

// SelectInputs returns the files a task reads. Globs are relative to the
// project root unless they start with {workspaceRoot}/; a glob starting with
// "!" excludes. No globs means every file of the project. Files under the
// task's outputs are never inputs, so a build cannot invalidate itself.
func SelectInputs(files []string, projectRoot string, globs, outputs []string) []string {
	if len(globs) == 0 {
		globs = []string{"**/*"}
	}
	var include, exclude []string
	for _, g := range globs {
		negate := strings.HasPrefix(g, "!")
		g = strings.TrimPrefix(g, "!")
		if strings.HasPrefix(g, WorkspacePrefix) {
			g = strings.TrimPrefix(g, WorkspacePrefix)
		} else {
			g = path.Join(projectRoot, g)
		}
		if negate {
			exclude = append(exclude, g)
		} else {
			include = append(include, g)
		}
	}
	for _, o := range outputs {
		exclude = append(exclude, path.Join(projectRoot, o), path.Join(projectRoot, o, "**"))
	}
	var out []string
	for _, f := range files {
		if matchAny(include, f) && !matchAny(exclude, f) {
			out = append(out, f)
		}
	}
	return out
}

// ProjectFiles returns every file under a project root.
func ProjectFiles(files []string, projectRoot string) []string {
	var out []string
	prefix := projectRoot + "/"
	for _, f := range files {
		if strings.HasPrefix(f, prefix) {
			out = append(out, f)
		}
	}
	return out
}

func matchAny(globs []string, file string) bool {
	for _, g := range globs {
		if ok, _ := doublestar.Match(g, file); ok {
			return true
		}
	}
	return false
}
