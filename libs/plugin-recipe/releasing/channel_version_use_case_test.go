package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// A release channel stamps a prerelease version into the recipe and into the
// release's copy of it; nothing is sent to a registry.
func TestPublishCopiesTheRecipeWithAChannelPrerelease(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "recipes", "team")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "team.recipe.toml"), []byte("name = \"team\"\nversion = \"1.0.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	project := protocol.Project{Name: "team", Root: "recipes/team"}
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "2.0.0-beta.3"}); err != nil {
		t.Fatal(err)
	}
	res, err := Publish(ws, protocol.PublishParams{Project: project, Version: "2.0.0-beta.3", Tag: "beta"})
	if err != nil || res.Skipped == "" || len(res.Assets) != 1 {
		t.Fatalf("%+v %v", res, err)
	}
	got, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(res.Assets[0])))
	if err != nil || !strings.Contains(string(got), `version = "2.0.0-beta.3"`) {
		t.Errorf("%q %v", got, err)
	}
}
