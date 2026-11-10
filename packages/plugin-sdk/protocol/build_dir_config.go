package protocol

import (
	"path"
	"strings"
)

// BuildDir is the directory at the workspace root every build writes into:
// smith's opinion, so all output is in one place, outside every source tree,
// and cleaning is deleting one folder. A plugin builds a project into
// ProjectBuildDir; smith doctor fails a task whose outputs fall outside it.
const BuildDir = "dist"

// ProjectBuildDir is where a project's build goes, workspace-relative:
// dist/<project dir>, mirroring the workspace (dist/apps/web).
func ProjectBuildDir(root string) string { return BuildDir + "/" + strings.Trim(root, "/") }

// BuildDirFromProject is the same directory relative to the project
// (../../dist/apps/web): what a tool configured from inside the project, or a
// task's outputs, name.
func BuildDirFromProject(root string) string {
	root = strings.Trim(root, "/")
	return strings.Repeat("../", strings.Count(root, "/")+1) + ProjectBuildDir(root)
}

// InBuildDir reports whether a workspace-relative path is inside BuildDir.
func InBuildDir(path string) bool {
	return path == BuildDir || strings.HasPrefix(path, BuildDir+"/")
}

// BuildOutputDir is the folder a project's "build" task writes, workspace-
// relative: its one declared output, or ProjectBuildDir when the task declares
// none or several. A web app built once per environment writes the default
// environment into a folder below ProjectBuildDir (dist/apps/web/prd), and it
// is that folder that a release asset or a package is made from.
func BuildOutputDir(p Project) string {
	if t, ok := p.Tasks["build"]; ok && len(t.Outputs) == 1 {
		return path.Clean(path.Join(p.Root, t.Outputs[0]))
	}
	return ProjectBuildDir(p.Root)
}
