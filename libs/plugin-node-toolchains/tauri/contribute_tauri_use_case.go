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

// ReleasePlatforms are the desktop platforms a Tauri app's release is built
// for by default, one runner each. linux-arm64 is left out: name it in the
// release table's platforms to add it.
var ReleasePlatforms = []string{"windows-x64", "macos-arm64", "macos-x64", "linux-x64"}

// releaseAssetName is what an installer is called on upload: Tauri's own names
// have spaces and say nothing of the platform.
const releaseAssetName = "{product}-{version}-{platform}.{ext}"

// releaseDefaults is how a Tauri app's release files are made: tauri-build on a
// runner of each platform, then the installers under its bundle folder.
func releaseDefaults(p protocol.Project) *protocol.ReleaseDefaults {
	return &protocol.ReleaseDefaults{
		Build:     "smith run " + TaskName + " -p " + p.Name,
		Assets:    []string{TargetDir(p.Root) + "/release/bundle/**/*.{msi,exe,dmg,deb,rpm,AppImage}"},
		Platforms: ReleasePlatforms,
		AssetName: releaseAssetName,
		Checksums: true,
	}
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
		// Absolute: cargo runs inside src-tauri and the CLI reads the same flag to find
		// the bundles, so a relative path would depend on which of them resolves it.
		target := "{workspaceRoot}/" + TargetDir(p.Root)
		res.Projects[p.Name] = protocol.Contribution{Release: releaseDefaults(p), Replace: map[string]protocol.Task{
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
