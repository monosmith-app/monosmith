package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func webApp(settings webenv.Settings) *projectgraph.Project {
	return &projectgraph.Project{Project: protocol.Project{Name: "web", Root: "apps/web", Settings: map[string]map[string]any{webenv.PluginName: settings.ToMap()}}}
}

func writeEnvFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		path := filepath.Join(root, "apps", "web", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEveryEnvironmentNeedsItsEnvFile(t *testing.T) {
	root := t.TempDir()
	settings := webenv.Settings{Names: []string{"dev", "uat", "prd"}}
	writeEnvFiles(t, root, map[string]string{".env.dev": "VITE_API_URL=a\n", ".env.prd": "VITE_API_URL=b\n"})
	got := webEnvironmentFindings(root, []*projectgraph.Project{webApp(settings)})
	if len(got) != 2 || got[0].OK || !strings.Contains(got[0].Detail, "apps/web/.env.uat") || strings.Contains(got[0].Detail, ".env.dev") {
		t.Fatalf("want .env.uat missing, got %+v", got)
	}
	if got[0].Fix == nil || got[0].Fix.Files[0].Path != "apps/web/.env.uat" || !strings.Contains(got[0].Fix.Files[0].Content, "never put a secret here") {
		t.Errorf("the fix writes the starter file: %+v", got[0].Fix)
	}
	if !got[1].OK {
		t.Errorf("no secret-looking name: %+v", got[1])
	}
}

func TestAnExampleOnlyAppNeedsTheExampleNotEachFile(t *testing.T) {
	root := t.TempDir()
	settings := webenv.Settings{Names: []string{"dev", "prd"}, EnvFiles: webenv.FilesExample}
	writeEnvFiles(t, root, map[string]string{".env.example": "VITE_API_URL=\n"})
	if got := webEnvironmentFindings(root, []*projectgraph.Project{webApp(settings)}); !got[0].OK {
		t.Errorf("a fresh clone has no .env.<env>: %+v", got[0])
	}
}

func TestAViteNameWithKeySecretOrTokenIsWarnedAbout(t *testing.T) {
	root := t.TempDir()
	settings := webenv.Settings{Names: []string{"prd"}}
	writeEnvFiles(t, root, map[string]string{".env.prd": "# VITE_COMMENT_TOKEN=x\nVITE_API_URL=a\nVITE_STRIPE_SECRET_KEY=sk\nexport VITE_AUTH_TOKEN=t\nAPI_TOKEN=notvite\n"})
	got := webEnvironmentFindings(root, []*projectgraph.Project{webApp(settings)})
	if got[1].OK || !got[1].Warning {
		t.Fatalf("want a warning, got %+v", got[1])
	}
	for _, want := range []string{"VITE_STRIPE_SECRET_KEY in apps/web/.env.prd", "VITE_AUTH_TOKEN in apps/web/.env.prd"} {
		if !strings.Contains(got[1].Detail, want) {
			t.Errorf("misses %q: %s", want, got[1].Detail)
		}
	}
	if strings.Contains(got[1].Detail, "COMMENT") || strings.Contains(got[1].Detail, "API_TOKEN in") {
		t.Errorf("a comment and a name without the prefix are not reported: %s", got[1].Detail)
	}
}

func TestAProjectWithoutEnvironmentsHasNoCheck(t *testing.T) {
	p := &projectgraph.Project{Project: protocol.Project{Name: "api", Root: "apps/api"}}
	if got := webEnvironmentFindings(t.TempDir(), []*projectgraph.Project{p}); got != nil {
		t.Errorf("got %+v", got)
	}
}
