package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-recipe/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerateWritesARecipeThatIsAProject(t *testing.T) {
	ws := t.TempDir()
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "team", Directory: "recipes/team"})
	if err != nil || len(res.Files) != 1 {
		t.Fatalf("%v %+v", err, res)
	}
	src, err := os.ReadFile(filepath.Join(ws, "recipes", "team", "team.recipe.toml"))
	if err != nil || !strings.Contains(string(src), `name = "team"`) || !strings.Contains(string(src), "[when.language.python.use]") {
		t.Fatalf("%v\n%s", err, src)
	}
	found, err := discovery.Discover(ws, []string{"recipes/team/team.recipe.toml"})
	if err != nil || len(found.Projects) != 1 || found.Projects[0].Name != "team" || found.Projects[0].Version != "0.0.0" {
		t.Fatalf("the new recipe is not a project: %v %+v", err, found.Projects)
	}
}

func TestGenerateRefusesWhatCannotBeReleased(t *testing.T) {
	ws := t.TempDir()
	if _, err := Generate(ws, protocol.GenerateParams{Name: "Team", Directory: "recipes/Team"}); err == nil || !strings.Contains(err.Error(), "lowercase words") {
		t.Errorf("a name a tag cannot use: %v", err)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Name: "team", Directory: "recipes/team"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Name: "team", Directory: "recipes/team"}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("an existing folder: %v", err)
	}
}
