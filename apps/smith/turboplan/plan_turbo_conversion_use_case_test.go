package turboplan

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/turboworkspace"
)

// pnpmWorkspace writes a pnpm Turborepo workspace to a temp dir and plans its
// conversion.
func pnpmWorkspace(t *testing.T) *convertplan.Plan {
	t.Helper()
	tree := map[string]string{
		"package.json": `{
  "name": "acme",
  "private": true,
  "packageManager": "pnpm@9.1.0",
  "scripts": {"build": "turbo run build", "dev": "turbo run dev --ui=stream", "lint": "turbo run lint --filter=web...", "test": "pnpm -r run test", "prepare": "npx only-allow pnpm"},
  "devDependencies": {"turbo": "^2.1.0", "typescript": "^5.6.0"},
  "engines": {"node": ">=20", "pnpm": ">=9"},
  "pnpm": {"overrides": {"semver": "7.5.2", "a>b": "1.0.0", "@emnapi/core": "1.4.0"}, "onlyBuiltDependencies": ["esbuild"]}
}`,
		"pnpm-workspace.yaml":          "packages:\n  - 'apps/*'\n  - 'packages/*'\ncatalog:\n  react: ^19.0.0\n",
		"pnpm-lock.yaml":               "lockfileVersion: '9.0'\n",
		".npmrc":                       "shamefully-hoist=true\nregistry=https://registry.npmjs.org/\n",
		".github/workflows/deploy.yml": "steps:\n  - run: pnpm install --frozen-lockfile\n  - run: pnpm turbo run build\n    env:\n      TURBO_TOKEN: x\n",
		"turbo.json": `{
  "$schema": "https://turborepo.com/schema.json",
  "ui": "stream",
  "globalDependencies": ["tsconfig.json"],
  "globalEnv": ["CI"],
  "remoteCache": {"enabled": true},
  "tasks": {
    "build": {"dependsOn": ["^build"], "outputs": ["dist/**", "!dist/cache/**"], "env": ["API_URL"]},
    "web#build": {"dependsOn": ["^build", "@acme/ui#codegen"], "outputs": [".next/**"]},
    "codegen": {"outputs": ["generated/**"]},
    "test": {"dependsOn": ["build"], "inputs": ["$TURBO_DEFAULT$", "!README.md"]},
    "lint": {},
    "dev": {"cache": false, "persistent": true},
    "deploy": {"interactive": true, "cache": false},
    "//#format": {},
    "ghost": {}
  }
}`,
		"apps/web/package.json":      `{"name": "web", "private": true, "scripts": {"build": "next build", "dev": "next dev", "lint": "eslint .", "test": "vitest run"}, "dependencies": {"@acme/ui": "workspace:*", "react": "catalog:"}}`,
		"packages/ui/package.json":   `{"name": "@acme/ui", "version": "1.2.0", "scripts": {"build": "tsc", "lint": "eslint .", "test": "vitest run", "codegen": "node gen.js"}, "peerDependencies": {"react": "catalog:"}}`,
		"packages/ui/src/index.ts":   "export {}\n",
		"packages/core/package.json": `{"name": "@acme/core", "version": "0.4.0", "scripts": {"build": "tsc"}, "dependencies": {"@acme/ui": "workspace:^"}}`,
	}
	root := t.TempDir()
	var files []string
	for rel, text := range tree {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		files = append(files, rel)
	}
	sort.Strings(files)
	w, err := turboworkspace.Read(root, files)
	if err != nil {
		t.Fatal(err)
	}
	return MakePlan(w, convertplan.Input{
		DirName: "acme",
		Files:   files,
		Read: func(rel string) string {
			raw, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			return string(raw)
		},
	})
}

func findings(plan *convertplan.Plan, severity convertplan.Severity) string {
	var out []string
	for _, f := range plan.Findings {
		if f.Severity == severity {
			out = append(out, f.Project+" "+f.Subject+": "+f.Message)
		}
	}
	return strings.Join(out, "\n")
}

func mustContain(t *testing.T, what, text string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Errorf("%s lacks %q:\n%s", what, want, text)
		}
	}
}

func TestMakePlanProjects(t *testing.T) {
	plan := pnpmWorkspace(t)
	if plan.Source != convertplan.SourceTurborepo || plan.Count(convertplan.Blocker) > 0 {
		t.Fatalf("source %q, blockers:\n%s", plan.Source, findings(plan, convertplan.Blocker))
	}
	kinds := map[string]string{}
	for _, p := range plan.Projects {
		kinds[p.Name] = string(p.Kind)
	}
	if want := map[string]string{"web": "app", "ui": "package", "core": "package"}; !reflect.DeepEqual(kinds, want) {
		t.Fatalf("projects %v, want %v", kinds, want)
	}
}

func TestMakePlanTasks(t *testing.T) {
	plan := pnpmWorkspace(t)
	web := plan.Config.Projects["web"].Tasks
	ui := plan.Config.Projects["ui"].Tasks

	if got := web["build"].DependsOn; !reflect.DeepEqual(got, []string{"^build", "ui:codegen"}) {
		t.Errorf("web#build depends on %v", got)
	}
	if got := web["build"].Outputs; !reflect.DeepEqual(got, []string{".next"}) {
		t.Errorf("web#build replaces build: outputs %v", got)
	}
	if ui["build"].DependsOn != nil {
		t.Errorf("^build is the Node plugin's own: %v", ui["build"].DependsOn)
	}
	if got := ui["build"].Outputs; !reflect.DeepEqual(got, []string{"dist"}) {
		t.Errorf("ui build outputs %v", got)
	}
	if got := ui["lint"].DependsOn; got == nil || len(got) != 0 {
		t.Errorf("a lint with no dependsOn must not wait for ^build, as in Turborepo: %#v", got)
	}
	if got := ui["test"].DependsOn; !reflect.DeepEqual(got, []string{"^build", "build"}) {
		t.Errorf("test depends on %v", got)
	}
	if got := ui["test"].Inputs; !reflect.DeepEqual(got, []string{"**/*", "!README.md", "{workspaceRoot}/tsconfig.json", "{workspaceRoot}/package-lock.json"}) {
		t.Errorf("test inputs %v", got)
	}
	if c := web["dev"].Cache; c == nil || *c {
		t.Errorf("a persistent task with cache: false is uncached: %v", c)
	}
	if c := ui["codegen"].Cache; c == nil || !*c {
		t.Errorf("Turborepo caches a task the Node plugin does not: %v", c)
	}
	if got := ui["codegen"].Outputs; !reflect.DeepEqual(got, []string{"generated"}) {
		t.Errorf("codegen outputs %v", got)
	}
	if _, has := ui["deploy"]; has {
		t.Error("no package has a deploy script")
	}
}

func TestMakePlanReportsWhatSmithCannotMap(t *testing.T) {
	plan := pnpmWorkspace(t)
	unsupported := findings(plan, convertplan.Unsupported)
	mustContain(t, "unsupported", unsupported,
		"task build: env [\"API_URL\"]", "globalEnv", "remoteCache", "task deploy: interactive", "pnpm.onlyBuiltDependencies")
	changed := findings(plan, convertplan.Changed)
	mustContain(t, "changed", changed,
		"persistent", "task //#format", "the output exclusion !dist/cache/**", "the lockfile's resolved versions", "kept, but npm ignores the pnpm settings shamefully-hoist")
	dropped := findings(plan, convertplan.Dropped)
	mustContain(t, "dropped", dropped,
		"task ghost", "pnpm@9.1.0", "options that only set how turbo runs: ui", "engines.pnpm", "pnpm-lock.yaml", "npx only-allow pnpm")
	manual := findings(plan, convertplan.Manual)
	mustContain(t, "manual", manual,
		"script lint", "a>b", ".github/workflows/deploy.yml", "the build writes to dist inside the package")
}

func TestMakePlanMovesToNpmWorkspaces(t *testing.T) {
	plan := pnpmWorkspace(t)
	if want := []string{"apps/*", "packages/*"}; !reflect.DeepEqual(plan.SetWorkspaces, want) {
		t.Errorf("workspaces %v", plan.SetWorkspaces)
	}
	if want := map[string]string{"semver": "7.5.2", "@emnapi/core": "$@emnapi/core"}; !reflect.DeepEqual(plan.SetOverrides, want) {
		t.Errorf("overrides %v", plan.SetOverrides)
	}
	if got := plan.AddDevDependencies["@emnapi/core"]; got != "1.4.0" {
		t.Errorf("an override of a lockfile platform package needs the matching devDependency, got %v", plan.AddDevDependencies)
	}
	wantSpecs := map[string]map[string]string{
		"apps/web":      {"@acme/ui": "*", "react": "^19.0.0"},
		"packages/ui":   {"react": "^19.0.0"},
		"packages/core": {"@acme/ui": "^1.2.0"},
	}
	if !reflect.DeepEqual(plan.Specifiers, wantSpecs) {
		t.Errorf("specifiers %v", plan.Specifiers)
	}
	for _, f := range []string{"turbo.json", "pnpm-lock.yaml", "pnpm-workspace.yaml"} {
		if !contains(plan.Delete, f) {
			t.Errorf("%s is removed: %v", f, plan.Delete)
		}
	}
	if want := []string{"packageManager", "engines.pnpm", "pnpm"}; !reflect.DeepEqual(plan.RemoveFields, want) {
		t.Errorf("removed fields %v", plan.RemoveFields)
	}
	if !reflect.DeepEqual(plan.RemoveDependencies, []string{"turbo"}) {
		t.Errorf("removed dependencies %v", plan.RemoveDependencies)
	}
	if got := plan.RootScripts; got["build"] != "smith run build --bail" || got["dev"] != "smith run dev --bail" || got["test"] != "smith run test" || got["prepare"] != "" {
		t.Errorf("root scripts %v", got)
	}
	if _, has := plan.RootScripts["lint"]; has {
		t.Error("a script with a filter smith cannot name is left as it is")
	}
}
