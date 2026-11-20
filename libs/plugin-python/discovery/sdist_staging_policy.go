package discovery

import "regexp"

// sdistStaging matches the `<name>-<version>` folder that `python -m build`
// stages inside the project and removes again.
var sdistStaging = regexp.MustCompile(`^[A-Za-z0-9_.]+-\d+(\.\w+)*$`)

// IsSdistStaging reports whether a folder name is a sdist staging folder: it
// is there for a moment while a build runs, so a walk must not read it.
func IsSdistStaging(name string) bool { return sdistStaging.MatchString(name) }

// StagingGlob matches the sdist staging folder by its name, such as
// monosmith-0.0.0: a dash, then a version with at least one dot. A module
// name has no dash, so no importable file matches it.
const StagingGlob = "*-[0-9]*.[0-9]*"

// stagingSkip returns the arguments that keep ruff and pytest out of the sdist
// staging folder of the project at root. The build of the same project can
// create and remove that folder while they run, and a file that is gone when
// a tool reads it fails the task. ruff's --extend-exclude and pytest's
// --ignore-glob add to the repository's own settings and replace nothing.
//
// Parameters: root is the project folder, workspace-relative ("." for the
// workspace root).
// Returns the ruff arguments and the pytest arguments.
func stagingSkip(root string) (ruff, pytest []string) {
	base := "{projectRoot}"
	if root == "." || root == "" {
		base = "{workspaceRoot}"
	}
	return []string{"--extend-exclude", StagingGlob}, []string{"--ignore-glob", base + "/" + StagingGlob}
}
