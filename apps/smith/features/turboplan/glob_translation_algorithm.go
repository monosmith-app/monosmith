package turboplan

import (
	"path"
	"strings"
)

// rootToken, in an input, stands for the workspace root; defaultToken for the
// files Turborepo hashes when a task lists no inputs.
const (
	rootToken    = "$TURBO_ROOT$"
	defaultToken = "$TURBO_DEFAULT$"
)

// Outputs are a task's output globs as smith states them: directories
// relative to the project, plus what could not be.
type Outputs struct {
	// Dirs are the directories the globs start in ("dist/**" is dist).
	Dirs []string
	// Excluded are the "!" globs: smith caches a directory whole.
	Excluded []string
	// Loose are globs with no directory to cache (`*.tsbuildinfo` at the
	// package root, or a path outside the package).
	Loose []string
}

// OutputDirs turns turbo.json output globs into the directories smith caches.
// A glob keeps the directory it starts in; a nested directory a parent
// already covers is left out.
func OutputDirs(globs []string) Outputs {
	var out Outputs
	var dirs []string
	for _, g := range globs {
		g = strings.TrimPrefix(strings.ReplaceAll(g, "\\", "/"), "./")
		switch {
		case g == "":
			continue
		case strings.HasPrefix(g, "!"):
			out.Excluded = append(out.Excluded, g)
			continue
		case strings.HasPrefix(g, rootToken) || strings.HasPrefix(g, "../"):
			out.Loose = append(out.Loose, g)
			continue
		}
		dir := g
		if i := strings.IndexAny(g, "*?[{"); i >= 0 {
			dir = path.Dir(g[:i] + "x")
		} else if path.Ext(g) != "" {
			dir = path.Dir(g)
		}
		dir = strings.TrimSuffix(dir, "/")
		if dir == "." || dir == "" {
			out.Loose = append(out.Loose, g)
			continue
		}
		dirs = append(dirs, dir)
	}
	for _, d := range dirs {
		covered := false
		for _, other := range dirs {
			if other != d && strings.HasPrefix(d, other+"/") {
				covered = true
			}
		}
		if !covered && !contains(out.Dirs, d) {
			out.Dirs = append(out.Dirs, d)
		}
	}
	return out
}

// InputGlobs turns turbo.json inputs into smith's: $TURBO_DEFAULT$ is every
// file of the project, $TURBO_ROOT$ the workspace root.
func InputGlobs(inputs []string) []string {
	out := make([]string, 0, len(inputs))
	for _, in := range inputs {
		negate := strings.HasPrefix(in, "!")
		in = strings.TrimPrefix(in, "!")
		switch {
		case in == defaultToken && !negate:
			in = "**/*"
		case strings.HasPrefix(in, rootToken+"/"):
			in = "{workspaceRoot}/" + strings.TrimPrefix(in, rootToken+"/")
		}
		if negate {
			in = "!" + in
		}
		out = append(out, in)
	}
	return out
}

// RootGlob is a workspace-relative glob (globalDependencies) as a smith
// input.
func RootGlob(g string) string {
	negate := strings.HasPrefix(g, "!")
	g = strings.TrimPrefix(strings.TrimPrefix(g, "!"), "./")
	g = "{workspaceRoot}/" + g
	if negate {
		return "!" + g
	}
	return g
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
