package adoptsplit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const orders = `{
  "name": "orders",
  "scripts": { "build": "tsc -p tsconfig.json", "start": "node dist/index.js" },
  "main": "dist/index.js",
  "dependencies": { "zod": "^3.0.0" },
  "devDependencies": { "typescript": "^5.0.0", "jest": "^29.0.0", "eslint": "^8.0.0", "supertest": "^7.0.0" }
}
`

const ordersTSConfig = `{
  // compiler options
  "compilerOptions": { "target": "ES2022", "strict": true, "outDir": "dist", "rootDir": "src", "paths": { "@/*": ["src/*"] } },
  "include": ["src"]
}
`

func put(t *testing.T, root, rel, text string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readJSON(t *testing.T, root, rel string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("%s: %v\n%s", rel, err, raw)
	}
	return m
}

func keys(m any) []string {
	var out []string
	for k := range m.(map[string]any) {
		out = append(out, k)
	}
	return out
}

func TestSortNodeSeparatesToolsLintAndUnknown(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", orders)
	got, err := SortNode(root)
	if err != nil {
		t.Fatal(err)
	}
	want := Sorted{Tools: []string{"jest", "typescript"}, Lint: true, Replaced: []string{"eslint"}, Unsure: []string{"supertest"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortNode = %+v, want %+v", got, want)
	}
}

func TestSplitNodeCutsTheProjectFromTheWorkspace(t *testing.T) {
	root := t.TempDir()
	put(t, root, "apps/orders/package.json", orders)
	put(t, root, "apps/orders/tsconfig.json", ordersTSConfig)
	put(t, root, "apps/orders/jest.config.js", "module.exports = {}\n")
	put(t, root, ".eslintrc.json", "{}\n")
	put(t, root, ".prettierrc", "{}\n")

	res, err := SplitNode(Request{
		Root: root, Dir: "apps/orders", Workspace: "shop", Patterns: []string{"apps/*"},
		Hoist: []string{"jest", "typescript", "supertest"}, Drop: []string{"eslint"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.ESLint {
		t.Error("the project used ESLint, so the shared lint config is wanted")
	}

	rootPkg := readJSON(t, root, "package.json")
	if rootPkg["name"] != "shop-workspace" || rootPkg["private"] != true {
		t.Errorf("root manifest = %v", rootPkg)
	}
	if !reflect.DeepEqual(rootPkg["workspaces"], []any{"apps/*"}) {
		t.Errorf("workspaces = %v", rootPkg["workspaces"])
	}
	dev := rootPkg["devDependencies"].(map[string]any)
	if len(dev) != 3 || dev["jest"] != "^29.0.0" || dev["supertest"] != "^7.0.0" || dev["typescript"] != "^5.0.0" {
		t.Errorf("root devDependencies = %v", dev)
	}

	pkg := readJSON(t, root, "apps/orders/package.json")
	if _, has := pkg["devDependencies"]; has {
		t.Errorf("the project keeps devDependencies: %v", pkg["devDependencies"])
	}
	if got := keys(pkg["dependencies"]); !reflect.DeepEqual(got, []string{"zod"}) {
		t.Errorf("dependencies = %v: the ones the code runs on stay", got)
	}
	if pkg["main"] != "../../dist/apps/orders/index.js" {
		t.Errorf("main = %v: the build moves into the workspace dist", pkg["main"])
	}
	scripts := pkg["scripts"].(map[string]any)
	if !strings.Contains(scripts["build"].(string), "--outDir ../../dist/apps/orders") || scripts["start"] != "node ../../dist/apps/orders/index.js" {
		t.Errorf("scripts = %v", scripts)
	}

	base := readJSON(t, root, "tsconfig.base.json")["compilerOptions"].(map[string]any)
	if !reflect.DeepEqual(base, map[string]any{"target": "ES2022", "strict": true}) {
		t.Errorf("tsconfig.base.json compilerOptions = %v", base)
	}
	ts := readJSON(t, root, "apps/orders/tsconfig.json")
	if ts["extends"] != "../../tsconfig.base.json" {
		t.Errorf("extends = %v", ts["extends"])
	}
	opts := ts["compilerOptions"].(map[string]any)
	if _, has := opts["paths"]; !has || opts["rootDir"] != "src" || opts["target"] != nil {
		t.Errorf("project compilerOptions = %v: paths and rootDir stay, shared options go", opts)
	}
	if !reflect.DeepEqual(ts["include"], []any{"src"}) {
		t.Errorf("include = %v", ts["include"])
	}

	if _, err := os.Stat(filepath.Join(root, ".eslintrc.json.pre-smith")); err != nil {
		t.Errorf("the legacy ESLint config is kept as .pre-smith: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".prettierrc")); err != nil {
		t.Errorf("the prettier config stays at the root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "apps/orders/jest.config.js")); err != nil {
		t.Errorf("jest.config.js stays with the project: %v", err)
	}
	notes := strings.Join(res.Notes, "\n")
	for _, want := range []string{"moved 3 devDependencies", "removed from the project", "tsconfig.base.json", "jest.config.js stays", ".eslintrc.json is kept"} {
		if !strings.Contains(notes, want) {
			t.Errorf("notes lack %q:\n%s", want, notes)
		}
	}
}

func TestSplitNodeLeavesATSConfigItCannotShareWhole(t *testing.T) {
	cases := map[string]string{
		"already extends": `{ "extends": "./other.json", "compilerOptions": { "strict": true } }`,
		"broken json":     `{ "compilerOptions": { "strict": true, } `,
	}
	for name, text := range cases {
		root := t.TempDir()
		put(t, root, "apps/orders/package.json", orders)
		put(t, root, "apps/orders/tsconfig.json", text)
		res, err := SplitNode(Request{Root: root, Dir: "apps/orders", Workspace: "shop", Patterns: []string{"apps/*"}})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		raw, _ := os.ReadFile(filepath.Join(root, "apps/orders/tsconfig.json"))
		if string(raw) != text {
			t.Errorf("%s: tsconfig.json changed: %s", name, raw)
		}
		if _, err := os.Stat(filepath.Join(root, "tsconfig.base.json")); err == nil {
			t.Errorf("%s: a base was written", name)
		}
		if !strings.Contains(strings.Join(res.Notes, "\n"), "tsconfig.json stays whole") {
			t.Errorf("%s: the person is not told: %v", name, res.Notes)
		}
	}
}

func TestSplitNodeKeepsAnExistingRootManifest(t *testing.T) {
	root := t.TempDir()
	put(t, root, "apps/orders/package.json", orders)
	put(t, root, "package.json", `{ "name": "mine", "workspaces": ["apps/orders"], "devDependencies": { "jest": "^30.0.0" } }`)
	if _, err := SplitNode(Request{Root: root, Dir: "apps/orders", Workspace: "shop", Patterns: []string{"apps/*"}, Hoist: []string{"jest", "typescript"}}); err != nil {
		t.Fatal(err)
	}
	rootPkg := readJSON(t, root, "package.json")
	dev := rootPkg["devDependencies"].(map[string]any)
	if rootPkg["name"] != "mine" || !reflect.DeepEqual(rootPkg["workspaces"], []any{"apps/orders"}) || dev["jest"] != "^30.0.0" || dev["typescript"] != "^5.0.0" {
		t.Errorf("root manifest = %v", rootPkg)
	}
}

func TestSplitNodeKeepsTheProjectsOwnESLintWhenAsked(t *testing.T) {
	root := t.TempDir()
	put(t, root, "apps/orders/package.json", orders)
	put(t, root, ".eslintrc.json", "{}\n")
	put(t, root, "eslint.config.mjs", "export default []\n")
	put(t, root, ".eslintignore", "dist\n")
	res, err := SplitNode(Request{Root: root, Dir: "apps/orders", Workspace: "shop", Patterns: []string{"apps/*"}, Hoist: []string{"jest"}, KeepLint: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.ESLint {
		t.Error("the shared lint config is not wanted when the person keeps their own")
	}
	for _, name := range []string{".eslintrc.json", "eslint.config.mjs", ".eslintignore"} {
		if _, err := os.Stat(filepath.Join(root, "apps/orders", name)); err != nil {
			t.Errorf("%s must move into the project: %v", name, err)
		}
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			t.Errorf("%s must leave the root", name)
		}
	}
	pkg := readJSON(t, root, "apps/orders/package.json")
	if dev := pkg["devDependencies"].(map[string]any); dev["eslint"] != "^8.0.0" {
		t.Errorf("the project keeps its eslint package: %v", dev)
	}
}

func TestSplitNodeRetiresAFlatConfigWhenTheSharedOneIsChosen(t *testing.T) {
	root := t.TempDir()
	put(t, root, "apps/orders/package.json", orders)
	put(t, root, "eslint.config.mjs", "export default []\n")
	res, err := SplitNode(Request{Root: root, Dir: "apps/orders", Workspace: "shop", Patterns: []string{"apps/*"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "eslint.config.mjs.pre-smith")); err != nil || !res.ESLint {
		t.Errorf("the flat config is kept as .pre-smith and the shared one is wanted: %v %v", err, res.ESLint)
	}
}
