package projecttooling

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// writeTree writes files (workspace-relative path to content) under a new
// temporary folder and returns it.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// readFile reads a workspace-relative file, "" when it is missing.
func readFile(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return ""
	}
	return string(raw)
}

// listFiles lists the files under root, workspace-relative.
func listFiles(root string) func() ([]string, error) {
	return func() ([]string, error) {
		var out []string
		err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(root, p)
			out = append(out, filepath.ToSlash(rel))
			return nil
		})
		sort.Strings(out)
		return out, err
	}
}

// libraryFixture is a TypeScript package after adopt's move: an ESLint 8
// setup put aside, a tsc build into lib/.
var libraryFixture = map[string]string{
	"package.json": `{
  "name": "acme-workspace",
  "private": true,
  "workspaces": ["packages/*"],
  "devDependencies": {
    "@typescript-eslint/parser": "^5.2.0",
    "eslint-config-standard": "^16.0.3",
    "typescript": "^4.4.4"
  }
}
`,
	".eslintrc.js.pre-smith": "module.exports = { extends: ['standard'] }\n",
	".eslintignore":          "lib\n",
	"packages/contrast/package.json": `{
  "name": "contrast",
  "version": "1.0.0",
  "main": "lib/index.js",
  "types": "lib/index.d.ts",
  "files": ["lib"],
  "scripts": {
    "build": "tsc",
    "test": "jest"
  }
}
`,
	"packages/contrast/tsconfig.json": `{
  "compilerOptions": {
    "declaration": true,
    "outDir": "./lib",
    "strict": true
  }
}
`,
	"packages/contrast/src/index.ts": "export const x = 1\n",
}

func TestAssessFindsTheLegacyToolsOfAMovedPackage(t *testing.T) {
	root := writeTree(t, libraryFixture)
	a, err := Assess(root, "packages/contrast")
	if err != nil {
		t.Fatal(err)
	}
	if !a.Node || a.Name != "contrast" || a.Kind != "package" || a.Recommended != Smith {
		t.Fatalf("assessment %+v", a)
	}
	joined := strings.Join(a.Legacy, "; ")
	if !strings.Contains(joined, ".eslintrc.js.pre-smith") || !strings.Contains(joined, "a tsc build into lib/") {
		t.Errorf("legacy %q", a.Legacy)
	}
}

func TestAssessOfAFolderWithoutPackageJSONIsNotNode(t *testing.T) {
	a, err := Assess(writeTree(t, map[string]string{"py/pyproject.toml": ""}), "py")
	if err != nil || a.Node {
		t.Errorf("%+v %v", a, err)
	}
}
