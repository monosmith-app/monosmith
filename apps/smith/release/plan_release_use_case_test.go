package release

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func repo(t *testing.T) *gitclient.Git {
	t.Helper()
	g := gitclient.New(t.TempDir())
	if err := g.Init(); err != nil {
		t.Skip("git unavailable")
	}
	for _, kv := range [][2]string{{"user.email", "t@e.com"}, {"user.name", "T"}, {"commit.gpgsign", "false"}, {"tag.gpgsign", "false"}} {
		_, _ = g.Run("config", kv[0], kv[1])
	}
	return g
}

func commit(t *testing.T, g *gitclient.Git, file, message string) {
	t.Helper()
	p := filepath.Join(g.Dir, filepath.FromSlash(file))
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	old, _ := os.ReadFile(p)
	_ = os.WriteFile(p, append(old, 'x'), 0o644)
	if err := g.CommitAll(message); err != nil {
		t.Fatal(err)
	}
}

func packages(t *testing.T) *projectgraph.Graph {
	t.Helper()
	mk := func(name, root, prefix string, kind protocol.ProjectKind, deps ...string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name, Root: root, Kind: kind, Version: "0.0.0", TagPrefix: prefix}, Deps: deps, Plugin: "test"}
	}
	g, err := projectgraph.New([]*projectgraph.Project{
		mk("core", "packages/core", "", protocol.ProjectPackage),
		mk("sdk", "packages/sdk", "", protocol.ProjectPackage, "core"),
		mk("gomod", "packages/gomod", "packages/gomod/v", protocol.ProjectPackage),
		mk("web", "apps/web", "", protocol.ProjectApp, "core"),
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func byProject(plan Plan) map[string]Item {
	out := map[string]Item{}
	for _, i := range plan.Items {
		out[i.Project] = i
	}
	return out
}

func TestPlan(t *testing.T) {
	g := repo(t)
	pg := packages(t)
	commit(t, g, "packages/core/a", "feat(core): first feature")
	commit(t, g, "packages/gomod/a", "fix: go fix")
	commit(t, g, "packages/sdk/a", "docs: readme only")
	commit(t, g, "apps/web/a", "feat(web): apps never release")

	plan, err := MakePlan(g, pg)
	if err != nil {
		t.Fatal(err)
	}
	items := byProject(plan)
	if items["core"].Next != "0.1.0" || items["core"].Tag != "core@0.1.0" {
		t.Fatalf("core %+v", items["core"])
	}
	if items["sdk"].Next != "0.0.1" || items["sdk"].Reason != "dependency core 0.1.0" {
		t.Fatalf("sdk must follow its dependency: %+v", items["sdk"])
	}
	if items["gomod"].Tag != "packages/gomod/v0.0.1" {
		t.Fatalf("gomod %+v", items["gomod"])
	}
	if _, ok := items["web"]; ok {
		t.Fatal("apps are never released")
	}
	if plan.Items[0].Project == "sdk" {
		t.Fatal("core must release before sdk")
	}

	for _, i := range plan.Items {
		if err := g.CreateTag(i.Tag, i.Tag); err != nil {
			t.Fatal(err)
		}
	}
	plan, _ = MakePlan(g, pg)
	if len(plan.Items) != 0 || len(plan.Unchanged) != 3 {
		t.Fatalf("after tagging nothing releases: %+v", plan)
	}

	commit(t, g, "packages/sdk/b", "feat(sdk)!: breaking before 1.0")
	commit(t, g, "README.md", "version(core)[1.0.0]: first stable")
	plan, _ = MakePlan(g, pg)
	items = byProject(plan)
	if items["sdk"].Next != "0.1.0" || items["sdk"].Current != "0.0.1" {
		t.Fatalf("sdk %+v", items["sdk"])
	}
	if items["core"].Next != "1.0.0" || !items["core"].Forced {
		t.Fatalf("core %+v", items["core"])
	}
}

func TestPlanReleasesOptedInAppsFromTheirBundledLibs(t *testing.T) {
	g := repo(t)
	mk := func(name, root string, kind protocol.ProjectKind, deps ...string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name, Root: root, Kind: kind}, Deps: deps, Plugin: "test"}
	}
	cli := mk("cli", "apps/cli", protocol.ProjectApp, "engine")
	cli.Release = &workspace.ProjectRelease{}
	pg, err := projectgraph.New([]*projectgraph.Project{cli, mk("engine", "libs/engine", protocol.ProjectLib), mk("web", "apps/web", protocol.ProjectApp, "engine")})
	if err != nil {
		t.Fatal(err)
	}
	commit(t, g, "apps/cli/a", "feat: first")
	if err := g.CreateTag("cli@0.1.0", "cli 0.1.0"); err != nil {
		t.Fatal(err)
	}
	commit(t, g, "libs/engine/a", "fix(engine): a bug the app ships")

	plan, err := MakePlan(g, pg)
	if err != nil {
		t.Fatal(err)
	}
	items := byProject(plan)
	if items["cli"].Next != "0.1.1" || items["cli"].Current != "0.1.0" {
		t.Fatalf("a fix in a bundled lib must release the app: %+v", plan)
	}
	if _, ok := items["web"]; ok {
		t.Fatalf("an app without a release section never releases: %+v", plan)
	}
}
