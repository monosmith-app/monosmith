package tauri

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TaskName is the cached task that builds the app's installers.
const TaskName = "tauri-build"

// DevTask runs the app in development mode; it is never cached.
const DevTask = "tauri-dev"

// TargetDir is where a project's Cargo target directory goes, workspace-relative:
// dist/tauri/<project dir>, beside the web build rather than in it.
func TargetDir(root string) string {
	return protocol.BuildDir + "/tauri/" + strings.Trim(root, "/")
}

// targetDirFromProject is TargetDir relative to the project.
func targetDirFromProject(root string) string {
	return strings.Repeat("../", strings.Count(strings.Trim(root, "/"), "/")+1) + TargetDir(root)
}

var inputs = []string{"**/*", "!node_modules/**", "!" + SrcDir + "/target/**", "{workspaceRoot}/package-lock.json"}

// Contribute gives every project with a Tauri app a tauri-build task, cached
// with its output under dist/, and a tauri-dev task.
func Contribute(workspaceRoot string, params protocol.ContributeParams) protocol.ContributeResult {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, p := range params.Projects {
		if !Detect(workspaceRoot, p.Root) {
			continue
		}
		deps := []string{"^build"}
		if _, ok := p.Tasks["build"]; ok {
			deps = append(deps, "build")
		}
		// Cargo runs inside src-tauri, so its relative target dir starts one level deeper.
		target := "../" + targetDirFromProject(p.Root)
		res.Projects[p.Name] = protocol.Contribution{Replace: map[string]protocol.Task{
			TaskName: {
				Command:     []string{"npx", "--no-install", "tauri", "build", "--", "--target-dir", target},
				Inputs:      inputs,
				Outputs:     []string{targetDirFromProject(p.Root)},
				DependsOn:   deps,
				Cache:       true,
				Description: "Build the Tauri app's installers (cargo target in " + TargetDir(p.Root) + ")",
			},
			DevTask: {
				Command:     []string{"npx", "--no-install", "tauri", "dev"},
				Description: "Run the Tauri app in development mode",
			},
		}}
	}
	return res
}
