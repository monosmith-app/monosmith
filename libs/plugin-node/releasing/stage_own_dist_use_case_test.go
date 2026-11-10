package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStageMakesAPackageThatKeepsItsOwnDistPathsRelativeToTheWorkspaceBuild(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("packages/kit/package.json", `{
  "name": "@a/kit",
  "files": ["dist", "!**/*.tsbuildinfo", "!**/*.js.map"],
  "main": "./dist/index.esm.js",
  "types": "./dist/src/index.d.ts",
  "exports": {
    "./package.json": "./package.json",
    ".": { "types": "./dist/src/index.d.ts", "default": "./dist/index.esm.js" }
  },
  "scripts": { "build": "esbuild" }
}
`)
	write("dist/packages/kit/index.esm.js", "")
	dir, err := Stage(ws, "packages/kit")
	if err != nil {
		t.Fatal(err)
	}
	staged, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	for _, want := range []string{`"main": "./index.esm.js"`, `"types": "./src/index.d.ts"`, `"default": "./index.esm.js"`, `"./package.json": "./package.json"`, `"**"`, `"!**/*.tsbuildinfo"`} {
		if !strings.Contains(string(staged), want) {
			t.Errorf("staged package.json lacks %s:\n%s", want, staged)
		}
	}
	for _, gone := range []string{"./dist/", `"dist"`, "scripts"} {
		if strings.Contains(string(staged), gone) {
			t.Errorf("staged package.json still has %s:\n%s", gone, staged)
		}
	}
}
