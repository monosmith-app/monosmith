package agentdoc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-architectures/rules"
	"github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/languages"
)

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
