package discovery

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// gradleInputs are what a Gradle module's results depend on: its sources,
// script and properties, and the workspace's settings, root script,
// properties and version catalog.
var gradleInputs = []string{
	"src/**", "build.gradle.kts", "build.gradle", "gradle.properties",
	"{workspaceRoot}/settings.gradle.kts", "{workspaceRoot}/settings.gradle",
	"{workspaceRoot}/build.gradle.kts", "{workspaceRoot}/build.gradle",
	"{workspaceRoot}/gradle.properties", "{workspaceRoot}/gradle/**",
}

// GradleTasks are Gradle's for one module, run from the workspace root and
// selected by its project path. build assembles it after the modules it
// depends on (^build); test and start wait for it; lint is what the
// module's plugins add (ktlint for Kotlin, Checkstyle for Java, Android
// Lint for an Android module). A module whose build folder smith knows ships
// its jar, or its debug APK, into dist/<root>. An Android module's build has
// a release profile (assembleRelease) that ships the release APK.
func GradleTasks(root string, b GradleBuild, kind protocol.ProjectKind) map[string]protocol.Task {
	gradle := func(tasks ...string) []string {
		argv := []string{"gradle", "--console=plain", "-q"}
		for _, t := range tasks {
			argv = append(argv, b.Path+":"+t)
		}
		return argv
	}
	build := protocol.Task{Command: gradle("assemble"), Cwd: ".", Inputs: gradleInputs, DependsOn: []string{"^build"}, Description: "gradle assemble"}
	if b.Android {
		// --release builds and ships only the release variant.
		build.Release = &protocol.TaskProfile{Command: gradle("assembleRelease")}
	}
	if b.BuildDir != "" {
		build = withShippedArchive(build, root, gradleArchive(b.BuildDir, b.Android, DebugVariant))
		if b.Android {
			build.Release.Ship = []string{protocol.FromProject(root, gradleArchive(b.BuildDir, true, ReleaseVariant))}
		}
	}
	tasks := map[string]protocol.Task{
		"build": build,
		"test":  {Command: gradle("test"), Cwd: ".", Inputs: gradleInputs, DependsOn: []string{"build"}, Cache: true, Description: "gradle test"},
	}
	if len(b.Lint) > 0 {
		// Lint writes its reports into the build directory, which the build
		// fills: it runs after the build. Android's tasks
		// also compile the same intermediates (assemble runs lint-vital, lint
		// compiles the unit tests), so two Gradle builds of one Android
		// module at once overwrite each other's outputs: its tasks run one
		// after another, build, test, lint.
		after := "build"
		if b.Android {
			after = "test"
		}
		tasks["lint"] = protocol.Task{Command: gradle(b.Lint...), Cwd: ".", Inputs: gradleInputs, DependsOn: []string{after}, Cache: true, Description: "gradle " + strings.Join(b.Lint, " ")}
	}
	// An Android app runs on a device, not from Gradle.
	if kind == protocol.ProjectApp && !b.Android {
		tasks["start"] = protocol.Task{Command: gradle("run"), Cwd: ".", DependsOn: []string{"build"}, Description: "Run the app (gradle run)"}
	}
	return tasks
}
