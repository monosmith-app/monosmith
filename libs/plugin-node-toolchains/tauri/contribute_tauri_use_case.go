package tauri

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TaskName is the cached task that builds the app's installers.
const TaskName = "tauri-build"

// DevTask runs the app in development mode; it is never cached.
const DevTask = "tauri-dev"

// SettingsEnv carries the project's settings.tauri table, as JSON, to the
// prepare step, and so into the task's cache key.
const SettingsEnv = "SMITH_TAURI_SETTINGS"

// TargetDir is where a project's Cargo target directory goes, workspace-relative:
// dist/tauri/<project dir>, beside the web build rather than in it.
func TargetDir(root string) string {
	return protocol.BuildDir + "/tauri/" + strings.Trim(root, "/")
}

// targetDirFromProject is TargetDir relative to the project.
func targetDirFromProject(root string) string {
	return strings.Repeat("../", strings.Count(strings.Trim(root, "/"), "/")+1) + TargetDir(root)
}

// prepareDirFromProject is PrepareDir relative to the project.
func prepareDirFromProject(root string) string {
	return strings.Repeat("../", strings.Count(strings.Trim(root, "/"), "/")+1) + PrepareDir(root)
}

// ReleasePlatforms are the desktop platforms a Tauri app's release is built
// for by default, one runner each. linux-arm64 is left out: name it in the
// release table's platforms to add it.
var ReleasePlatforms = []string{"windows-x64", "macos-arm64", "macos-x64", "linux-x64"}

// releaseAssetName is what an installer is called on upload: Tauri's own names
// have spaces and say nothing of the platform.
const releaseAssetName = "{product}-{version}-{platform}.{ext}"

// releaseDefaults is how a Tauri app's release files are made: tauri-build on a
// runner of each platform (and, with sidecars or resources, tauri-verify to
// check the app holds them), then the installers under its bundle folder.
func releaseDefaults(p protocol.Project, extras bool) *protocol.ReleaseDefaults {
	tasks := TaskName
	if extras {
		tasks += " " + VerifyTask
	}
	return &protocol.ReleaseDefaults{
		Build:     "smith run " + tasks + " -p " + p.Name,
		Assets:    []string{TargetDir(p.Root) + "/release/bundle/**/*.{msi,exe,dmg,deb,rpm,AppImage}"},
		Platforms: ReleasePlatforms,
		AssetName: releaseAssetName,
		Checksums: true,
	}
}

var inputs = []string{"**/*", "!node_modules/**", "!" + SrcDir + "/target/**", "{workspaceRoot}/package-lock.json"}

// Contribute gives every project with a Tauri app a tauri-build task, cached
// with its output under dist/, and a tauri-dev task. A project whose
// [projects.<p>.settings.tauri] declares sidecars or resources also gets
// tauri-prepare, which both run first, and tauri-verify.
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
		build := []string{"npx", "--no-install", "tauri", "build"}
		dev := []string{"npx", "--no-install", "tauri", "dev"}
		replace := map[string]protocol.Task{}
		cache, devDeps := true, []string(nil)
		settings := p.Settings[Name]
		if len(settings) > 0 {
			config := "{workspaceRoot}/" + PrepareDir(p.Root) + "/tauri.config.json"
			build, dev = append(build, "--config", config), append(dev, "--config", config)
			deps, devDeps = append(deps, PrepareTask), []string{PrepareTask}
			prepare := prepareTask(p, settings)
			cache = prepare.Cache
			replace[PrepareTask] = prepare
			replace[VerifyTask] = protocol.Task{
				Command:     []string{"{smith}", "tauri-prepare", p.Root, "--verify"},
				Env:         prepare.Env,
				DependsOn:   []string{TaskName},
				Description: "Check that the built app holds the declared sidecars",
			}
		}
		replace[TaskName] = protocol.Task{
			Command:     append(build, "--", "--target-dir", target),
			Inputs:      inputs,
			Outputs:     []string{targetDirFromProject(p.Root)},
			DependsOn:   deps,
			Cache:       cache,
			Description: "Build the Tauri app's installers (cargo target in " + TargetDir(p.Root) + ")",
		}
		replace[DevTask] = protocol.Task{Command: dev, DependsOn: devDeps, Description: "Run the Tauri app in development mode"}
		res.Projects[p.Name] = protocol.Contribution{Release: releaseDefaults(p, len(settings) > 0), Replace: replace}
	}
	return res
}

// prepareTask copies the sidecars and resources. It is cached, and so is the
// build after it, only when everything it reads is in the cache key: a tracked
// file of the workspace, or a Rust project smith builds. A file under dist/,
// outside the workspace or named by an environment variable is not, so then
// both run every time.
func prepareTask(p protocol.Project, settings map[string]any) protocol.Task {
	raw, _ := json.Marshal(settings)
	task := protocol.Task{
		Command:     []string{"{smith}", "tauri-prepare", p.Root},
		Env:         map[string]string{SettingsEnv: string(raw)},
		Inputs:      []string{SrcDir + "/tauri.conf.json"},
		Outputs:     []string{prepareDirFromProject(p.Root)},
		Cache:       true,
		Description: "Build the sidecar programs, copy them and the resources for tauri-build (into " + PrepareDir(p.Root) + ")",
	}
	extras, err := ParseExtras(settings)
	if err != nil {
		task.Cache = false
		return task
	}
	for _, s := range extras.Sidecars {
		if s.Project == "" {
			task.Cache = false
			continue
		}
		// The project's folder is not known here, so any Rust change counts.
		task.Inputs = append(task.Inputs, "{workspaceRoot}/**/*.rs", "{workspaceRoot}/**/Cargo.toml", "{workspaceRoot}/**/Cargo.lock", "{workspaceRoot}/.cargo/config.toml")
	}
	for _, r := range extras.Resources {
		from := strings.TrimSuffix(strings.TrimSuffix(r.From, "/*"), "/**")
		if strings.Contains(from, "$") || filepath.IsAbs(from) || strings.HasPrefix(from, "/") || strings.HasPrefix(from, "..") || strings.HasPrefix(from, protocol.BuildDir+"/") {
			task.Cache = false
			continue
		}
		task.Inputs = append(task.Inputs, "{workspaceRoot}/"+from, "{workspaceRoot}/"+from+"/**")
	}
	return task
}
