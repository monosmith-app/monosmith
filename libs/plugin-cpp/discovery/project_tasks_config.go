package discovery

import (
	"runtime"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// inputs are what a project's results depend on: its sources, and those of
// the workspace projects it adds as subdirectories (their sources are
// upstream of it in the graph).
var inputs = []string{"**/*", "!build/**"}

// NativeTag marks a project built on each operating system (set(SMITH_NATIVE
// ON)): smith ci verify leaves it to the native jobs, which also build its
// release archive, each for its own system.
const NativeTag = "native"

// Shape is what a project's tasks depend on.
type Shape struct {
	Root string
	Kind protocol.ProjectKind
	// Executable is the app's target, to start.
	Executable string
	// Lint are the directories cppcheck reads (src, include, tests).
	Lint []string
}

// Tasks are CMake's, CTest's and cppcheck's for one project, run in its
// directory. configure writes the build tree to the workspace's
// dist/<project dir> with Ninja; build and test (CTest) use that tree, so
// CMake's own incremental build applies; lint is cppcheck, failing on what
// it reports.
func Tasks(s Shape) map[string]protocol.Task {
	tree := "{workspaceRoot}/" + protocol.ProjectBuildDir(s.Root)
	tasks := map[string]protocol.Task{
		"configure": {Command: []string{"cmake", "-S", ".", "-B", tree, "-G", "Ninja", "-DCMAKE_BUILD_TYPE=Debug"}, Description: "cmake configure into " + protocol.ProjectBuildDir(s.Root) + "/"},
		"build":     {Command: []string{"cmake", "--build", tree}, DependsOn: []string{"configure"}, Description: "cmake --build"},
		"test":      {Command: []string{"ctest", "--test-dir", tree, "--output-on-failure", "--no-tests=error"}, Inputs: inputs, DependsOn: []string{"build"}, Cache: true, Description: "CTest"},
	}
	if len(s.Lint) > 0 {
		lint := []string{"cppcheck", "--enable=warning,style,performance,portability", "--error-exitcode=1", "--inline-suppr", "--quiet", "--suppress=missingIncludeSystem", "--std=c++20", "-I", "include"}
		tasks["lint"] = protocol.Task{Command: append(lint, s.Lint...), Inputs: inputs, Cache: true, Description: "cppcheck, failing on what it reports"}
	}
	if s.Kind == protocol.ProjectApp && s.Executable != "" {
		exe := tree + "/" + s.Executable
		if runtime.GOOS == "windows" {
			exe += ".exe"
		}
		tasks["start"] = protocol.Task{Command: []string{exe}, DependsOn: []string{"build"}, Description: "Run the app"}
	}
	return tasks
}
