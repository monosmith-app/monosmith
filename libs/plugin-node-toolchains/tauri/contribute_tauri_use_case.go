package tauri

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/extras"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TaskName is the cached task that builds the app's installers.
const TaskName = "tauri-build"

// DevTask runs the app in development mode; it is never cached.
const DevTask = "tauri-dev"

// SettingsEnv carries the project's settings.tauri table, as JSON, to the
// prepare step, and so into the task's cache key.
const SettingsEnv = "SMITH_TAURI_SETTINGS"

// targetDirFromProject is TargetDir relative to the project.
func targetDirFromProject(root string) string {
	return strings.Repeat("../", strings.Count(strings.Trim(root, "/"), "/")+1) + extras.TargetDir(root)
}

// prepareDirFromProject is PrepareDir relative to the project.
func prepareDirFromProject(root string) string {
	return strings.Repeat("../", strings.Count(strings.Trim(root, "/"), "/")+1) + extras.PrepareDir(root)
}

// ReleasePlatforms are the desktop platforms a Tauri app's release is built
// for by default, one runner each. linux-arm64 is left out: name it in the
// release table's platforms to add it.
var ReleasePlatforms = []string{"windows-x64", "macos-arm64", "macos-x64", "linux-x64"}

// releaseAssetName is what an installer is called on upload: Tauri's own names
// have spaces and say nothing of the platform.
const releaseAssetName = "{product}-{version}-{platform}.{ext}"

// bundleKinds are the installers tauri build writes, each in a folder of its
// own under <target>/release/bundle (msi, nsis, dmg, deb, rpm, appimage), and
// the macOS app folder (bundle/macos/<product>.app).
var bundleKinds = []string{"msi", "exe", "dmg", "deb", "rpm", "AppImage", "app"}

// shippedBundles are the Ship globs of tauri-build, relative to the project:
// every installer of the release profile, which smith links flat into the
// shipped folder dist/<project dir> (ADR 0092).
func shippedBundles(root string) []string {
	globs := make([]string, len(bundleKinds))
	for i, kind := range bundleKinds {
		globs[i] = targetDirFromProject(root) + "/release/bundle/*/*." + kind
	}
	return globs
}

// releaseDefaults is how a Tauri app's release files are made: tauri-build on a
// runner of each platform (and, with sidecars or resources, tauri-verify to
// check the app holds them), then the installers in its shipped folder.
func releaseDefaults(p protocol.Project, withExtras bool) *protocol.ReleaseDefaults {
	tasks := TaskName
	if withExtras {
		tasks += " " + extras.VerifyTask
	}
	return &protocol.ReleaseDefaults{
		Build:     "smith run " + tasks + " -p " + p.Name,
		Assets:    []string{protocol.ProjectBuildDir(p.Root) + "/*.{msi,exe,dmg,deb,rpm,AppImage}"},
		Platforms: ReleasePlatforms,
		AssetName: releaseAssetName,
		Checksums: true,
	}
}

var inputs = []string{"**/*", "!node_modules/**", "!" + tauriapp.SrcDir + "/target/**", "{workspaceRoot}/package-lock.json"}

// Contribute gives every project with a Tauri app a tauri-build task and a
// tauri-dev task. tauri-build keeps cargo's work files in dist/tauri/<project
// dir>, links the installers into the shipped folder dist/<project dir> and
// caches only that folder (ADR 0092). A project whose tauri.conf.json takes the
// web build from dist/tauri/<project dir>/frontend has its build task cache
// that folder. A project whose [projects.<p>.settings.tauri] declares sidecars
// or resources also gets tauri-prepare, which both run first, and
// tauri-verify. Its tauri-build is then not cached: tauri-verify reads the
// work folder, which a cache hit does not restore.
func Contribute(workspaceRoot string, params protocol.ContributeParams) protocol.ContributeResult {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, p := range params.Projects {
		if !tauriapp.Detect(workspaceRoot, p.Root) {
			continue
		}
		deps := []string{"^build"}
		if _, ok := p.Tasks["build"]; ok {
			deps = append(deps, "build")
		}
		// Absolute: cargo runs inside src-tauri and the CLI reads the same flag to find
		// the bundles, so a relative path would depend on which of them resolves it.
		target := "{workspaceRoot}/" + extras.TargetDir(p.Root)
		build := []string{"npx", "--no-install", "tauri", "build"}
		dev := []string{"npx", "--no-install", "tauri", "dev"}
		replace := map[string]protocol.Task{}
		cache, devDeps := true, []string(nil)
		settings := p.Settings[Name]
		if len(settings) > 0 {
			config := "{workspaceRoot}/" + extras.PrepareDir(p.Root) + "/tauri.config.json"
			build, dev = append(build, "--config", config), append(dev, "--config", config)
			deps, devDeps = append(deps, extras.PrepareTask), []string{extras.PrepareTask}
			prepare := prepareTask(p, settings)
			cache = false
			replace[extras.PrepareTask] = prepare
			replace[extras.VerifyTask] = protocol.Task{
				Command:     []string{"{smith}", "tauri-prepare", p.Root, "--verify"},
				Env:         prepare.Env,
				DependsOn:   []string{TaskName},
				Description: "Check that the built app holds the declared sidecars",
			}
		}
		replace[TaskName] = protocol.Task{
			Command:     append(build, "--", "--target-dir", target),
			Inputs:      inputs,
			Outputs:     []string{protocol.BuildDirFromProject(p.Root)},
			Ship:        shippedBundles(p.Root),
			DependsOn:   deps,
			Cache:       cache,
			Description: "Build the Tauri app's installers into " + protocol.ProjectBuildDir(p.Root) + " (cargo work files in " + extras.TargetDir(p.Root) + ")",
		}
		if build, ok := p.Tasks["build"]; ok && frontendDistOf(workspaceRoot, p.Root) == extras.FrontendDir(p.Root) {
			build.Outputs = []string{protocol.FromProject(p.Root, extras.FrontendDir(p.Root))}
			replace["build"] = build
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
		Inputs:      []string{tauriapp.SrcDir + "/tauri.conf.json"},
		Outputs:     []string{prepareDirFromProject(p.Root)},
		Cache:       true,
		Description: "Build the sidecar programs, copy them and the resources for tauri-build (into " + extras.PrepareDir(p.Root) + ")",
	}
	declared, err := extras.ParseExtras(settings)
	if err != nil {
		task.Cache = false
		return task
	}
	for _, s := range declared.Sidecars {
		if s.Project == "" {
			task.Cache = false
			continue
		}
		// The project's folder is not known here, so any Rust change counts.
		task.Inputs = append(task.Inputs, "{workspaceRoot}/**/*.rs", "{workspaceRoot}/**/Cargo.toml", "{workspaceRoot}/**/Cargo.lock", "{workspaceRoot}/.cargo/config.toml")
	}
	for _, r := range declared.Resources {
		from := strings.TrimSuffix(strings.TrimSuffix(r.From, "/*"), "/**")
		if strings.Contains(from, "$") || filepath.IsAbs(from) || strings.HasPrefix(from, "/") || strings.HasPrefix(from, "..") || strings.HasPrefix(from, protocol.BuildDir+"/") {
			task.Cache = false
			continue
		}
		task.Inputs = append(task.Inputs, "{workspaceRoot}/"+from, "{workspaceRoot}/"+from+"/**")
	}
	return task
}
