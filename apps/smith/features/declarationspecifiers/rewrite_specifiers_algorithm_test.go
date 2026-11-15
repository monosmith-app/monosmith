package declarationspecifiers

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewriteGivesASpecifierItsFile(t *testing.T) {
	has := map[string]bool{"/d/hooks.d.ts": true, "/d/lib/index.d.ts": true}
	exists := func(p string) bool { return has[filepath.ToSlash(p)] }
	text := "import { a } from './hooks'\nexport * from \"./lib\"\nimport('./hooks')\nimport x from './done.js'\nimport y from './missing'\nimport z from 'pkg'\n"
	got := Rewrite(text, "/d", exists)
	want := "import { a } from './hooks.js'\nexport * from \"./lib/index.js\"\nimport('./hooks.js')\nimport x from './done.js'\nimport y from './missing'\nimport z from 'pkg'\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRunRewritesOnlyDeclarationFiles(t *testing.T) {
	dir := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("index.d.ts", "export * from './hooks'\n")
	write("hooks.d.ts", "export declare const a: number\n")
	write("notes.txt", "from './hooks'\n")
	changed, err := Run(dir)
	if err != nil || changed != 1 {
		t.Fatalf("changed %d, %v", changed, err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "index.d.ts"))
	if !strings.Contains(string(raw), "'./hooks.js'") {
		t.Fatalf("%s", raw)
	}
	if raw, _ := os.ReadFile(filepath.Join(dir, "notes.txt")); string(raw) != "from './hooks'\n" {
		t.Fatalf("a text file was touched: %s", raw)
	}
}

func TestMigrateScriptReplacesTheLegacyCall(t *testing.T) {
	got, ok := MigrateScript("esbuild src/index.ts && tsc -p tsconfig.lib.json && node ../../declaration-specifiers.mjs dist/src")
	if !ok || got != "esbuild src/index.ts && tsc -p tsconfig.lib.json && smith declaration-specifiers dist/src" {
		t.Fatalf("%v %q", ok, got)
	}
	if got, ok := MigrateScript("tsc -p tsconfig.json"); ok || got != "tsc -p tsconfig.json" {
		t.Fatalf("%v %q", ok, got)
	}
}

// The command replaces a script that ran under Node; both must agree on a
// tree with every kind of specifier.
func TestTheCommandDoesWhatTheLegacyScriptDid(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed")
	}
	files := map[string]string{
		"index.d.ts":         "export { a } from './hook.contract';\nexport type { B } from \"./hooks\";\nexport * from './already.js';\nexport { c } from 'zod';\nexport { d } from './missing';\nexport type E = import('./hooks').E;\n",
		"hook.contract.d.ts": "export declare const a: number;\n",
		"already.d.ts":       "export {};\n",
		"hooks/index.d.ts":   "export type { B } from '../hook.contract';\nexport type E = string;\n",
	}
	build := func(dir string) {
		for name, text := range files {
			file := filepath.Join(dir, "dist", "src", filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	viaScript, viaCommand := t.TempDir(), t.TempDir()
	build(viaScript)
	build(viaCommand)
	if err := os.WriteFile(filepath.Join(viaScript, LegacyFile), []byte(LegacyContent), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(node, LegacyFile, "dist/src")
	cmd.Dir = viaScript
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if _, err := Run(filepath.Join(viaCommand, "dist", "src")); err != nil {
		t.Fatal(err)
	}
	for name := range files {
		want, _ := os.ReadFile(filepath.Join(viaScript, "dist", "src", filepath.FromSlash(name)))
		got, _ := os.ReadFile(filepath.Join(viaCommand, "dist", "src", filepath.FromSlash(name)))
		if string(got) != string(want) {
			t.Errorf("%s:\ngot  %q\nwant %q", name, got, want)
		}
	}
}
