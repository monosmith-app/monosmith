package commands

import "slices"

// Describe says, for the confirmation dialog, what a command that changes the
// workspace does: a title and one sentence. The dialog adds the exact command line.
func Describe(argv []string) (title, why string) {
	if len(argv) == 0 {
		return "Run a command", "This runs a smith command."
	}
	switch verb := argv[0]; {
	case verb == "release":
		return "Release", "This versions, publishes and tags every package with releasable commits."
	case verb == "doctor" && slices.Contains(argv, "--fix") && slices.Contains(argv, "--only"):
		return "Apply one doctor fix", "This applies the safe fix of the chosen check and rewrites the files it names."
	case verb == "doctor" && slices.Contains(argv, "--fix"):
		return "Fix with doctor", "This applies the safe fixes doctor found and rewrites files."
	case verb == "install":
		return "Install dependencies", "This installs packages with each project's own toolchain."
	case verb == "upgrade":
		return "Upgrade the workspace", "This migrates smith.toml and regenerates the files smith owns."
	case verb == "ci":
		return "Regenerate CI pipelines", "This rewrites the pipeline files smith owns."
	case verb == "add":
		return "Add a project", "This creates the project's files and registers it in the workspace."
	case verb == "remove":
		return "Remove a project", "This deletes the project's directory and its traces in the workspace."
	case verb == "new":
		return "Create a workspace", "This creates a directory with smith.toml, git and CI pipelines."
	case verb == "config":
		return "Change a setting", "This edits smith.toml."
	case verb == "deps":
		if slices.Contains(argv, "--align") {
			return "Align dependency versions", "This rewrites manifests so each dependency has one version across the workspace, then installs."
		}
		return "Update dependencies", "This moves the chosen dependencies to their latest release in every project that declares them, then installs."
	case verb == "plugin":
		return "Change plugins", "This edits smith.toml."
	}
	return "Run smith " + argv[0], "This command changes the workspace or publishes."
}
