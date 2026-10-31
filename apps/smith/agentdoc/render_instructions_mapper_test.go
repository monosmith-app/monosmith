package agentdoc

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func sampleFacts() Facts {
	return Facts{
		Name:          "acme",
		Architectures: []Architecture{{Name: "vertical-slices", Meaning: "code grouped by what it does"}},
		Languages:     []string{"go", "node"},
		Projects: []Project{
			{Name: "api", Root: "apps/api", Kind: "app", Language: "go", Stack: "go"},
			{Name: "web", Root: "apps/web", Kind: "app", Language: "node", Stack: "react-vitest", Choices: []string{"react", "vitest"}},
		},
		Tasks:   []string{"lint", "test", "build"},
		Derive:  true,
		Derived: []string{"apps/web/tsconfig.lib.json"},
		Ejected: []string{"apps/web/jest.config.js"},
		Branch:  "main",
	}
}

func TestRenderMatchesTheGoldenFile(t *testing.T) {
	got := Render(sampleFacts())
	golden := filepath.Join("testdata", "agents_region.golden.md")
	if *update {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("Render differs from %s; run go test ./agentdoc -update and read the diff\n%s", golden, got)
	}
}

func TestRenderIsDeterministic(t *testing.T) {
	first, second := Render(sampleFacts()), Render(sampleFacts())
	if first != second {
		t.Error("two renders of the same facts differ")
	}
}

func TestRenderOfAWorkspaceWithNoRulesStaysShort(t *testing.T) {
	got := Render(Facts{Branch: "main"})
	for _, want := range []string{"no architecture convention", "Derived files are off", "no project yet", "smith ci verify"} {
		if !strings.Contains(got, want) {
			t.Errorf("render lacks %q", want)
		}
	}
}
