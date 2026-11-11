package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// DebugVariant is the Android build variant a build ships
// (outputs/apk/debug); `gradle assemble` builds every variant.
const DebugVariant = "debug"

// withShippedArchive makes a build task ship what the module builds into its
// work folder: smith links each archive the globs name into the project's
// shipped folder, which becomes the task's one output. The work folder is
// never an output, so a build keeps Gradle's and Maven's incremental state.
//
// Parameters: build is the task to change; root the module's project folder;
// globs the workspace-relative archives (dist/maven/api/*.jar). It returns the
// changed task.
func withShippedArchive(build protocol.Task, root string, globs ...string) protocol.Task {
	build.Outputs = []string{protocol.BuildDirFromProject(root)}
	build.Ship = nil
	for _, g := range globs {
		build.Ship = append(build.Ship, protocol.FromProject(root, g))
	}
	return build
}

// gradleArchive is what a Gradle module with buildDir ships: the APK of
// variant for an Android module, else the jars of libs/.
func gradleArchive(buildDir string, android bool, variant string) string {
	if android {
		return buildDir + "/outputs/apk/" + variant + "/*.apk"
	}
	return buildDir + "/libs/*.jar"
}
