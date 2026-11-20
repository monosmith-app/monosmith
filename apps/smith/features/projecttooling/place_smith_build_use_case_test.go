package projecttooling

import (
	"bytes"
	"strings"
	"testing"
)

func TestPlaceSmithBuildMovesATscBuildFromLibIntoDist(t *testing.T) {
	root := writeTree(t, libraryFixture)
	a, err := Assess(root, "packages/contrast")
	if err != nil {
		t.Fatal(err)
	}
	got, err := PlaceSmithBuild(root, a, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ready {
		t.Fatalf("outcome %+v", got)
	}
	manifest := readFile(t, root, "packages/contrast/package.json")
	for _, want := range []string{`"main": "./dist/index.js"`, `"types": "./dist/index.d.ts"`} {
		if !strings.Contains(manifest, want) {
			t.Errorf("package.json lacks %s:\n%s", want, manifest)
		}
	}
	if strings.Contains(manifest, `"lib"`) {
		t.Errorf("files still names lib:\n%s", manifest)
	}
	if got := readFile(t, root, ".gitignore"); got != "dist\nnode_modules\n" {
		t.Errorf(".gitignore %q", got)
	}
	if a, _ := Assess(root, "packages/contrast"); ClassifyBuild(a.Facts) != BuildSmith {
		t.Errorf("the build is not smith's afterwards: %+v", a.Facts)
	}
}

func TestPlaceSmithBuildLeavesABundlerAndSaysSo(t *testing.T) {
	root := writeTree(t, map[string]string{"package.json": `{"name": "w", "scripts": {"build": "rollup -c"}}`})
	a, _ := Assess(root, ".")
	got, err := PlaceSmithBuild(root, a, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Ready || len(got.Left) != 1 || !strings.Contains(got.Left[0], "rollup") {
		t.Errorf("outcome %+v", got)
	}
}

// craFixture is a create-react-app project in apps/demo with one test.
var craFixture = map[string]string{
	"package.json": `{"name": "w", "private": true, "workspaces": ["apps/*"], "devDependencies": {"@types/node": "^16.11.6"}}`,
	"apps/demo/package.json": `{
  "name": "demo",
  "homepage": "https://example.org/demo",
  "scripts": {
    "start": "react-scripts start",
    "build": "react-scripts build",
    "test": "react-scripts test",
    "eject": "react-scripts eject"
  },
  "dependencies": {"react": "^17.0.2", "react-scripts": "4.0.3"}
}
`,
	"apps/demo/public/index.html":      "<html><head><link rel=\"icon\" href=\"%PUBLIC_URL%/favicon.ico\"></head><body><div id=\"root\"></div></body></html>",
	"apps/demo/src/index.tsx":          "import App from './App'\n",
	"apps/demo/src/App.test.tsx":       "it('works', () => { expect(jest.fn()).toBeDefined() })\n",
	"apps/demo/src/setupTests.ts":      "import '@testing-library/jest-dom'\n",
	"apps/demo/src/react-app-env.d.ts": "/// <reference types=\"react-scripts\" />\n",
}

func TestPlaceSmithBuildMovesCreateReactAppOntoVite(t *testing.T) {
	root := writeTree(t, craFixture)
	a, _ := Assess(root, "apps/demo")
	got, err := PlaceSmithBuild(root, a, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ready || len(got.Left) != 1 || !strings.Contains(got.Left[0], "jest.") {
		t.Fatalf("outcome %+v", got)
	}
	if readFile(t, root, "apps/demo/public/index.html") != "" {
		t.Error("public/index.html stays")
	}
	if !strings.Contains(readFile(t, root, "apps/demo/index.html"), `src="/src/index.tsx"`) {
		t.Errorf("index.html: %s", readFile(t, root, "apps/demo/index.html"))
	}
	config := readFile(t, root, "apps/demo/vite.config.ts")
	for _, want := range []string{"'../../dist/apps/demo'", "base:    '/demo/'", "setupFiles:  ['src/setupTests.ts']"} {
		if !strings.Contains(config, want) {
			t.Errorf("vite.config.ts lacks %s:\n%s", want, config)
		}
	}
	manifest := readFile(t, root, "apps/demo/package.json")
	if strings.Contains(manifest, "react-scripts start") || strings.Contains(manifest, "eject") || !strings.Contains(manifest, `"test": "vitest run"`) {
		t.Errorf("scripts:\n%s", manifest)
	}
	rootManifest := readFile(t, root, "package.json")
	if strings.Contains(rootManifest, `"^16.11.6"`) {
		t.Errorf("an @types/node older than Vite's peer stays:\n%s", rootManifest)
	}
	for _, want := range []string{`"vite"`, `"@vitejs/plugin-react"`, `"vitest"`, `"jsdom"`} {
		if !strings.Contains(rootManifest, want) {
			t.Errorf("root package.json lacks %s:\n%s", want, rootManifest)
		}
	}
	if !strings.Contains(readFile(t, root, "apps/demo/src/react-app-env.d.ts"), "vite/client") {
		t.Error("react-app-env.d.ts still names react-scripts")
	}
}
