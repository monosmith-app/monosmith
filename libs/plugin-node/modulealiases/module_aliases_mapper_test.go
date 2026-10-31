package modulealiases

import (
	"strings"
	"testing"
)

func TestNoFactGivesNoAliasSettings(t *testing.T) {
	a := FromFacts(nil)
	if a.TSConfigPaths() != "" || a.ViteResolveAlias() != "" || len(a.JestModuleNameMapper()) != 0 {
		t.Fatalf("no aliases must change nothing: %q %q %v", a.TSConfigPaths(), a.ViteResolveAlias(), a.JestModuleNameMapper())
	}
	if len(FromFacts(map[string]any{"module.aliases": "src"})) != 0 {
		t.Error("a fact of the wrong shape gives no aliases")
	}
}

func TestOneAliasReachesEveryTool(t *testing.T) {
	a := FromFacts(map[string]any{"module.aliases": map[string]any{"@app": "./src/"}})
	if got := a["@app"]; got != "src" {
		t.Fatalf("dir = %q", got)
	}
	for tool, got := range map[string]string{"tsconfig": a.TSConfigPaths(), "vite": a.ViteResolveAlias()} {
		if !strings.Contains(got, `"@app"`) || !strings.Contains(got, "src") {
			t.Errorf("%s lacks the alias:\n%s", tool, got)
		}
	}
	if !strings.Contains(a.TSConfigPaths(), `"@app/*": ["./src/*"]`) {
		t.Errorf("paths = %s", a.TSConfigPaths())
	}
	jest := a.JestModuleNameMapper()
	if jest[`^@app/(.*)$`] != "<rootDir>/src/$1" || jest[`^@app$`] != "<rootDir>/src" {
		t.Errorf("jest = %v", jest)
	}
	if !strings.Contains(a.ViteResolveAlias(), `"@app": "/src",`) {
		t.Errorf("vite = %s", a.ViteResolveAlias())
	}
}

func TestAnAliasOfTheProjectRootAndRegexpCharacters(t *testing.T) {
	a := FromFacts(map[string]any{"module.aliases": map[string]string{"~lib.x": "."}})
	if got := a.JestModuleNameMapper(); got[`^~lib\.x/(.*)$`] != "<rootDir>/$1" {
		t.Errorf("jest = %v", got)
	}
	if !strings.Contains(a.TSConfigPaths(), `"~lib.x/*": ["./*"]`) {
		t.Errorf("paths = %s", a.TSConfigPaths())
	}
}
