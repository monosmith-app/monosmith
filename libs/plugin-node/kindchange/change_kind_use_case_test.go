package kindchange

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func edited(r protocol.ChangeKindResult) map[string]string {
	out := map[string]string{}
	for _, f := range r.Files {
		out[f.Path] = f.Content
	}
	return out
}

func workspace() map[string]string {
	return map[string]string{
		"package.json":       `{"name":"root","private":true,"workspaces":["libs/*","apps/*"]}`,
		"tsconfig.base.json": `{"compilerOptions":{"paths":{"@acme/core":["libs/core/src/index.ts"]}}}`,
		"package-lock.json":  `{"packages":{"libs/core":{"version":"1.0.0"},"node_modules/@acme/core":{"resolved":"libs/core","link":true}}}`,
		"libs/core/package.json": `{
  "name": "@acme/core",
  "version": "1.0.0",
  "private": true,
  "smith": { "kind": "lib" },
  "main": "../../dist/libs/core/index.js",
  "scripts": { "build": "tsc --outDir ../../dist/libs/core" },
  "dependencies": { "@acme/util": "file:../util" }
}
`,
		"libs/core/tsconfig.json": `{"extends":"../../tsconfig.base.json","compilerOptions":{"outDir":"../../dist/libs/core"},"references":[{"path":"../util"}]}`,
		"libs/util/package.json":  `{"name":"@acme/util","version":"1.0.0","private":true,"smith":{"kind":"lib"}}`,
		"apps/web/package.json":   `{"name":"web","private":true,"smith":{"kind":"app"},"dependencies":{"@acme/core":"file:../../libs/core"},"scripts":{"start":"node ."}}`,
		"apps/web/tsconfig.json":  `{"references":[{"path":"../../libs/core"}]}`,
	}
}

func projects() []protocol.Project {
	return []protocol.Project{
		{Name: "core", Root: "libs/core", Language: "node"},
		{Name: "util", Root: "libs/util", Language: "node"},
		{Name: "web", Root: "apps/web", Language: "node"},
	}
}

func TestLibBecomesPackage(t *testing.T) {
	dir := writeFiles(t, workspace())
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: projects()[0], To: protocol.ProjectPackage, NewRoot: "packages/core", Projects: projects(),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	core := got["libs/core/package.json"]
	for _, want := range []string{`"kind": "package"`, `"main": "../../dist/packages/core/index.js"`, `--outDir ../../dist/packages/core`, `"file:../../libs/util"`} {
		if !strings.Contains(core, want) {
			t.Errorf("package.json lacks %s:\n%s", want, core)
		}
	}
	if strings.Contains(core, "private") {
		t.Errorf("a package is not private:\n%s", core)
	}
	ts := got["libs/core/tsconfig.json"]
	if !strings.Contains(ts, `"../../dist/packages/core"`) || !strings.Contains(ts, `"path":"../../libs/util"`) || !strings.Contains(ts, `"../../tsconfig.base.json"`) {
		t.Errorf("tsconfig follows the move:\n%s", ts)
	}
	if !strings.Contains(got["apps/web/package.json"], `file:../../packages/core`) || !strings.Contains(got["apps/web/tsconfig.json"], `../../packages/core`) {
		t.Errorf("dependents follow: %v", got)
	}
	if !strings.Contains(got["tsconfig.base.json"], `packages/core/src/index.ts`) {
		t.Errorf("tsconfig paths follow:\n%s", got["tsconfig.base.json"])
	}
	if !strings.Contains(got["package-lock.json"], `"packages/core"`) || !strings.Contains(got["package-lock.json"], `"resolved":"packages/core"`) {
		t.Errorf("the lockfile follows:\n%s", got["package-lock.json"])
	}
	if !strings.Contains(got["package.json"], `"packages/*"`) {
		t.Errorf("workspaces cover the new directory:\n%s", got["package.json"])
	}
	if _, touched := got["libs/util/package.json"]; touched {
		t.Error("a project that does not use core is not edited")
	}
	if !strings.Contains(strings.Join(r.Warnings, "\n"), "smith install") {
		t.Errorf("the reinstall is named: %v", r.Warnings)
	}
}

func TestPackageBecomesLib(t *testing.T) {
	files := map[string]string{
		"package.json":              `{"name":"root","private":true,"workspaces":["packages/*","libs/*"]}`,
		"packages/sdk/package.json": `{"name":"@acme/sdk","version":"1.0.0","license":"MIT","description":"x","smith":{"kind":"package"},"publishConfig":{"access":"public"}}`,
	}
	dir := writeFiles(t, files)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "sdk", Root: "packages/sdk", Language: "node"},
		To:      protocol.ProjectLib, NewRoot: "libs/sdk",
		Projects: []protocol.Project{{Name: "sdk", Root: "packages/sdk", Language: "node"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)["packages/sdk/package.json"]
	if !strings.Contains(got, `"private": true`) || !strings.Contains(got, `"kind": "lib"`) {
		t.Fatalf("private again, kind lib:\n%s", got)
	}
	if _, touched := edited(r)["package.json"]; touched {
		t.Fatal("libs/* already covers the new directory")
	}
}

func TestAppNeedsSomethingToRun(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"package.json":           `{"name":"root","workspaces":["libs/*"]}`,
		"libs/core/package.json": `{"name":"core","private":true}`,
	})
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "core", Root: "libs/core", Language: "node"},
		To:      protocol.ProjectApp, NewRoot: "apps/core",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(r.Warnings, "\n"), "no start script") {
		t.Fatalf("an app with nothing to run is named: %v", r.Warnings)
	}
}

func TestLibToPackageAndBackLeavesTheManifestByteForByte(t *testing.T) {
	original := "{\n  \"name\": \"@acme/x\",\n  \"version\": \"0.0.0\",\n  \"private\": true,\n  \"smith\": { \"kind\": \"lib\" },\n  \"main\": \"src/index.js\"\n}\n"
	dir := writeFiles(t, map[string]string{"libs/x/package.json": original})
	project := protocol.Project{Name: "x", Root: "libs/x", Language: "node"}
	for _, to := range []protocol.ProjectKind{protocol.ProjectPackage, protocol.ProjectLib} {
		r, err := Change(dir, protocol.ChangeKindParams{Project: project, To: to, NewRoot: "libs/x"})
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range r.Files {
			if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(e.Path)), []byte(e.Content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if to == protocol.ProjectPackage {
			if got, _ := os.ReadFile(filepath.Join(dir, "libs/x/package.json")); strings.Contains(string(got), "private") || !strings.Contains(string(got), `"kind": "package"`) {
				t.Fatalf("a package:\n%s", got)
			}
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "libs/x/package.json")); string(got) != original {
		t.Fatalf("round trip changed the manifest:\n%s", got)
	}
}
