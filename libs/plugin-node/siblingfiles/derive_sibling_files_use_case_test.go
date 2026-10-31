package siblingfiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func put(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func projects(names ...string) []protocol.Project {
	var out []protocol.Project
	for _, n := range names {
		out = append(out, protocol.Project{Name: n, Root: "packages/" + n, Kind: protocol.ProjectKind("package"), Exports: []string{"@s/" + n}})
	}
	return out
}

// noRefs answers that no project has an internal dependency.
func noRefs(protocol.Project, []byte) []any { return []any{} }

func fileOf(res Result, rel string) (protocol.DerivedFile, bool) {
	for _, f := range res.Files {
		if f.Path == rel {
			return f, true
		}
	}
	return protocol.DerivedFile{}, false
}

func divergingOf(res Result, rel string) (protocol.DivergentFile, bool) {
	for _, d := range res.Diverging {
		if d.Path == rel {
			return d, true
		}
	}
	return protocol.DivergentFile{}, false
}

const swc = "{\n  \"jsc\": {\n    \"target\": \"es2017\"\n  },\n  \"sourceMaps\": true\n}\n"

func TestAFileThatMostSiblingsShareIsDerivedWholeForEachProject(t *testing.T) {
	root := t.TempDir()
	for _, n := range []string{"a", "b", "c", "d"} {
		put(t, root, "packages/"+n+"/.spec.swcrc", swc)
	}
	put(t, root, "packages/d/.spec.swcrc", strings.Replace(swc, "es2017", "es5", 1))
	res := Derive(Input{Root: root, Projects: projects("a", "b", "c", "d"), References: noRefs})
	for _, n := range []string{"a", "b", "c"} {
		f, ok := fileOf(res, "packages/"+n+"/.spec.swcrc")
		if !ok || f.Mode != protocol.DeriveFile || f.Content != swc || f.Renderer != Renderer {
			t.Fatalf("%s: %+v", n, f)
		}
	}
	if f, ok := fileOf(res, "packages/d/.spec.swcrc"); !ok || f.Mode != protocol.DeriveKeys || f.Keys["jsc"] != nil || f.Keys["sourceMaps"] != true {
		t.Fatalf("d derives the key it shares and keeps its own jsc: %+v", f)
	}
	d, ok := divergingOf(res, "packages/d/.spec.swcrc")
	if !ok || !strings.Contains(d.Detail, "diverges from siblings") {
		t.Fatalf("diverging: %+v", res.Diverging)
	}
}

func TestADeletedOwnedFileComesBackWithTheValuesOfItsProject(t *testing.T) {
	root := t.TempDir()
	cfg := "module.exports = {\n  displayName: '@s/%s',\n  name: '%s',\n  preset: '../../jest.preset.js'\n}\n"
	for _, n := range []string{"a", "b", "c"} {
		put(t, root, "packages/"+n+"/jest.config.cts", strings.ReplaceAll(cfg, "%s", n))
	}
	res := Derive(Input{Root: root, Projects: projects("a", "b", "c", "gone"), Owned: []string{"packages/gone/jest.config.cts"}, References: noRefs})
	f, ok := fileOf(res, "packages/gone/jest.config.cts")
	if !ok || f.Content != strings.ReplaceAll(cfg, "%s", "gone") || f.Mode != protocol.DeriveFile {
		t.Fatalf("%+v", f)
	}
}

func TestAFileNobodyOwnsIsNotCreatedForAProjectThatLacksIt(t *testing.T) {
	root := t.TempDir()
	for _, n := range []string{"a", "b", "c"} {
		put(t, root, "packages/"+n+"/.spec.swcrc", swc)
	}
	res := Derive(Input{Root: root, Projects: projects("a", "b", "c", "none"), References: noRefs})
	if _, ok := fileOf(res, "packages/none/.spec.swcrc"); ok || len(res.Files) != 3 {
		t.Fatalf("%+v", res.Files)
	}
}

func TestTooFewSiblingsOrNoMajorityDeriveNothing(t *testing.T) {
	root := t.TempDir()
	put(t, root, "packages/a/.swcrc", swc)
	put(t, root, "packages/b/.swcrc", swc)
	for i, n := range []string{"a", "b", "c", "d", "e"} {
		put(t, root, "packages/"+n+"/tsconfig.json", `{"x": `+string(rune('1'+i))+`}`)
	}
	put(t, root, "packages/a/tsconfig.json", `{"x": 1}`)
	put(t, root, "packages/b/tsconfig.json", `{"x": 1}`)
	res := Derive(Input{Root: root, Projects: projects("a", "b", "c", "d", "e"), References: noRefs})
	if len(res.Files) != 0 || len(res.Diverging) != 0 {
		t.Fatalf("two files are not a convention, nor two of five: %+v", res)
	}
}

func TestReferencesFollowTheDependenciesOfEachProject(t *testing.T) {
	root := t.TempDir()
	lib := "{\n  \"extends\": \"../../tsconfig.base.json\",\n  \"references\": [\n    {\n      \"path\": \"../%s/tsconfig.lib.json\"\n    }\n  ],\n  \"include\": [\n    \"src\"\n  ]\n}\n"
	deps := map[string]string{"a": "x", "b": "y", "c": "z"}
	for n, d := range deps {
		put(t, root, "packages/"+n+"/tsconfig.lib.json", strings.Replace(lib, "%s", d, 1))
	}
	put(t, root, "packages/leaf/tsconfig.lib.json", "{\n  \"extends\": \"../../tsconfig.base.json\",\n  \"include\": [\n    \"src\"\n  ]\n}\n")
	refs := func(p protocol.Project, _ []byte) []any {
		if d, ok := deps[p.Name]; ok {
			return []any{map[string]any{"path": "../" + d + "/tsconfig.lib.json"}}
		}
		if p.Name == "gone" {
			return []any{map[string]any{"path": "../q/tsconfig.lib.json"}}
		}
		return []any{}
	}
	res := Derive(Input{Root: root, Projects: projects("a", "b", "c", "leaf", "gone"), Owned: []string{"packages/gone/tsconfig.lib.json"}, References: refs})
	for _, n := range []string{"a", "b", "c", "leaf"} {
		if f, ok := fileOf(res, "packages/"+n+"/tsconfig.lib.json"); !ok || f.Mode != protocol.DeriveFile {
			t.Fatalf("%s: %+v", n, f)
		}
	}
	gone, _ := fileOf(res, "packages/gone/tsconfig.lib.json")
	if want := strings.Replace(lib, "%s", "q", 1); gone.Content != want {
		t.Fatalf("restored:\n%s\nwant:\n%s", gone.Content, want)
	}
}

func TestAJsonFileThatDiffersInOneKeyIsDerivedForTheOthersAndReported(t *testing.T) {
	root := t.TempDir()
	body := "{\n  \"extends\": \"../../tsconfig.base.json\",\n  \"include\": [\n    \"src\"\n  ],\n  \"exclude\": [\n    %s\n  ]\n}\n"
	put(t, root, "packages/a/tsconfig.lib.json", strings.Replace(body, "%s", `"e1"`, 1))
	put(t, root, "packages/b/tsconfig.lib.json", strings.Replace(body, "%s", `"e2"`, 1))
	put(t, root, "packages/c/tsconfig.lib.json", strings.Replace(body, "%s", `"e3"`, 1))
	put(t, root, "packages/d/tsconfig.lib.json", strings.Replace(strings.Replace(body, "%s", `"e4"`, 1), `"src"`, `"lib"`, 1))
	res := Derive(Input{Root: root, Projects: projects("a", "b", "c", "d"), References: noRefs})
	a, ok := fileOf(res, "packages/a/tsconfig.lib.json")
	if !ok || a.Mode != protocol.DeriveKeys || a.Keys["extends"] != "../../tsconfig.base.json" || a.Keys["include"] == nil || a.Keys["exclude"] != nil {
		t.Fatalf("a: %+v", a)
	}
	d, ok := fileOf(res, "packages/d/tsconfig.lib.json")
	if !ok || d.Keys["include"] != nil || d.Keys["extends"] == nil {
		t.Fatalf("d keeps its own include: %+v", d)
	}
	if dv, ok := divergingOf(res, "packages/d/tsconfig.lib.json"); !ok || !strings.Contains(dv.Detail, "include") {
		t.Fatalf("diverging: %+v", res.Diverging)
	}
}

func TestOnlyConfigFilesAreCandidates(t *testing.T) {
	for name, want := range map[string]bool{
		"tsconfig.json": true, "tsconfig.lib.json": true, ".spec.swcrc": true, "jest.config.cts": true, "vitest.config.ts": true,
		"package.json": false, "test-setup.ts": false, "index.ts": false, "README.md": false, "tsconfig.tsbuildinfo": false,
	} {
		if IsConfigFile(name) != want {
			t.Errorf("%s: want %v", name, want)
		}
	}
}
