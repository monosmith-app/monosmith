package agentdoc

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-architectures/rules"
	"github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/languages"
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

// The agent file must say what `smith lint` enforces, from the same data.
func TestTheAgentFileStatesEveryRoleAndForbiddenFolderTheCheckKnows(t *testing.T) {
	got := Render(sampleFacts())
	for _, role := range languages.Roles {
		if !strings.Contains(got, "`"+role+"`") {
			t.Errorf("the agent file lacks the role %q", role)
		}
	}
	for name := range languages.ForbiddenFolders {
		if strings.HasSuffix(name, "s") && languages.ForbiddenFolders[strings.TrimSuffix(name, "s")] {
			continue // the plural is covered by "(or their plurals)"
		}
		if !strings.Contains(got, name) {
			t.Errorf("the agent file lacks the forbidden folder %q", name)
		}
	}
}

func TestTheConventionsPageListsEveryRole(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "reference", "conventions.md"))
	if err != nil {
		t.Skip("the docs are not next to the code")
	}
	for _, role := range languages.Roles {
		if !strings.Contains(string(raw), "`"+role+"`") {
			t.Errorf("docs/reference/conventions.md lacks the role %q", role)
		}
	}
}

func TestTheAgentFileListsTheLayersOfEachRuleset(t *testing.T) {
	for _, name := range rules.Names() {
		rs, _ := rules.For(name)
		var layers []Layer
		for _, l := range rs.Layers {
			layers = append(layers, Layer{Name: l.Name, Meaning: l.Meaning, Folders: l.Folders})
		}
		got := Render(Facts{Architectures: []Architecture{{Name: name, Meaning: "m", Rules: rs.Summary, Layers: layers}}, Branch: "main"})
		for _, l := range rs.Layers {
			if !strings.Contains(got, "`"+l.Name+"`") {
				t.Errorf("%s: the agent file lacks the layer %q", name, l.Name)
			}
		}
	}
}
