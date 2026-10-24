package pipeline

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckScriptsNamesMissingFilesAndTools(t *testing.T) {
	ws := t.TempDir()
	writeFile(t, ws, "package.json", `{"scripts":{"build":"node ../../declaration-specifiers.mjs","format:check":"oxfmt --check .","lint":"eslint . && echo done","ok":"node ./run.mjs","cfg":"tool -c missing.config.js"}}`)
	writeFile(t, ws, "run.mjs", "")
	writeFile(t, ws, "node_modules/.bin/eslint", "")

	f, ok := CheckScripts(ws, "", "the root package.json", protocol.Project{})
	if !ok || f.OK {
		t.Fatalf("finding %+v", f)
	}
	for _, want := range []string{
		"build runs node ../../declaration-specifiers.mjs, which does not exist",
		"format:check starts oxfmt, which is not installed",
		"cfg runs tool -c missing.config.js, which does not exist",
	} {
		if !strings.Contains(f.Detail, want) {
			t.Errorf("detail lacks %q:\n%s", want, f.Detail)
		}
	}
	for _, unwanted := range []string{"lint ", "ok "} {
		if strings.Contains(f.Detail, "; "+unwanted) || strings.HasPrefix(f.Detail, unwanted) {
			t.Errorf("detail blames %q:\n%s", unwanted, f.Detail)
		}
	}
	if f.Warning {
		t.Error("a broken build script is a failure, not a warning")
	}
	if !strings.Contains(f.Remedy, "oxfmt") || !strings.Contains(f.Remedy, "restore the file") {
		t.Errorf("remedy %q", f.Remedy)
	}
	// The missing tool is a question: install it, or drop the script.
	if f.Fix == nil || f.Fix.Safe || len(f.Fix.Choices) != 1 || len(f.Fix.Choices[0].Options) != 2 {
		t.Fatalf("fix %+v", f.Fix)
	}
	remove := f.Fix.Choices[0].Options[1].Fix.Files[0].Content
	if strings.Contains(remove, "oxfmt") || !strings.Contains(remove, "node ../../declaration-specifiers.mjs") {
		t.Errorf("removing the script keeps the others:\n%s", remove)
	}
}

func TestCheckScriptsDoesNotJudgeToolsBeforeInstall(t *testing.T) {
	ws := t.TempDir()
	writeFile(t, ws, "package.json", `{"scripts":{"format:check":"oxfmt --check ."}}`)
	f, ok := CheckScripts(ws, "", "the root package.json", protocol.Project{})
	if !ok || !f.OK {
		t.Fatalf("no node_modules yet, so no tool can be judged: %+v", f)
	}
}

func TestCheckScriptsPassesWhenEverythingExists(t *testing.T) {
	ws := t.TempDir()
	writeFile(t, ws, "packages/a/package.json", `{"scripts":{"build":"tsc -p tsconfig.json && node scripts/post.mjs","test":"vitest run"}}`)
	writeFile(t, ws, "packages/a/scripts/post.mjs", "")
	writeFile(t, ws, "node_modules/.bin/tsc", "")
	writeFile(t, ws, "node_modules/.bin/vitest", "")
	f, ok := CheckScripts(ws, "packages/a", "a", protocol.Project{})
	if !ok || !f.OK {
		t.Fatalf("%+v", f)
	}
}

func TestCheckConfigImportsFindsAnUninstalledPackage(t *testing.T) {
	ws := t.TempDir()
	writeFile(t, ws, "eslint.config.mjs", "import path from 'node:path'\nimport fs from 'fs'\nimport mnci from '@mnci/eslint-config'\nimport ok from '@acme/present/sub'\nimport local from './local.mjs'\n")
	if _, judged := CheckConfigImports(ws); judged {
		t.Fatal("before install nothing is judged")
	}
	writeFile(t, ws, "node_modules/@acme/present/package.json", "{}")
	f, judged := CheckConfigImports(ws)
	if !judged || f.OK || !strings.Contains(f.Detail, "eslint.config.mjs imports @mnci/eslint-config, which is not installed") || strings.Contains(f.Detail, "present") {
		t.Fatalf("%+v", f)
	}
}

func TestCheckScriptsDoesNotFailAFileTheBuildCreates(t *testing.T) {
	ws := t.TempDir()
	writeFile(t, ws, "packages/app/package.json", `{"scripts":{"build":"esbuild main=src/main.ts --bundle --outdir=dist","dev":"node ./dist/main.esm.js","start":"node ./out/x.js","run":"node ./gone.js"}}`)
	p := protocol.Project{Name: "app", Root: "packages/app", Tasks: map[string]protocol.Task{"build": {Outputs: []string{"../../out"}}}}

	f, ok := CheckScripts(ws, "packages/app", "app", p)
	if !ok || f.OK || f.Warning {
		t.Fatalf("the really missing file must still fail: %+v", f)
	}
	if strings.Contains(f.Detail, "main.esm.js") || !strings.Contains(f.Detail, "gone.js") {
		t.Errorf("detail %q", f.Detail)
	}
	if !strings.Contains(f.Check, "dist/main.esm.js") || !strings.Contains(f.Check, "app:build") {
		t.Errorf("the unbuilt file should be noted: %q", f.Check)
	}

	writeFile(t, ws, "packages/app/package.json", `{"scripts":{"build":"esbuild --outdir=dist","dev":"node ./dist/main.esm.js"}}`)
	f, _ = CheckScripts(ws, "packages/app", "app", p)
	if !f.OK || !strings.Contains(f.Check, "run `smith build` first") {
		t.Errorf("not built yet is a pass with a note: %+v", f)
	}
}
