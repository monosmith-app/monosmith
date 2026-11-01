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
		return "Release", "This versions every package with releasable commits, publishes it to the registry and pushes its tags. It cannot be undone."
	case verb == "doctor" && slices.Contains(argv, "--fix") && slices.Contains(argv, "--only"):
		return "Apply one doctor fix", "This applies the safe fix of the chosen check and rewrites the files listed below."
	case verb == "doctor" && slices.Contains(argv, "--fix"):
		return "Fix with doctor", "This applies the safe fixes doctor found and rewrites files."
	case verb == "install":
		return "Install everything", "This installs what every project already declares, with its own toolchain (as after a fresh clone). To add one package, use Install a package instead."
	case verb == "upgrade":
		return "Upgrade the workspace", "This updates smith.toml to the current format and rewrites the files smith manages."
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
	case verb == "undo":
		return "Undo the last change", "This reverts the last change smith made with a new commit. Your history stays as it is."
	case verb == "agent" && len(argv) > 1 && argv[1] == "file":
		return "Write the agent file", "This writes AGENTS.md, and a short CLAUDE.md that imports it, from what smith knows about the workspace."
	case verb == "derive" && slices.Contains(argv, "--eject"):
		return "Own a derived file", "This marks the file as yours. smith stops rewriting it."
	case verb == "derive" && slices.Contains(argv, "--adopt"):
		return "Keep your edit of a derived file", "This takes your edit as the new baseline. smith keeps it and rewrites the rest."
	case verb == "derive" && slices.Contains(argv, "--replace"):
		return "Replace a derived file", "This overwrites the file with what smith derives. Your edit is lost."
	case verb == "derive":
		return "Update the derived files", "This rewrites the files that follow from the project graph. A file you edited is never overwritten."
	case verb == "plugin" && len(argv) > 1 && argv[1] == "install":
		return "Install a plugin", "This puts the plugin on this machine after checking it. smith does not start it."
	case verb == "plugin":
		return "Change plugins", "This edits smith.toml."
	}
	return "Run smith " + argv[0], "This command changes the workspace or publishes."
}
