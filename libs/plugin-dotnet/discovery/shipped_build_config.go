package discovery

import (
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// DebugConfiguration is the folder `dotnet build` writes in MSBuild's
// artifacts layout (dist/dotnet/bin/<project>/debug): the configuration,
// lower case.
const DebugConfiguration = "debug"

// withShippedBuild makes an app's build task ship what it built: smith links
// every file of dist/dotnet/bin/<project>/<configuration> into the project's
// shipped folder, which becomes the task's one output. The shared artifacts
// folder dist/dotnet is never an output: a cache restore would empty it for
// every project.
//
// Parameters: build is the task to change; root the project folder; file
// its project file (project-relative); configuration the artifacts pivot
// (DebugConfiguration). It returns the changed task.
func withShippedBuild(build protocol.Task, root, file, configuration string) protocol.Task {
	name := strings.TrimSuffix(path.Base(file), path.Ext(file))
	build.Outputs = []string{protocol.BuildDirFromProject(root)}
	build.Ship = []string{protocol.FromProject(root, protocol.BuildDir+"/dotnet/bin/"+name+"/"+configuration+"/*")}
	return build
}
