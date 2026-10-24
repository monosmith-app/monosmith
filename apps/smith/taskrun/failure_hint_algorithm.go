package taskrun

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	missingModule  = regexp.MustCompile(`Cannot find module '([^']+)'`)
	commandMissing = regexp.MustCompile(`(?m)(?:^|[\s:'"])([A-Za-z0-9_.@/-]+): (?:command )?not found|'([^']+)' is not recognized as an internal or external command`)
)

// FailureHint reads what a failed task printed and, when the cause is one
// smith can name, says it in a line: a file of the workspace that is missing,
// a package that is not installed, a command that is not on PATH. It returns
// "" when nothing matches, so the task's own output stands alone.
func FailureHint(output, root string) string {
	if m := missingModule.FindStringSubmatch(output); m != nil {
		module := m[1]
		if rel, inside := insideWorkspace(module, root); inside {
			return "it needs " + rel + ", which is missing from the workspace. If it was tracked, restore it with `git checkout -- " + rel + "`; `smith doctor` checks the workspace for other missing files"
		}
		if !strings.HasPrefix(module, ".") && !filepath.IsAbs(module) {
			return "the package " + module + " is not installed: run `smith i`"
		}
	}
	if m := commandMissing.FindStringSubmatch(output); m != nil {
		name := m[1]
		if name == "" {
			name = m[2]
		}
		if name == "smith" {
			return "smith is not on PATH for this task: add monosmith to the root devDependencies (npm i -D monosmith) or install it globally"
		}
		return "the command " + name + " was not found: install it, or on a package.json script run `smith i` so node_modules/.bin has it"
	}
	return ""
}

// insideWorkspace reports whether an absolute module path is under the
// workspace root, and its path relative to it with forward slashes.
func insideWorkspace(module, root string) (string, bool) {
	if !filepath.IsAbs(module) && !regexp.MustCompile(`^[A-Za-z]:[\\/]`).MatchString(module) {
		return "", false
	}
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(module))
	if err != nil || strings.HasPrefix(rel, "..") || strings.Contains(rel, "node_modules") {
		return "", false
	}
	return filepath.ToSlash(rel), true
}
