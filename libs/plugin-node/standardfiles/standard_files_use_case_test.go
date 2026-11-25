package standardfiles

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func put(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func compact(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("not JSON: %s", raw)
	}
	return string(raw)
}

func TestSourceExports(t *testing.T) {
	cases := []struct {
		name, manifest, want string
		ok                   bool
	}{
		{"main and types lead to the build: an exports map with every other path kept",
			`{"main": "./dist/index.js", "types": "./dist/index.d.ts"}`,
			`{".":{"source":"./src/index.ts","types":"./dist/index.d.ts","default":"./dist/index.js"},"./*":"./*"}`, true},
		{"a string exports",
			`{"exports": "./dist/index.js", "types": "dist/index.d.ts"}`,
			`{".":{"source":"./src/index.ts","types":"./dist/index.d.ts","default":"./dist/index.js"}}`, true},
		{"a subpath map keeps its other paths and its order",
			`{"exports": {"./package.json": "./package.json", ".": {"types": "./dist/index.d.ts", "import": "./dist/index.js"}, "./utils": "./dist/utils.js"}}`,
			`{"./package.json":"./package.json",".":{"source":"./src/index.ts","types":"./dist/index.d.ts","import":"./dist/index.js"},"./utils":"./dist/utils.js"}`, true},
		{"a conditions-only exports gets the condition first",
			`{"exports": {"types": "./dist/index.d.ts", "default": "./dist/index.js"}}`,
			`{"source":"./src/index.ts","types":"./dist/index.d.ts","default":"./dist/index.js"}`, true},
		{"an old source condition is moved first, not repeated",
			`{"exports": {".": {"types": "./dist/index.d.ts", "source": "./src/old.ts", "default": "./dist/index.js"}}}`,
			`{".":{"source":"./src/index.ts","types":"./dist/index.d.ts","default":"./dist/index.js"}}`, true},
		{"a library that exports its sources is left alone", `{"exports": "./src/index.ts"}`, "", false},
		{"a manifest with no entry point is left alone", `{"name": "x"}`, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := SourceExports([]byte(c.manifest), "./src/index.ts")
			if ok != c.ok || ok && compact(t, got) != c.want {
				t.Fatalf("got %s %v\nwant %s", got, ok, c.want)
			}
		})
	}
}

func TestBaseCompilerOptions(t *testing.T) {
	got, ok := BaseCompilerOptions([]byte(`{"extends": "@monosmith/tsconfig/node.json", "compilerOptions": {"noUnusedLocals": true, "strict": true}}`))
	if !ok || string(got) != `{"noUnusedLocals":true,"strict":true,"customConditions":["source"]}` {
		t.Fatalf("the condition joins the options, which keep their order: %s %v", got, ok)
	}
	if got, ok := BaseCompilerOptions([]byte(`{"compilerOptions": {"customConditions": ["development", "source"]}}`)); !ok || string(got) != `{"customConditions":["development","source"]}` {
		t.Fatalf("a base that has it is unchanged: %s", got)
	}
	if got, ok := BaseCompilerOptions([]byte(`{"extends": "x"}`)); !ok || string(got) != `{"customConditions":["source"]}` {
		t.Fatalf("a base with no options gets them: %s", got)
	}
	for _, base := range []string{
		`{"compilerOptions": {"moduleResolution": "node"}}`,
		`{"compilerOptions": {"module": "commonjs"}}`,
		"{\n  // a comment\n  \"compilerOptions\": {}\n}",
	} {
		if _, ok := BaseCompilerOptions([]byte(base)); ok {
			t.Errorf("node10 resolution ignores conditions, and comments would be lost: %s", base)
		}
	}
}

func TestReplaceableTsconfig(t *testing.T) {
	standard := ProjectTsconfig("libs/models")
	if !ReplaceableTsconfig("libs/models", []byte(standard), "") {
		t.Fatal("the standard itself is replaceable")
	}
	cases := map[string]bool{
		`{"extends": "../../tsconfig.base.json", "compilerOptions": {"outDir": "dist", "declaration": true}, "include": ["src"]}`:   true,
		`{"extends": "../../tsconfig.base.json", "compilerOptions": {"strict": false}, "include": ["src"]}`:                         false,
		`{"extends": "../../tsconfig.base.json", "compilerOptions": {"jsx": "react-jsx"}}`:                                          false,
		`{"extends": "../../tsconfig.base.json", "compilerOptions": {"paths": {"x": ["y"]}}}`:                                       false,
		`{"extends": "@tsconfig/node20/tsconfig.json", "include": ["src"]}`:                                                         false,
		`{"compilerOptions": {"rootDir": "src"}}`:                                                                                   false,
		`{"extends": "../../tsconfig.base.json", "include": ["src", "scripts"]}`:                                                    false,
		`{"extends": "../../tsconfig.base.json", "references": [{"path": "./tsconfig.lib.json"}]}`:                                  false,
		`{"extends": "../../tsconfig.base.json", "compilerOptions": {"types": ["node", "jest"]}, "include": ["src"]}`:               false,
		"{\n  // comments are fine: the copy keeps them\n  \"extends\": \"../../tsconfig.base.json\",\n  \"include\": [\"src\"]\n}": true,
	}
	for raw, want := range cases {
		if got := ReplaceableTsconfig("libs/models", []byte(raw), "tsc -p tsconfig.build.json --outDir ../../dist/libs/models"); got != want {
			t.Errorf("%s: got %v", raw, got)
		}
	}
	if ReplaceableTsconfig("libs/models", []byte(`{"extends": "../../tsconfig.base.json", "compilerOptions": {"outDir": "dist"}}`), "tsc -p tsconfig.build.json") {
		t.Error("a build that relies on the tsconfig's outDir keeps it")
	}
}

func TestStandard(t *testing.T) {
	var root string
	beforeEach := func() {
		root = t.TempDir()
		put(t, root, "tsconfig.base.json", `{"extends": "@monosmith/tsconfig/node.json", "compilerOptions": {"strict": true}}`)
		put(t, root, "libs/models/package.json", `{"name": "@quotes/models", "main": "./dist/index.js", "types": "./dist/index.d.ts", "scripts": {"build": "tsc -p tsconfig.build.json --outDir dist"}}`)
		put(t, root, "libs/models/src/index.ts", "export const x = 1\n")
		put(t, root, "libs/models/tsconfig.json", ProjectTsconfig("libs/models"))
		put(t, root, "libs/models/tsconfig.build.json", `{"extends": "./tsconfig.json", "exclude": ["src/**/*.spec.ts"]}`)
		put(t, root, "libs/models/src/index.test.ts", "")
		put(t, root, "apps/web/package.json", `{"name": "web", "main": "./dist/main.js"}`)
		put(t, root, "apps/web/src/main.tsx", "")
		put(t, root, "apps/web/tsconfig.json", `{"extends": "../../tsconfig.base.json", "compilerOptions": {"jsx": "react-jsx"}}`)
	}
	projects := []protocol.Project{{Name: "models", Root: "libs/models", Kind: protocol.ProjectLib}, {Name: "web", Root: "apps/web", Kind: protocol.ProjectApp}}
	paths := func(res protocol.StandardResult) map[string]protocol.StandardFileContent {
		out := map[string]protocol.StandardFileContent{}
		for _, f := range res.Files {
			out[f.Path] = f
		}
		return out
	}

	t.Run("a converted library gets the source condition, and the base reads it", func(t *testing.T) {
		beforeEach()
		res, err := Standard(root, protocol.StandardParams{Projects: projects})
		if err != nil {
			t.Fatal(err)
		}
		got := paths(res)
		pkg, base := got["libs/models/package.json"], got["tsconfig.base.json"]
		if pkg.Mode != protocol.StandardKeys || string(pkg.Keys["exports"]) == "" || pkg.Reason == "" {
			t.Fatalf("package.json: %+v", pkg)
		}
		if base.Project != "" || string(base.Keys["compilerOptions"]) != `{"strict":true,"customConditions":["source"]}` {
			t.Fatalf("base: %+v", base)
		}
		if b := got["libs/models/tsconfig.build.json"]; b.Content != BuildTsconfig("test") {
			t.Fatalf("the build tsconfig leaves the workspace's tests out: %q", b.Content)
		}
		if _, ok := got["apps/web/tsconfig.json"]; ok {
			t.Fatal("a React app keeps its tsconfig")
		}
	})

	t.Run("a base that ignores conditions leaves the libraries alone", func(t *testing.T) {
		beforeEach()
		put(t, root, "tsconfig.base.json", `{"compilerOptions": {"module": "commonjs"}}`)
		res, err := Standard(root, protocol.StandardParams{Projects: projects})
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := paths(res)["libs/models/package.json"]; ok {
			t.Fatal("a condition nothing reads is not offered")
		}
	})

	t.Run("an Nx layout keeps its tsconfig files", func(t *testing.T) {
		beforeEach()
		put(t, root, "libs/models/tsconfig.lib.json", "{}")
		res, _ := Standard(root, protocol.StandardParams{Projects: projects})
		if _, ok := paths(res)["libs/models/tsconfig.json"]; ok {
			t.Fatal("tsconfig.json of an Nx project only references tsconfig.lib.json")
		}
	})
}
