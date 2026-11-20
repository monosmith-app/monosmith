package adoptsplit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, dir, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestIncludeTestsSplitsTheBuildConfigFromTheOneThatCoversTests(t *testing.T) {
	dir := t.TempDir()
	tsconfig := `{
  // the build
  "extends": "../../tsconfig.base.json",
  "include": ["./src/*.ts"],
  "exclude": ["./src/*.test.ts", "./src/.jest"]
}`
	writeFile(t, dir, "tsconfig.json", tsconfig)
	writeFile(t, dir, "package.json", `{"name": "x", "scripts": {"build": "tsc --outDir ../../dist/x", "watch": "tsc -p other.json -w", "test": "jest"}}`)
	note, err := includeTests(dir)
	if err != nil || note == "" {
		t.Fatalf("note %q, %v", note, err)
	}
	if got := readFile(t, dir, buildTSConfig); got != tsconfig {
		t.Errorf("tsconfig.build.json = %q, want the old tsconfig.json as it was", got)
	}
	var project struct {
		Extends string   `json:"extends"`
		Exclude []string `json:"exclude"`
	}
	if err := json.Unmarshal([]byte(readFile(t, dir, "tsconfig.json")), &project); err != nil {
		t.Fatal(err)
	}
	if project.Extends != "./tsconfig.build.json" || !reflect.DeepEqual(project.Exclude, []string{"./src/.jest"}) {
		t.Errorf("tsconfig.json = %+v", project)
	}
	pkg := readFile(t, dir, "package.json")
	if !strings.Contains(pkg, `"tsc -p tsconfig.build.json --outDir ../../dist/x"`) {
		t.Errorf("the build script does not read tsconfig.build.json: %s", pkg)
	}
	if !strings.Contains(pkg, `"tsc -p other.json -w"`) {
		t.Errorf("a script that names its tsconfig changed: %s", pkg)
	}
	if !strings.Contains(note, "the build script") {
		t.Errorf("note = %q", note)
	}
}

func TestIncludeTestsLeavesAConfigWithoutTestExclusionsAlone(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "tsconfig.json", `{"include": ["src"], "exclude": ["node_modules"]}`)
	if note, err := includeTests(dir); note != "" || err != nil {
		t.Errorf("note %q, %v", note, err)
	}
	if _, err := os.Stat(filepath.Join(dir, buildTSConfig)); err == nil {
		t.Error("tsconfig.build.json written for a config that covers the tests")
	}
}

func TestIsTestPattern(t *testing.T) {
	for p, want := range map[string]bool{
		"./src/*.test.ts": true, "**/*.spec.tsx": true, "src/__tests__": true, "test": true, "./tests/**": true,
		"node_modules": false, "dist": false, "./src/.jest": false, "src/contest.ts": false,
	} {
		if got := isTestPattern(p); got != want {
			t.Errorf("isTestPattern(%q) = %v, want %v", p, got, want)
		}
	}
}
