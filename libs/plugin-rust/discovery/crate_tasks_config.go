package discovery

import (
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// inputs are what a crate's results depend on: its sources and manifest, and
// the workspace's manifest and lockfile.
var inputs = []string{"**/*", "!target/**", "{workspaceRoot}/Cargo.toml", "{workspaceRoot}/Cargo.lock"}

// formatInputs add the repository's rustfmt configuration to the sources, so
// editing it reruns the check. A configuration inside the crate is already one
// of its files.
var formatInputs = append(append([]string{}, inputs...), "{workspaceRoot}/rustfmt.toml", "{workspaceRoot}/.rustfmt.toml")

// formatTask is cargo fmt --check. fmtArgs are rustfmt's own arguments (smith's
// default configuration) and go after cargo's `--`; none when the repository
// has its own rustfmt.toml.
func formatTask(cargo func(...string) []string, name string, fmtArgs []string) protocol.Task {
	command := cargo("fmt", "-p", name, "--check")
	description := "cargo fmt --check (the repository's rustfmt config)"
	if len(fmtArgs) > 0 {
		command = slices.Concat(command, []string{"--"}, fmtArgs)
		description = "cargo fmt --check (smith's default config)"
	}
	return protocol.Task{Command: command, Inputs: tooldefaults.WithDefaultInputs(formatInputs, fmtArgs), Cache: true, Description: description}
}

// Tasks are cargo's, scoped to one package with -p: build, test, lint
// (clippy, warnings are errors), typecheck (check), format (fmt --check) and
// package (cargo package, what crates.io would get). Cargo builds the
// crates a crate depends on itself, so no task depends on another project's.
func Tasks(name string, kind protocol.ProjectKind, fmtArgs []string) map[string]protocol.Task {
	cargo := func(args ...string) []string { return append([]string{"cargo"}, args...) }
	tasks := map[string]protocol.Task{
		"build":     {Command: cargo("build", "-p", name), Inputs: inputs, Description: "cargo build"},
		"test":      {Command: cargo("test", "-p", name), Inputs: inputs, Cache: true, Description: "cargo test"},
		"lint":      {Command: cargo("clippy", "-p", name, "--all-targets", "--", "-D", "warnings"), Inputs: inputs, Cache: true, Description: "cargo clippy, warnings are errors"},
		"typecheck": {Command: cargo("check", "-p", name, "--all-targets"), Inputs: inputs, Cache: true, Description: "cargo check"},
		"format":    formatTask(cargo, name, fmtArgs),
	}
	if kind == protocol.ProjectApp {
		tasks["start"] = protocol.Task{Command: cargo("run", "-p", name), Description: "Run the app"}
	}
	if kind == protocol.ProjectPackage {
		tasks["package"] = protocol.Task{Command: cargo("package", "-p", name, "--allow-dirty"), Description: "cargo package: the .crate crates.io would get"}
	}
	return tasks
}
