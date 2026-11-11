package discovery

import (
	"path"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// DebugConfiguration is the folder `dotnet build` writes in MSBuild's
// artifacts layout (dist/dotnet/bin/<project>/debug): the configuration,
// lower case.
const DebugConfiguration = "debug"

// ReleaseConfiguration is the folder `dotnet build -c Release` writes
// (dist/dotnet/bin/<project>/release): what `smith build --release` ships.
const ReleaseConfiguration = "release"

// withReleaseConfiguration gives a dotnet task its release profile: the same
// command with -c Release. The cache key holds the command, so each profile
// has its own entry.
//
// Parameters: t is the task to change. It returns the changed task.
func withReleaseConfiguration(t protocol.Task) protocol.Task {
	t.Release = &protocol.TaskProfile{Command: append(slices.Clone(t.Command), "-c", "Release")}
	return t
}

// withShippedBuild makes an app's build task ship what it built: smith links
// every file of dist/dotnet/bin/<project>/<configuration> into the project's
// shipped folder, which becomes the task's one output; the debug folder by
// default, the release folder with --release. The shared artifacts folder
// dist/dotnet is never an output: a cache restore would empty it for every
// project.
//
// Parameters: build is the task to change; root the project folder; file
// its project file (project-relative). It returns the changed task.
func withShippedBuild(build protocol.Task, root, file string) protocol.Task {
	name := strings.TrimSuffix(path.Base(file), path.Ext(file))
	shipped := func(configuration string) []string {
		return []string{protocol.FromProject(root, protocol.BuildDir+"/dotnet/bin/"+name+"/"+configuration+"/*")}
	}
	build.Outputs = []string{protocol.BuildDirFromProject(root)}
	build.Ship = shipped(DebugConfiguration)
	if build.Release == nil {
		build = withReleaseConfiguration(build)
	}
	release := *build.Release
	release.Ship = shipped(ReleaseConfiguration)
	build.Release = &release
	return build
}
