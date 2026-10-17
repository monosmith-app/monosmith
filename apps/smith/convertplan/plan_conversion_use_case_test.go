package convertplan

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func target(t *testing.T, raw string) nxworkspace.Target {
	t.Helper()
	var tg nxworkspace.Target
	if err := json.Unmarshal([]byte(raw), &tg); err != nil {
		t.Fatal(err)
	}
	return tg
}

func project(name, root, typ string, files []string, targets map[string]nxworkspace.Target, tags ...string) *nxworkspace.Project {
	p := &nxworkspace.Project{Name: name, Root: root, Type: typ, Targets: targets, Tags: tags}
	for _, f := range files {
		p.Files = append(p.Files, root+"/"+f)
	}
	return p
}

func findings(plan *Plan, severity Severity) string {
	var out []string
	for _, f := range plan.Findings {
		if f.Severity == severity {
			out = append(out, f.Project+" "+f.Subject+": "+f.Message)
		}
	}
	return strings.Join(out, "\n")
}

// legacyWorkspace looks like what the old mnci generates.
func legacyWorkspace(t *testing.T) (*nxworkspace.Workspace, Input) {
	var nx nxworkspace.NxJSON
	_ = json.Unmarshal([]byte(`{
		"plugins": ["@nx-go/nx-go"],
		"release": {"projects": ["packages/*", "python-packages/*", "tag:release:go", "!tag:type:go-lib"], "releaseTag": {"pattern": "{projectName}@{version}"}},
		"mnci": {"workspaceName": "legacy", "scope": "@legacy", "ci": "both"}
	}`), &nx)
	goLint := target(t, `{"executor": "@nx-go/nx-go:lint", "dependsOn": ["slice-check"], "parallelism": false, "options": {"linter": "golangci-lint", "args": ["run"]}}`)
	slice := target(t, `{"executor": "nx:run-commands", "cache": true, "inputs": ["{projectRoot}/**/*.go", "{workspaceRoot}/tools/go-slice-check.cjs"], "options": {"command": "node tools/go-slice-check.cjs apps/tool"}}`)
	tool := project("tool", "apps/tool", "app", []string{"go.mod", "main.go"}, map[string]nxworkspace.Target{
		"build":       target(t, `{"executor": "@nx-go/nx-go:build"}`),
		"lint":        goLint,
		"slice-check": slice,
		"build-all":   target(t, `{"executor": "nx:run-commands", "options": {"command": "node -e \"build()\""}}`),
		"tidy":        target(t, `{"executor": "@nx-go/nx-go:tidy"}`),
	}, "type:go-app", "release:go")
	tool.Release = map[string]json.RawMessage{"asset": json.RawMessage(`{"name": "{product}_{version}_{os}_{arch}.{ext}", "product": "tl"}`)}
	gokit := project("gokit", "packages/gokit", "lib", []string{"go.mod"}, map[string]nxworkspace.Target{"test": target(t, `{"executor": "@nx-go/nx-go:test"}`)}, "type:go-lib")
	kit := project("@legacy/kit", "packages/kit", "lib", []string{"package.json", "rollup.config.cjs"}, map[string]nxworkspace.Target{
		"build":      target(t, `{"executor": "nx:run-commands", "options": {"cwd": "packages/kit", "command": "rollup -c rollup.config.cjs"}}`),
		"test":       target(t, `{"executor": "nx:run-commands", "dependsOn": ["^build"], "options": {"cwd": "packages/kit", "command": "jest", "env": {"TS_NODE": "1"}, "passWithNoTests": true}}`),
		"typecheck":  target(t, `{"executor": "nx:run-commands", "dependsOn": ["build", "^typecheck"], "options": {"cwd": "packages/kit", "command": "tsc --build tsconfig.json --emitDeclarationOnly"}}`),
		"watch-deps": target(t, `{"executor": "nx:run-commands", "options": {"command": "npx nx watch"}}`),
	})
	kit.Package = &nxworkspace.PackageJSON{Name: "@legacy/kit", Module: "./dist/index.esm.js", Types: "./dist/src/index.d.ts"}
	pykit := project("pykit", "python-packages/pykit", "lib", []string{"pyproject.toml"}, map[string]nxworkspace.Target{
		"typecheck": target(t, `{"executor": "@mnci/nx-python-pip:typecheck"}`),
		"start":     target(t, `{"executor": "nx:run-commands", "options": {"cwd": "python-packages/pykit", "command": "python3 main.py"}}`),
	})
	flutter := project("mobile", "apps/mobile", "app", []string{"pubspec.yaml", "project.json"}, map[string]nxworkspace.Target{})
	flutter.HasProjectJSON = true
	tool.HasProjectJSON = true
	w := &nxworkspace.Workspace{Root: "/ws", Nx: nx, NxVersion: "23.2.0", Projects: []*nxworkspace.Project{tool, gokit, kit, pykit, flutter}}
	w.RootPackage.Scripts = map[string]string{"build": "nx run-many -t build", "tool:qa": "nx run tool:lint", "python:install": "node -e x", "deploy": "nx deploy-it"}
	w.RootPackage.DevDependencies = map[string]string{"nx": "23.2.0", "@nx/js": "23.2.0", "@mnci/cli": "4", "eslint": "10"}
	files := []string{"nx.json", "package.json", "tools/go-app-release.cjs", "packages/kit/rollup.config.cjs", "jest.preset.js", "requirements-dev.txt", "apps/tool/go.mod", ".github/workflows/ci.yml"}
	contents := map[string]string{
		"packages/kit/rollup.config.cjs": "const { withNx } = require('@nx/rollup/with-nx')\nmodule.exports = withNx({ main: './src/index.ts', outputPath: './dist', tsConfig: './tsconfig.lib.json', format: ['esm'] })\n",
		"tools/go-app-release.cjs":       "require('nx/release')",
		"jest.preset.js":                 "const nxPreset = require('@nx/jest/preset').default\nmodule.exports = { ...nxPreset }\n",
		"requirements-dev.txt":           "build\nruff\npytest\npip-audit\ntwine\nmypy\n",
		"apps/tool/go.mod":               "module github.com/acme/legacy/apps/tool\n",
		".github/workflows/ci.yml":       "# mnci:slot after-install\n# mnci:slot-end after-install\n",
	}
	return w, Input{DirName: "legacy", Files: files, Tags: []string{"@legacy/kit@0.1.0"}, Read: func(f string) string { return contents[f] }}
}

func TestPlanLegacyWorkspace(t *testing.T) {
	w, in := legacyWorkspace(t)
	plan := MakePlan(w, in)

	if plan.Config.Workspace.Name != "legacy" || plan.Config.Workspace.Scope != "@legacy" || plan.Config.Workspace.ModulePrefix != "github.com/acme/legacy" {
		t.Fatalf("workspace %+v", plan.Config.Workspace)
	}
	if strings.Join(plan.Config.CI.Providers, ",") != "github,azure" || !plan.Config.Release.GitHubReleases {
		t.Fatalf("ci/release %+v %+v", plan.Config.CI, plan.Config.Release)
	}
	kinds := map[string]protocol.ProjectKind{}
	for _, p := range plan.Projects {
		kinds[p.Name] = p.Kind
	}
	if kinds["tool"] != protocol.ProjectApp || kinds["gokit"] != protocol.ProjectPackage || kinds["kit"] != protocol.ProjectPackage || kinds["pykit"] != protocol.ProjectPackage {
		t.Fatalf("kinds %v", kinds)
	}
	if !plan.Project("pykit").KindMarker || plan.Project("kit").KindMarker {
		t.Fatal("python-packages/ needs a kind marker; packages/ does not")
	}

	tool := plan.Config.Projects["tool"]
	if tool.Release == nil || tool.Release.AssetName != "{product}_{version}_{os}_{arch}.{ext}" || tool.Release.Product != "tl" {
		t.Fatalf("tool release %+v", tool.Release)
	}
	if lint := tool.Tasks["lint"]; strings.Join(lint.Command, " ") != "golangci-lint run --allow-serial-runners" || lint.DependsOn[0] != "slice-check" {
		t.Fatalf("lint %+v", lint)
	}
	slice := tool.Tasks["slice-check"]
	if slice.Cwd != "." || strings.Join(slice.Inputs, ",") != "**/*.go,{workspaceRoot}/tools/go-slice-check.cjs" || slice.Cache == nil || !*slice.Cache {
		t.Fatalf("slice-check %+v", slice)
	}
	if _, kept := tool.Tasks["build-all"]; kept {
		t.Fatal("the old mnci's build-all is replaced by release assets")
	}
	if plan.Config.Projects["kit"].TagPrefix != "@legacy/kit@" {
		t.Fatalf("kit keeps its tags: %+v", plan.Config.Projects["kit"])
	}
	if plan.Config.Projects["gokit"].TagPrefix != "" {
		t.Fatal("Go libraries keep smith's <dir>/v tags")
	}

	kit := plan.Project("kit")
	if !strings.HasPrefix(kit.Scripts["build"], "esbuild src/index.ts --bundle --packages=external --format=esm --outfile=dist/index.esm.js") || !strings.Contains(kit.Scripts["build"], "--declarationDir dist/src --tsBuildInfoFile dist/src/tsconfig.lib.tsbuildinfo && node ../../"+DeclarationSpecifiersFile+" dist/src") || plan.Write[DeclarationSpecifiersFile] != DeclarationSpecifiers {
		t.Fatalf("rollup build %q", kit.Scripts["build"])
	}
	if kit.Scripts["test"] != "jest --passWithNoTests" || plan.Config.Projects["kit"].Tasks["test"].Env["TS_NODE"] != "1" {
		t.Fatalf("test script %q %+v", kit.Scripts["test"], plan.Config.Projects["kit"].Tasks["test"])
	}
	if tc := plan.Config.Projects["kit"].Tasks["typecheck"]; strings.Join(tc.DependsOn, ",") != "build,^typecheck" {
		t.Fatalf("typecheck keeps running after the project's build: %+v", tc)
	}
	if test := plan.Config.Projects["kit"].Tasks["test"]; test.DependsOn != nil {
		t.Fatalf("a bare ^build is the plugin's default: %+v", test)
	}
	if start := plan.Config.Projects["pykit"].Tasks["start"]; strings.Join(start.Command, " ") != "python main.py" {
		t.Fatalf("python3 becomes the venv's python: %v", start.Command)
	}

	if plan.RootScripts["build"] != "smith run build" || plan.RootScripts["tool:qa"] != "" || plan.RootScripts["python:install"] != "" {
		t.Fatalf("root scripts %v", plan.RootScripts)
	}
	if _, ok := plan.RootScripts["deploy"]; ok {
		t.Fatal("an untranslatable script stays as it is")
	}
	if strings.Join(plan.RemoveDependencies, ",") != "@mnci/cli,@nx/js,nx" {
		t.Fatalf("removes %v", plan.RemoveDependencies)
	}
	for _, path := range []string{"nx.json", ".nx", "apps/tool/project.json", "tools/go-app-release.cjs", "packages/kit/rollup.config.cjs"} {
		if !contains(plan.Delete, path) {
			t.Errorf("must delete %s: %v", path, plan.Delete)
		}
	}
	if contains(plan.Delete, "apps/mobile/project.json") {
		t.Error("an unsupported project keeps its project.json")
	}
	if !strings.Contains(plan.Write["jest.preset.js"], "moduleNameMapper") || !strings.Contains(plan.Write["requirements-dev.txt"], "setuptools==") {
		t.Fatalf("writes %v", plan.WriteFiles)
	}

	manual := findings(plan, Manual)
	if !strings.Contains(manual, "script deploy") || strings.Contains(manual, "go-app-release") || strings.Contains(manual, "rollup") {
		t.Fatalf("manual findings:\n%s", manual)
	}
	if !strings.Contains(findings(plan, Unsupported), "mobile apps/mobile: Flutter/Dart project") {
		t.Fatalf("unsupported:\n%s", findings(plan, Unsupported))
	}
	if plan.Blocked(false) || !plan.Blocked(true) {
		t.Fatal("manual work blocks only --strict")
	}
}

func TestPlanKeepsPackageTargetsForSmithCIPack(t *testing.T) {
	w, in := legacyWorkspace(t)
	w.Projects[0].Targets["package"] = target(t, `{"executor": "nx:run-commands", "options": {"command": "node -e \"zip()\""}}`)
	ext := project("@legacy/ext", "apps/ext", "app", []string{"package.json"}, map[string]nxworkspace.Target{
		"package": target(t, `{"executor": "nx:run-commands", "dependsOn": ["build"], "options": {"command": "node packaging/package-vsix.mjs", "cwd": "apps/ext"}}`),
	})
	host := project("@legacy/host", "apps/host", "app", []string{"package.json"}, map[string]nxworkspace.Target{
		"package": target(t, `{"executor": "nx:run-commands", "outputs": ["{workspaceRoot}/dist/drop/host.zip"], "options": {"command": "node -e \"zip()\""}}`),
	})
	w.Projects = append(w.Projects, ext, host)
	plan := MakePlan(w, in)

	if got := plan.Project("ext").Scripts["package"]; got != "node packaging/package-vsix.mjs" {
		t.Fatalf("ext package script %q", got)
	}
	hostPackage := plan.Config.Projects["host"].Tasks["package"]
	if strings.Join(hostPackage.Command, " ") != "node -e zip()" || hostPackage.Cwd != "." || strings.Join(hostPackage.Outputs, ",") != "../../dist/drop" {
		t.Fatalf("host package task %+v", hostPackage)
	}
	if _, kept := plan.Config.Projects["tool"].Tasks["package"]; kept {
		t.Fatal("a Go app's package task is the go plugin's")
	}
	changed := findings(plan, Changed)
	if !strings.Contains(changed, "tool package: the go plugin's package task replaces the old mnci's: smith ci pack puts the app's binary in dist/drop") {
		t.Fatalf("changed:\n%s", changed)
	}
	if all := findings(plan, Dropped); strings.Contains(all, "pack phase") {
		t.Fatalf("dropped:\n%s", all)
	}
}

func TestPlanKeepsDependenciesOnOtherProjectsTargets(t *testing.T) {
	cli := project("@acme/cli", "packages/cli", "lib", []string{"package.json"}, nil)
	ui := project("@acme/studio-ui", "apps/studio-ui", "app", []string{"package.json"}, nil)
	ext := project("@acme/vscode", "apps/vscode-extension", "app", []string{"package.json"}, map[string]nxworkspace.Target{
		// OpenCraw's extension: the string form, as package.json writes it.
		"package": target(t, `{"executor": "nx:run-commands", "dependsOn": ["build", "@acme/cli:build", "@acme/studio-ui:build"], "options": {"command": "node packaging/package-vsix.mjs", "cwd": "apps/vscode-extension"}}`),
		// The object form nx graph normalizes it to, a scope-less name, self,
		// and a pattern smith cannot name.
		"e2e": target(t, `{"executor": "nx:run-commands", "dependsOn": [{"projects": ["cli", "self"], "target": "build"}, {"projects": ["tag:scope:ui"], "target": "build"}], "options": {"command": "node e2e/run.mjs"}}`),
	})
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{cli, ui, ext}}
	plan := MakePlan(w, Input{DirName: "acme", Read: func(string) string { return "" }})

	tasks := plan.Config.Projects["vscode-extension"].Tasks
	if got := strings.Join(tasks["package"].DependsOn, ","); got != "build,cli:build,studio-ui:build" {
		t.Fatalf("package depends on %s", got)
	}
	if got := strings.Join(tasks["e2e"].DependsOn, ","); got != "cli:build,build" {
		t.Fatalf("e2e depends on %s", got)
	}
	if manual := findings(plan, Manual); !strings.Contains(manual, `vscode-extension e2e: Nx ran tag:scope:ui's build first; smith cannot name that project, so add it to depends_on by hand ("<project>:build")`) {
		t.Fatalf("manual:\n%s", manual)
	}
	if mapped := findings(plan, Mapped); !strings.Contains(mapped, "vscode-extension package: runs after cli:build, studio-ui:build, as Nx did (depends_on)") {
		t.Fatalf("mapped:\n%s", mapped)
	}
}

func TestPlanConvertsAnUnbundledEsbuildBuild(t *testing.T) {
	// OpenCraw's azure-host: ESM, bundle: false.
	host := project("@acme/host", "apps/host", "app", []string{"package.json"}, map[string]nxworkspace.Target{
		"build": target(t, `{"executor": "@nx/esbuild:esbuild", "options": {"platform": "node", "outputPath": "apps/host/dist", "format": ["esm"], "bundle": false, "main": "apps/host/src/main.ts", "tsConfig": "apps/host/tsconfig.app.json"}}`),
	})
	api := project("@acme/api", "apps/api", "app", []string{"package.json"}, map[string]nxworkspace.Target{
		"build": target(t, `{"executor": "@nx/esbuild:esbuild", "options": {"outputPath": "apps/api/dist", "bundle": false, "main": "apps/api/src/main.ts", "tsConfig": "apps/api/tsconfig.app.json"}}`),
	})
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{host, api}}
	plan := MakePlan(w, Input{DirName: "acme", Read: func(string) string { return "" }})

	build := plan.Project("host").Scripts["build"]
	if !strings.Contains(build, "getParsedCommandLineOfConfigFile('tsconfig.app.json'") || !strings.Contains(build, "outdir:'dist',platform:'node',format:'esm'") || strings.Contains(build, "--bundle") {
		t.Fatalf("host build %q", build)
	}
	if _, ok := plan.AddDevDependencies["typescript"]; !ok {
		t.Fatalf("dev dependencies %v", plan.AddDevDependencies)
	}
	if changed := findings(plan, Changed); !strings.Contains(changed, "host build: bundle: false: like Nx, it transpiles each file the tsconfig includes") {
		t.Fatalf("changed:\n%s", changed)
	}
	if _, converted := plan.Project("api").Scripts["build"]; converted {
		t.Fatal("a Node CommonJS build with bundle: false needs Nx's loader; it is not converted")
	}
	if manual := findings(plan, Manual); !strings.Contains(manual, "api build: with bundle: false for Node CommonJS, Nx also transpiles the workspace libraries") {
		t.Fatalf("manual:\n%s", manual)
	}
}

func TestPlanBuildsEveryRollupEntry(t *testing.T) {
	// OpenCraw's office-reader: Rollup's input map, one entry per subpath.
	reader := project("@acme/reader", "packages/reader", "lib", []string{"package.json", "rollup.config.cjs"}, map[string]nxworkspace.Target{})
	reader.Package = &nxworkspace.PackageJSON{Name: "@acme/reader", Module: "./dist/index.esm.js", Types: "./dist/src/index.d.ts"}
	dynamic := project("@acme/dyn", "packages/dyn", "lib", []string{"package.json", "rollup.config.cjs"}, map[string]nxworkspace.Target{})
	dynamic.Package = &nxworkspace.PackageJSON{Name: "@acme/dyn", Module: "./dist/index.esm.js", Types: "./dist/src/index.d.ts"}
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{reader, dynamic}}
	contents := map[string]string{
		"packages/reader/rollup.config.cjs": officeReaderRollup,
		"packages/dyn/rollup.config.cjs":    "const { withNx } = require('@nx/rollup/with-nx')\nmodule.exports = withNx({ main: './src/index.ts', tsConfig: './tsconfig.lib.json', format: ['esm'] }, { input: entries() })\n",
	}
	plan := MakePlan(w, Input{DirName: "acme", Files: []string{"packages/reader/rollup.config.cjs", "packages/dyn/rollup.config.cjs"}, Read: func(f string) string { return contents[f] }})

	build := plan.Project("reader").Scripts["build"]
	if !strings.HasPrefix(build, "esbuild index=src/index.ts xlsx=src/spreadsheet/index.ts pptx=src/presentation/index.ts docx=src/document/index.ts --bundle --splitting --packages=external --format=esm --outdir=dist --out-extension:.js=.esm.js") || !strings.Contains(build, "&& tsc -p tsconfig.lib.json --emitDeclarationOnly") {
		t.Fatalf("reader build %q", build)
	}
	if changed := findings(plan, Changed); !strings.Contains(changed, "every entry Rollup built (index, xlsx, pptx, docx, sharing code in dist/chunks)") {
		t.Fatalf("changed:\n%s", changed)
	}
	if _, converted := plan.Project("dyn").Scripts["build"]; converted {
		t.Fatal("an input convert cannot read must not build main alone")
	}
	if manual := findings(plan, Manual); !strings.Contains(manual, "packages/dyn/rollup.config.cjs: builds several entries (input or additionalEntryPoints) convert cannot read") {
		t.Fatalf("manual:\n%s", manual)
	}
}

func TestPlanReportsEsbuildOptionsItCannotCarry(t *testing.T) {
	ext := project("@acme/ext", "apps/ext", "app", []string{"package.json"}, map[string]nxworkspace.Target{
		"build": target(t, `{"executor": "@nx/esbuild:esbuild", "options": {"main": "apps/ext/src/extension.ts", "outputPath": "apps/ext/dist", "esbuildOptions": {"outExtension": {".js": ".cjs"}, "banner": {"js": "// x"}}}}`),
	})
	plan := MakePlan(&nxworkspace.Workspace{Projects: []*nxworkspace.Project{ext}}, Input{DirName: "acme", Read: func(string) string { return "" }})
	if build := plan.Project("ext").Scripts["build"]; !strings.Contains(build, "--outdir=dist --out-extension:.js=.cjs --sourcemap") {
		t.Fatalf("build %q", build)
	}
	if manual := findings(plan, Manual); !strings.Contains(manual, "ext build: esbuildOptions not carried to the build script: banner") {
		t.Fatalf("manual:\n%s", manual)
	}
}

func TestPlanShipsANodeAppWithADockerfileAsAnImage(t *testing.T) {
	host := project("@acme/host", "apps/host", "app", []string{"package.json", "Dockerfile"}, nil)
	lib := project("@acme/tools", "libs/tools", "lib", []string{"package.json", "Dockerfile"}, nil)
	bare := project("image", "apps/image", "app", []string{"Dockerfile"}, nil)
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{host, lib, bare}}
	plan := MakePlan(w, Input{DirName: "acme", Read: func(string) string { return "" }})

	if pp := plan.Project("host"); pp.Language != "node" || pp.Kind != protocol.ProjectApp {
		t.Fatalf("host %+v", pp)
	}
	if _, on := plan.Config.Conventions["docker"]; !on {
		t.Fatalf("conventions %+v", plan.Config.Conventions)
	}
	changed := findings(plan, Changed)
	if !strings.Contains(changed, "host Dockerfile: [conventions.docker]") || !strings.Contains(changed, "tools Dockerfile: kept, but the docker convention builds images of apps only") {
		t.Fatalf("changed:\n%s", changed)
	}
	if unsupported := findings(plan, Unsupported); !strings.Contains(unsupported, "image apps/image: a container image project") || strings.Contains(unsupported, "host") {
		t.Fatalf("a Dockerfile with no manifest stays unsupported:\n%s", unsupported)
	}
}

func TestPlanBlocksDuplicateDirectoryNames(t *testing.T) {
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{
		project("a-ui", "apps/ui", "app", []string{"package.json"}, nil),
		project("l-ui", "libs/ui", "lib", []string{"package.json"}, nil),
	}}
	plan := MakePlan(w, Input{DirName: "x", Read: func(string) string { return "" }})
	if !plan.Blocked(false) || !strings.Contains(findings(plan, Blocker), "would both be called \"ui\"") {
		t.Fatalf("%v", plan.Findings)
	}
}

func TestPlanPlainNxReleasesOnlyPublicPackages(t *testing.T) {
	pub := project("@org/pub", "libs/pub", "lib", []string{"package.json"}, nil)
	pub.Package = &nxworkspace.PackageJSON{Name: "@org/pub"}
	priv := project("@org/priv", "libs/priv", "lib", []string{"package.json"}, nil)
	priv.Package = &nxworkspace.PackageJSON{Name: "@org/priv", Private: true}
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{pub, priv}}
	w.RootPackage.Name = "@org/source"
	plan := MakePlan(w, Input{DirName: "plain", Read: func(string) string { return "" }})
	if plan.Project("pub").Kind != protocol.ProjectPackage || !plan.Project("pub").KindMarker || plan.Project("priv").Kind != protocol.ProjectLib {
		t.Fatalf("%+v %+v", plan.Project("pub"), plan.Project("priv"))
	}
	if plan.Config.Workspace.Name != "plain" || plan.Config.Workspace.Scope != "@org" {
		t.Fatalf("%+v", plan.Config.Workspace)
	}
}

func TestPlanLeavesStandaloneNodeFoldersOutOfTheWorkspace(t *testing.T) {
	node := func(name, root string, projectJSON bool) *nxworkspace.Project {
		p := project(name, root, "app", []string{"package.json"}, nil)
		p.Package = &nxworkspace.PackageJSON{Name: name, Private: true}
		p.HasProjectJSON = projectJSON
		return p
	}
	web := node("@acme/web", "apps/web", false)
	movies := node("example-movies", "examples/movies", false)
	shop := node("example-shop", "examples/shop", false)
	// Declared to Nx with a project.json: kept.
	gen := node("gen", "tools/gen", true)
	// Standalone beside a kept project: excluded by its own path.
	scratch := node("scratch", "tools/scratch", false)
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{web, movies, shop, gen, scratch}}
	w.RootPackage.Workspaces = json.RawMessage(`["packages/*", "apps/*"]`)
	plan := MakePlan(w, Input{DirName: "acme", Read: func(string) string { return "" }})

	if got := strings.Join(plan.Config.Workspace.Exclude, ","); got != "examples/*,tools/scratch" {
		t.Fatalf("exclude = %s", got)
	}
	for _, name := range []string{"movies", "shop", "scratch"} {
		if plan.Project(name) != nil {
			t.Errorf("%s is still a project", name)
		}
	}
	if plan.Project("web") == nil || plan.Project("gen") == nil {
		t.Error("a workspace member or a project.json project was left out")
	}
	if mapped := findings(plan, Mapped); !strings.Contains(mapped, "movies examples/movies: outside the root package.json \"workspaces\"") {
		t.Fatalf("mapped:\n%s", mapped)
	}
}
