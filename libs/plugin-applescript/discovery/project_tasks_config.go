package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// NativeTag marks the project as built on one operating system only:
// osacompile exists on macOS alone, so the generated pipelines leave the
// project to their native job, which [ci] native_runners can restrict to
// macos-latest.
const NativeTag = "native"

// inputs are what a project's result depends on: its scripts.
var inputs = []string{"**/*.applescript"}

// Tasks are osacompile's for one project, run in it: build compiles the main
// script to dist/<project dir>/<name>.scpt. osacompile does not make the
// directory it writes into, so sh does. There is no lint or test: no tool
// checks or runs AppleScript off a Mac.
func Tasks(root, name, main string) map[string]protocol.Task {
	out := protocol.BuildDirFromProject(root)
	return map[string]protocol.Task{
		"build": {
			Command:     []string{"sh", "-c", `mkdir -p "$1" && osacompile -o "$1/$2" "$3"`, "sh", out, name + ".scpt", main},
			Inputs:      inputs,
			Outputs:     []string{out},
			Cache:       true,
			Description: "osacompile: compile " + main + " to a .scpt in dist/ (macOS only)",
		},
	}
}
