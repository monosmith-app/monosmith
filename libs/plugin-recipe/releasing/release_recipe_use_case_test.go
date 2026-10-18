package releasing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSetVersion(t *testing.T) {
	for name, tc := range map[string]struct{ in, want string }{
		"replaces the version": {"name = \"team\"\nversion = \"1.2.0\"\n\n[choices.x]\nversion = \"keep\"\n", "name = \"team\"\nversion = \"1.3.0\"\n\n[choices.x]\nversion = \"keep\"\n"},
		"keeps a comment":      {"name = \"team\"\nversion = \"1.2.0\" # released\n", "name = \"team\"\nversion = \"1.3.0\" # released\n"},
		"adds it after name":   {"name = \"team\"\nextends = \"b.toml\"\n", "name = \"team\"\nversion = \"1.3.0\"\nextends = \"b.toml\"\n"},
		"adds it at the top":   {"# a recipe\n\n[use]\nlanguage = \"go\"\n", "version = \"1.3.0\"\n# a recipe\n\n[use]\nlanguage = \"go\"\n"},
		"single quotes":        {"name = 'team'\nversion = '1.2.0'\n", "name = 'team'\nversion = \"1.3.0\"\n"},
		"windows line ends":    {"name = \"team\"\r\nextends = \"b.toml\"\r\n", "name = \"team\"\r\nversion = \"1.3.0\"\r\nextends = \"b.toml\"\r\n"},
	} {
		if got, changed := SetVersion(tc.in, "1.3.0"); got != tc.want || !changed {
			t.Errorf("%s: got %q (changed %v), want %q", name, got, changed, tc.want)
		}
	}
	if _, changed := SetVersion("name = \"t\"\nversion = \"1.3.0\"\n", "1.3.0"); changed {
		t.Error("the same version changed the file")
	}
}

func recipeWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	dir := filepath.Join(ws, "recipes", "team")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "team.recipe.toml"), []byte("name = \"team\"\nversion = \"1.2.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return ws
}

func TestWriteVersionWritesTheRecipeFile(t *testing.T) {
	ws := recipeWorkspace(t)
	project := protocol.Project{Name: "team", Root: "recipes/team"}
	dry, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "1.3.0", DryRun: true})
	if err != nil || len(dry.Files) != 1 || dry.Files[0] != "recipes/team/team.recipe.toml" {
		t.Fatalf("dry run: %v %+v", err, dry)
	}
	if src, _ := os.ReadFile(filepath.Join(ws, "recipes", "team", "team.recipe.toml")); string(src) != "name = \"team\"\nversion = \"1.2.0\"\n" {
		t.Fatalf("a dry run wrote the file: %q", src)
	}
	res, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "1.3.0"})
	if err != nil || len(res.Files) != 1 {
		t.Fatalf("%v %+v", err, res)
	}
	if src, _ := os.ReadFile(filepath.Join(ws, "recipes", "team", "team.recipe.toml")); string(src) != "name = \"team\"\nversion = \"1.3.0\"\n" {
		t.Errorf("file = %q", src)
	}
	if again, _ := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "1.3.0"}); len(again.Files) != 0 {
		t.Errorf("the same version again changed %v", again.Files)
	}
}

func TestPublishAttachesTheRecipeStampedWithTheReleasesVersion(t *testing.T) {
	ws := recipeWorkspace(t)
	project := protocol.Project{Name: "team", Root: "recipes/team"}
	asset := filepath.Join(ws, "dist", "release", "team", "team.recipe.toml")
	dry, err := Publish(ws, protocol.PublishParams{Project: project, Version: "1.3.0", DryRun: true})
	if err != nil || dry.Skipped != "released by its tag" || len(dry.Assets) != 1 || dry.Assets[0] != "dist/release/team/team.recipe.toml" {
		t.Fatalf("dry run: %v %+v", err, dry)
	}
	if _, err := os.Stat(asset); err == nil {
		t.Error("a dry run built the asset")
	}
	res, err := Publish(ws, protocol.PublishParams{Project: project, Version: "1.3.0"})
	if err != nil || len(res.Assets) != 1 || res.Assets[0] != dry.Assets[0] {
		t.Fatalf("%v %+v", err, res)
	}
	if got, _ := os.ReadFile(asset); string(got) != "name = \"team\"\nversion = \"1.3.0\"\n" {
		t.Errorf("asset = %q; the copy says which release it is", got)
	}
	if src, _ := os.ReadFile(filepath.Join(ws, "recipes", "team", "team.recipe.toml")); string(src) != "name = \"team\"\nversion = \"1.2.0\"\n" {
		t.Errorf("the recipe in the repository changed: %q", src)
	}
	if _, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Name: "x", Root: "recipes/missing"}}); err == nil {
		t.Error("a folder with no recipe was published")
	}
}
