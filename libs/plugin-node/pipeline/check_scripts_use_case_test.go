package pipeline

import (
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

	f, ok := CheckScripts(ws, "", "the root package.json")
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
}

func TestCheckScriptsDoesNotJudgeToolsBeforeInstall(t *testing.T) {
	ws := t.TempDir()
	writeFile(t, ws, "package.json", `{"scripts":{"format:check":"oxfmt --check ."}}`)
	f, ok := CheckScripts(ws, "", "the root package.json")
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
	f, ok := CheckScripts(ws, "packages/a", "a")
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
