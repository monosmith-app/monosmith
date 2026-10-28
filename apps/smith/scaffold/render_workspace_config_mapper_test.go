package scaffold

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func render(t *testing.T, opts NewOptions) (string, workspace.Config) {
	t.Helper()
	opts.Name, opts.Scope, opts.ModulePrefix = "acme", "@acme", "github.com/acme/acme"
	text, err := renderWorkspaceConfig(opts, `"github"`)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := workspace.Parse([]byte(text))
	if err != nil {
		t.Fatalf("%v\n%s", err, text)
	}
	return text, cfg
}

func TestNoAnswersWriteTheTemplateAsItWas(t *testing.T) {
	text, _ := render(t, NewOptions{})
	if want := fmt.Sprintf(configTemplate, "acme", "@acme", "github.com/acme/acme", `"github"`); text != want {
		t.Fatalf("defaults changed the file:\n%s", text)
	}
}

func TestAnswersReplaceTheCommentedDefaults(t *testing.T) {
	text, cfg := render(t, NewOptions{
		DefaultStacks: map[string]string{"node": "node-vitest"}, Architecture: "hexagonal",
		ReleaseBranch: "trunk", ReleaseRemote: "upstream", GitHubReleases: true,
		CacheDir: ".cache/smith", CacheMaxMB: 512,
	})
	if cfg.DefaultStacks["node"] != "node-vitest" {
		t.Errorf("stacks %v", cfg.DefaultStacks)
	}
	if _, ok := cfg.Conventions["hexagonal"]; !ok || len(cfg.Conventions) != 1 {
		t.Errorf("conventions %v", cfg.Conventions)
	}
	if r := cfg.Release; r.Branch != "trunk" || r.Remote != "upstream" || !r.GitHubReleases {
		t.Errorf("release %+v", r)
	}
	if c := cfg.Cache; c.Dir != ".cache/smith" || c.MaxSizeMB != 512 {
		t.Errorf("cache %+v", c)
	}
	if strings.Contains(text, "# [release]") || strings.Contains(text, "# [cache]") {
		t.Errorf("the commented blocks stay beside the real ones:\n%s", text)
	}
}

func TestNoLayoutAndNoCacheLimit(t *testing.T) {
	_, cfg := render(t, NewOptions{Architecture: "none", CacheMaxMB: -1})
	if len(cfg.Conventions) != 0 {
		t.Errorf("none turns no convention on: %v", cfg.Conventions)
	}
	if cfg.Cache.MaxSizeMB != -1 {
		t.Errorf("0 in smith.toml loads as no limit: %d", cfg.Cache.MaxSizeMB)
	}
}

func TestAnAnswerSmithCannotWriteIsRefused(t *testing.T) {
	for name, opts := range map[string]NewOptions{
		"layout":        {Architecture: "spaghetti"},
		"stack":         {DefaultStacks: map[string]string{"node": "node-cobol"}},
		"stack of a go": {DefaultStacks: map[string]string{"go": "go-x"}},
	} {
		opts.Name = "acme"
		if _, err := renderWorkspaceConfig(opts, `"github"`); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
