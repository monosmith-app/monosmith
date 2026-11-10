package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// inputs are what a module's results depend on: its sources and pom, and the
// workspace's root pom (plugin versions, Checkstyle rules).
var inputs = []string{"src/**", "pom.xml", "{workspaceRoot}/pom.xml", "{workspaceRoot}/.mvn/**"}

// BuildDir is where a module builds (the root pom's <directory>): the
// workspace's dist/maven/<artifactId>.
func BuildDir(artifactID string) string { return protocol.BuildDir + "/maven/" + artifactID }

// Tasks are Maven's for one module, run from the workspace root and selected
// by its coordinates (-pl groupId:artifactId; a path selector is compared
// as a file, which an 8.3 short temp path on Windows defeats). Each touches
// only its own module: build installs it into the local repository
// after the modules it depends on (^build), so the modules that use it
// resolve it there; test, start and lint wait for the module's build (lint
// is Checkstyle with the root pom's rules, which writes its configuration
// into the build directory). The build ships the module's jar into
// dist/<root>.
func Tasks(root, coordinates, artifactID string, kind protocol.ProjectKind) map[string]protocol.Task {
	mvn := func(args ...string) []string { return append([]string{"mvn", "-B", "-q", "-pl", coordinates}, args...) }
	build := protocol.Task{Command: mvn("install", "-DskipTests"), Cwd: ".", Inputs: inputs, DependsOn: []string{"^build"}, Description: "mvn install (into the local repository, for the modules that use it)"}
	tasks := map[string]protocol.Task{
		"build": withShippedArchive(build, root, BuildDir(artifactID)+"/*.jar"),
		"test":  {Command: mvn("test"), Cwd: ".", Inputs: inputs, DependsOn: []string{"build"}, Cache: true, Description: "mvn test"},
		"lint":  {Command: mvn("checkstyle:check"), Cwd: ".", Inputs: inputs, DependsOn: []string{"build"}, Cache: true, Description: "Checkstyle (the root pom's rules)"},
	}
	if kind == protocol.ProjectApp {
		tasks["start"] = protocol.Task{Command: mvn("exec:java"), Cwd: ".", DependsOn: []string{"build"}, Description: "Run the app (exec:java)"}
	}
	return tasks
}
