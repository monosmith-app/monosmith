package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestARecipeInAFolderOfItsOwnIsAPackageProject(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "recipes/company/company.recipe.toml", "name = \"company\"\nversion = \"3.0.0\"\n")
	write(t, ws, "recipes/company/README.md", "# company")
	write(t, ws, "recipes/team/team.recipe.toml", "name = \"team\"\nversion = \"1.2.0\"\nextends = \"../company/company.recipe.toml\"\n")
	got, err := Discover(ws, []string{"recipes/company/company.recipe.toml", "recipes/team/team.recipe.toml"})
	if err != nil || len(got.Projects) != 2 {
		t.Fatalf("%v %+v", err, got.Projects)
	}
	team := got.Projects[1]
	if team.Name != "team" || team.Root != "recipes/team" || team.Language != "recipe" || team.Kind != protocol.ProjectPackage || team.Version != "1.2.0" {
		t.Errorf("team = %+v", team)
	}
	if len(team.Imports) != 1 || team.Imports[0] != "recipes/company" {
		t.Errorf("the parent's folder is imported, so a change there affects the child: %v", team.Imports)
	}
	if len(got.Projects[0].Imports) != 0 {
		t.Errorf("company imports %v", got.Projects[0].Imports)
	}
	if task := team.Tasks["test"]; len(task.Command) != 5 || task.Command[2] != "check" || task.Command[3] != "team.recipe.toml" || task.Command[4] != "--all" || !task.Cache {
		t.Errorf("test = %+v", task)
	}
}

func TestRecipesThatAreNotProjectsAreLeftAlone(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "team.recipe.toml", "name = \"root\"\n")                        // at the root: smith.toml's recipe
	write(t, ws, "recipes/a.recipe.toml", "name = \"a\"\n")                      // two in a folder
	write(t, ws, "recipes/b.recipe.toml", "name = \"b\"\n")                      //
	write(t, ws, "apps/web/web.recipe.toml", "name = \"web\"\n")                 // beside another project's files
	write(t, ws, "apps/web/package.json", "{}")                                  //
	write(t, ws, "recipes/unnamed/unnamed.recipe.toml", "version = \"1.0.0\"\n") // no name
	write(t, ws, "recipes/broken/broken.recipe.toml", "name = \n")               // not TOML
	write(t, ws, "recipes/nested/x/x.recipe.toml", "name = \"x\"\n")             // fine
	write(t, ws, "recipes/nested/y.recipe.toml", "name = \"y\"\n")               // a folder holding a folder: not its own
	manifests := []string{"team.recipe.toml", "recipes/a.recipe.toml", "recipes/b.recipe.toml", "apps/web/web.recipe.toml", "recipes/unnamed/unnamed.recipe.toml", "recipes/broken/broken.recipe.toml", "recipes/nested/x/x.recipe.toml", "recipes/nested/y.recipe.toml"}
	got, err := Discover(ws, manifests)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Projects) != 1 || got.Projects[0].Name != "x" {
		t.Errorf("only the recipe in a folder of its own is a project: %+v", got.Projects)
	}
}
