package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// inputs are what a package's results depend on: its sources and the
// workspace's resolution.
var inputs = []string{"**/*", "!.dart_tool/**", "!build/**", "{workspaceRoot}/pubspec.yaml", "{workspaceRoot}/pubspec.lock"}

// Shape is what a package's tasks depend on.
type Shape struct {
	Root, Name string
	Kind       protocol.ProjectKind
	Flutter    bool
	// Tests is true when the package has a test/ directory.
	Tests bool
}

// Tasks are dart's (or flutter's) for one package, run in its directory.
// Flutter commands skip their implicit pub get (--no-pub): the workspace is
// resolved once, by smith install or CI's setup, so parallel tasks never
// rewrite the shared lockfile at once. An app builds into the workspace's
// dist/<package dir>: a Flutter app for the web, a Dart app to a native
// executable bundle.
func Tasks(s Shape) map[string]protocol.Task {
	tool := "dart"
	if s.Flutter {
		tool = "flutter"
	}
	analyze := []string{tool, "analyze", "--fatal-infos"}
	if s.Flutter {
		analyze = append(analyze, "--no-pub")
	}
	tasks := map[string]protocol.Task{
		"lint":   {Command: analyze, Inputs: inputs, Cache: true, Description: tool + " analyze --fatal-infos",
			// dart fix applies the analyzer's fixes, for a Flutter package too.
			Fix: &protocol.TaskFix{Command: []string{"dart", "fix", "--apply"}}},
		"format": {Command: []string{"dart", "format", "--output=none", "--set-exit-if-changed", "."}, Inputs: inputs, Cache: true, Description: "dart format, failing on unformatted files"},
	}
	if s.Tests {
		test := []string{"dart", "test"}
		if s.Flutter {
			test = []string{"flutter", "test", "--no-pub"}
		}
		tasks["test"] = protocol.Task{Command: test, Inputs: inputs, Cache: true, Description: tool + " test"}
	}
	if s.Kind != protocol.ProjectApp {
		return tasks
	}
	out := "{workspaceRoot}/" + protocol.ProjectBuildDir(s.Root)
	if s.Flutter {
		tasks["build"] = protocol.Task{Command: []string{"flutter", "build", "web", "--no-pub", "--no-wasm-dry-run", "-o", out}, Inputs: inputs, Outputs: []string{protocol.BuildDirFromProject(s.Root)}, Cache: true, Description: "flutter build web into " + protocol.ProjectBuildDir(s.Root) + "/"}
		tasks["start"] = protocol.Task{Command: []string{"flutter", "run", "--no-pub", "-d", "web-server"}, Description: "Serve the app (flutter run -d web-server)"}
		return tasks
	}
	tasks["build"] = protocol.Task{Command: []string{"dart", "build", "cli", "-t", "bin/main.dart", "-o", out}, Inputs: inputs, Outputs: []string{protocol.BuildDirFromProject(s.Root)}, Cache: true, Description: "dart build cli into " + protocol.ProjectBuildDir(s.Root) + "/bundle"}
	tasks["start"] = protocol.Task{Command: []string{"dart", "run", "bin/main.dart"}, Description: "Run the app"}
	return tasks
}
