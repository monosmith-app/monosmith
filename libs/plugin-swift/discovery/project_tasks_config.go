package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// inputs are what a project's results depend on: its manifest, its resolved
// dependencies and its Swift sources.
var inputs = []string{"Package.swift", "Package.resolved", "**/*.swift"}

// Tasks are SwiftPM's for one project, run in it: build compiles into
// dist/<project dir> (--scratch-path) and test runs the test targets (in
// .build, which the scaffold ignores). There is no lint: the toolchain's
// swift format is not run, because its verdict depends on a style the team
// has to choose.
func Tasks(root string) map[string]protocol.Task {
	out := protocol.BuildDirFromProject(root)
	return map[string]protocol.Task{
		"build": {Command: []string{"swift", "build", "--scratch-path", out}, Inputs: inputs, Outputs: []string{out}, Cache: true, Description: "swift build, into dist/"},
		"test":  {Command: []string{"swift", "test"}, Inputs: inputs, Cache: true, Description: "swift test"},
	}
}
