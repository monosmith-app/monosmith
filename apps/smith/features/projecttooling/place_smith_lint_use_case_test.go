package projecttooling

import (
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/removalplan"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/eslintpeer"
)

func TestPlaceSmithLintWritesTheRootConfigTheDependenciesAndTheLintScript(t *testing.T) {
	root := writeTree(t, libraryFixture)
	got, err := PlaceSmithLint(root, "packages/contrast", &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ready || len(got.Left) != 0 {
		t.Fatalf("outcome %+v", got)
	}
	if readFile(t, root, "eslint.config.mjs") != removalplan.SmithLintConfig {
		t.Errorf("config %q", readFile(t, root, "eslint.config.mjs"))
	}
	manifest := readFile(t, root, "package.json")
	for _, want := range []string{`"@monosmith/eslint-config": "` + eslintpeer.ConfigRange, `"eslint": "` + eslintpeer.Range, `"eslint-plugin-jsx-a11y"`, `"eslint": "$eslint"`} {
		if !strings.Contains(manifest, want) {
			t.Errorf("root package.json lacks %s:\n%s", want, manifest)
		}
	}
	if !strings.Contains(readFile(t, root, "packages/contrast/package.json"), `"lint": "eslint ."`) {
		t.Errorf("no lint script:\n%s", readFile(t, root, "packages/contrast/package.json"))
	}
	if readFile(t, root, ".eslintrc.js.pre-smith") == "" {
		t.Error("the old config is the removal plan's to delete, not this step's")
	}
}

func TestPlaceSmithLintStopsAtARootConfigItCannotMakeSmiths(t *testing.T) {
	files := map[string]string{
		"package.json":      `{"name": "w", "private": true}`,
		"eslint.config.mjs": "import mnci from '@mnci/eslint-config'\nexport default mnci({ workspaceRoot: import.meta.dirname, strict: true })\n",
	}
	root := writeTree(t, files)
	got, err := PlaceSmithLint(root, ".", &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Ready || len(got.Left) != 1 || !strings.Contains(got.Left[0], "eslint.config.mjs") {
		t.Errorf("outcome %+v", got)
	}
	if readFile(t, root, "eslint.config.mjs") != files["eslint.config.mjs"] {
		t.Error("the config was changed")
	}
}

func TestPlaceSmithLintDropsTheEslintConfigKey(t *testing.T) {
	root := writeTree(t, map[string]string{
		"package.json": `{"name": "w", "private": true, "eslintConfig": {"extends": ["react-app"]}, "scripts": {"lint": "eslint src --ext .ts"}}`,
	})
	if _, err := PlaceSmithLint(root, ".", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	manifest := readFile(t, root, "package.json")
	if strings.Contains(manifest, "eslintConfig") || !strings.Contains(manifest, `"lint": "eslint ."`) {
		t.Errorf("package.json:\n%s", manifest)
	}
}
