package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// inputs are what a project's results depend on: its sources (its tests
// project's too) and the workspace's shared MSBuild and SDK settings.
var inputs = []string{"**/*", "!bin/**", "!obj/**", "{workspaceRoot}/Directory.Build.props", "{workspaceRoot}/Directory.Packages.props", "{workspaceRoot}/global.json"}

// Tasks are dotnet's for one project file (file and tests are
// project-relative). MSBuild builds a project's references itself, so two
// builds that share a project must not run at once: build waits for the
// referenced projects' builds (^build), builds the tests project (which
// builds the project), and the tasks that read the build (test, lint, start)
// wait for it and do not build again. lint is dotnet format, which also runs
// the analyzers; a package packs into the workspace's dist/<project dir>, and
// an app's build ships its output there (withShippedBuild). build, test and
// start have a release profile (-c Release): test and start read the build
// of the same configuration.
func Tasks(root, file, tests string, kind protocol.ProjectKind) map[string]protocol.Task {
	target := file
	if tests != "" {
		target = tests
	}
	tasks := map[string]protocol.Task{
		"build": withReleaseConfiguration(protocol.Task{Command: []string{"dotnet", "build", target, "--nologo"}, Inputs: inputs, DependsOn: []string{"^build"}, Description: "dotnet build (the project and its tests)"}),
		"lint":  {Command: []string{"dotnet", "format", file, "--verify-no-changes", "--no-restore"}, Inputs: inputs, DependsOn: []string{"build"}, Cache: true, Description: "dotnet format --verify-no-changes (style and analyzers)",
			Fix: &protocol.TaskFix{Command: []string{"dotnet", "format", file, "--no-restore"}}},
	}
	if tests != "" {
		tasks["test"] = withReleaseConfiguration(protocol.Task{Command: []string{"dotnet", "test", tests, "--nologo", "--no-build"}, Inputs: inputs, DependsOn: []string{"build"}, Cache: true, Description: "dotnet test"})
	}
	switch kind {
	case protocol.ProjectApp:
		tasks["build"] = withShippedBuild(tasks["build"], root, file)
		tasks["start"] = withReleaseConfiguration(protocol.Task{Command: []string{"dotnet", "run", "--project", file, "--no-build"}, DependsOn: []string{"build"}, Description: "Run the app"})
	case protocol.ProjectPackage:
		out := protocol.ProjectBuildDir(root)
		tasks["pack"] = protocol.Task{Command: []string{"dotnet", "pack", file, "--nologo", "-c", "Release", "-o", "{workspaceRoot}/" + out}, Inputs: inputs, DependsOn: []string{"^build"}, Outputs: []string{protocol.BuildDirFromProject(root)}, Cache: true, Description: "dotnet pack into " + out + "/"}
	}
	return tasks
}
