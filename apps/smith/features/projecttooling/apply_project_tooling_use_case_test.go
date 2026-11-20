package projecttooling

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestApplyKeepChangesNothing(t *testing.T) {
	root := writeTree(t, libraryFixture)
	res, err := Apply(context.Background(), Input{Root: root, Dir: "packages/contrast", Answer: Keep}, RunOptions{Out: &bytes.Buffer{}, Files: listFiles(root)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Lint || res.Build || len(res.Applied) > 0 {
		t.Errorf("result %+v", res)
	}
	if readFile(t, root, "packages/contrast/package.json") != libraryFixture["packages/contrast/package.json"] {
		t.Error("the project changed")
	}
}

func TestApplySmithPutsSmithsToolsInPlaceThenRetiresTheOldOnes(t *testing.T) {
	root := writeTree(t, libraryFixture)
	var out bytes.Buffer
	res, err := Apply(context.Background(), Input{Root: root, Dir: "packages/contrast", Answer: Smith}, RunOptions{
		Out: &out, Yes: true, SkipFinish: true, Smith: "smith", Files: listFiles(root), WholeRoot: true,
	})
	if err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if !res.Lint || !res.Build {
		t.Fatalf("result %+v\n%s", res, out.String())
	}
	if readFile(t, root, ".eslintrc.js.pre-smith") != "" || readFile(t, root, ".eslintignore") != "" {
		t.Errorf("the old lint files stay:\n%s", out.String())
	}
	manifest := readFile(t, root, "package.json")
	if strings.Contains(manifest, "eslint-config-standard") || strings.Contains(manifest, "@typescript-eslint/parser") {
		t.Errorf("the old lint packages stay:\n%s", manifest)
	}
	if !strings.Contains(manifest, "@monosmith/eslint-config") {
		t.Errorf("smith's config is not there:\n%s", manifest)
	}
}

func TestApplyWithoutTheWholeRootLeavesTheWorkspacesOwnLintFiles(t *testing.T) {
	root := writeTree(t, libraryFixture)
	var out bytes.Buffer
	if _, err := Apply(context.Background(), Input{Root: root, Dir: "packages/contrast", Answer: Smith}, RunOptions{
		Out: &out, Yes: true, SkipFinish: true, Smith: "smith", Files: listFiles(root),
	}); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if readFile(t, root, ".eslintrc.js.pre-smith") == "" || readFile(t, root, ".eslintignore") == "" {
		t.Errorf("an imported project removed the workspace's own lint files:\n%s", out.String())
	}
}

func TestInScopeIsTheProjectFolderAndTheRootFiles(t *testing.T) {
	files := []string{"package.json", ".eslintrc.js", "eslint.config.mjs", "apps/web/src/a.ts", "apps/web/package.json", "apps/api/package.json", "docs/readme.md"}
	got := InScope(files, "apps/web", false)
	want := []string{"package.json", "eslint.config.mjs", "apps/web/src/a.ts", "apps/web/package.json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v", got)
	}
	got = InScope(files, "apps/web", true)
	want = []string{"package.json", ".eslintrc.js", "eslint.config.mjs", "apps/web/src/a.ts", "apps/web/package.json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("with the whole root: got %v", got)
	}
	if !reflect.DeepEqual(InScope(files, ".", false), files) {
		t.Error("a project at the root sees every file")
	}
}
