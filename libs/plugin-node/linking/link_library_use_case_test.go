package linking

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/dependencies"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAddDependencyKeepsTheFileAndSortsTheDependencies(t *testing.T) {
	src := "{\n  \"name\": \"web\",\n  \"version\": \"1.0.0\",\n  \"dependencies\": {\n    \"zod\": \"^3.0.0\",\n    \"react\": \"^18.0.0\"\n  },\n  \"scripts\": {\n    \"test\": \"vitest\"\n  }\n}\n"
	got, changed, err := AddDependency([]byte(src), "@acme/core", "*")
	if err != nil || !changed {
		t.Fatalf("changed %v err %v", changed, err)
	}
	want := "{\n  \"name\": \"web\",\n  \"version\": \"1.0.0\",\n  \"dependencies\": {\n    \"@acme/core\": \"*\",\n    \"react\": \"^18.0.0\",\n    \"zod\": \"^3.0.0\"\n  },\n  \"scripts\": {\n    \"test\": \"vitest\"\n  }\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	again, changed, err := AddDependency(got, "@acme/core", "*")
	if err != nil || changed || string(again) != want {
		t.Fatalf("a second add must change nothing: %v %v", changed, err)
	}
}

func TestAddDependencyCreatesTheSectionAndRefusesJunk(t *testing.T) {
	got, changed, err := AddDependency([]byte("{\n  \"name\": \"web\"\n}\n"), "core", "*")
	if err != nil || !changed || !strings.Contains(string(got), "\"dependencies\": {\n    \"core\": \"*\"\n  }") {
		t.Fatalf("got %s %v %v", got, changed, err)
	}
	if _, _, err := AddDependency([]byte("not json"), "core", "*"); err == nil {
		t.Fatal("a manifest that is not JSON must be refused")
	}
}

func TestRemoveDependencyDropsAnEmptiedSection(t *testing.T) {
	src := "{\n  \"name\": \"web\",\n  \"dependencies\": {\n    \"core\": \"*\"\n  },\n  \"devDependencies\": {\n    \"core\": \"*\",\n    \"vitest\": \"1\"\n  }\n}\n"
	got, from, err := RemoveDependency([]byte(src), "core")
	if err != nil || strings.Join(from, ",") != "dependencies,devDependencies" {
		t.Fatalf("from %v err %v", from, err)
	}
	want := "{\n  \"name\": \"web\",\n  \"devDependencies\": {\n    \"vitest\": \"1\"\n  }\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s", got)
	}
	same, from, _ := RemoveDependency(got, "core")
	if len(from) != 0 || string(same) != string(got) {
		t.Fatal("removing what is not there changes nothing")
	}
}

func workspace(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func nodeProject(name, root string, exports ...string) protocol.Project {
	return protocol.Project{Name: name, Root: root, Language: "node", Exports: exports, Version: "1.2.0"}
}

func TestPlanOffersTheRealChoiceWhenTheWorkspaceHasNoHabit(t *testing.T) {
	root := workspace(t, map[string]string{
		"package.json":           `{"name":"root","workspaces":["apps/*","libs/*"]}`,
		"apps/web/package.json":  `{"name":"web","dependencies":{}}`,
		"libs/core/package.json": `{"name":"@acme/core","version":"1.2.0"}`,
	})
	res, err := Link(root, protocol.LinkParams{Plan: true, Project: nodeProject("web", "apps/web", "web"), Library: nodeProject("core", "libs/core", "@acme/core")})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Options) != 2 || res.Options[0].Spec != "*" || res.Options[1].Spec != "^1.2.0" {
		t.Fatalf("options %+v", res.Options)
	}
	if res.Manifest != "apps/web/package.json" || res.Line != `"@acme/core": "*"` {
		t.Fatalf("plan %+v", res)
	}
	if raw, _ := os.ReadFile(filepath.Join(root, "apps/web/package.json")); strings.Contains(string(raw), "acme") {
		t.Fatal("a plan writes nothing")
	}
}

func TestPlanFollowsTheHabitOfTheWorkspace(t *testing.T) {
	root := workspace(t, map[string]string{
		"package.json":           `{"name":"root","workspaces":["apps/*","libs/*"]}`,
		"apps/web/package.json":  `{"name":"web"}`,
		"apps/api/package.json":  `{"name":"api","dependencies":{"@acme/core":"^1.0.0"}}`,
		"libs/core/package.json": `{"name":"@acme/core","version":"1.2.0"}`,
		"libs/ui/package.json":   `{"name":"@acme/ui","version":"1.0.0"}`,
	})
	res, err := Link(root, protocol.LinkParams{Plan: true, Project: nodeProject("web", "apps/web", "web"), Library: nodeProject("ui", "libs/ui", "@acme/ui")})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Options) != 0 || res.Line != `"@acme/ui": "^1.0.0"` {
		t.Fatalf("the habit must be followed without asking: %+v", res)
	}
}

func TestLinkWritesTheDependencyAndInstalls(t *testing.T) {
	root := workspace(t, map[string]string{
		"package.json":           `{"name":"root","workspaces":["apps/*","libs/*"]}`,
		"apps/web/package.json":  "{\n  \"name\": \"web\"\n}\n",
		"libs/core/package.json": `{"name":"@acme/core","version":"1.2.0"}`,
	})
	var ran [][]string
	saved := dependencies.FoldedNpmRunner
	dependencies.FoldedNpmRunner = func(_ string, argv []string) ([]string, error) { ran = append(ran, argv); return nil, nil }
	defer func() { dependencies.FoldedNpmRunner = saved }()

	res, err := Link(root, protocol.LinkParams{Spec: "^1.2.0", Project: nodeProject("web", "apps/web", "web"), Library: nodeProject("core", "libs/core", "@acme/core")})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "apps/web/package.json"))
	if !strings.Contains(string(raw), `"@acme/core": "^1.2.0"`) {
		t.Fatalf("manifest:\n%s", raw)
	}
	if len(ran) != 1 || ran[0][1] != "install" || strings.Join(res.Files, ",") != "apps/web/package.json,package-lock.json" {
		t.Fatalf("ran %v files %v", ran, res.Files)
	}

	un, err := Unlink(root, protocol.UnlinkParams{Project: nodeProject("web", "apps/web", "web"), Library: nodeProject("core", "libs/core", "@acme/core")})
	if err != nil || len(un.Files) == 0 {
		t.Fatalf("unlink %+v %v", un, err)
	}
	un, err = Unlink(root, protocol.UnlinkParams{Project: nodeProject("web", "apps/web", "web"), Library: nodeProject("core", "libs/core", "@acme/core")})
	if err != nil || len(un.Files) != 0 {
		t.Fatalf("a second unlink must report nothing linked: %+v %v", un, err)
	}
}

func TestLinkNamesTheMissingManifest(t *testing.T) {
	root := t.TempDir()
	_, err := Link(root, protocol.LinkParams{Project: nodeProject("gone", "apps/gone"), Library: nodeProject("core", "libs/core", "core")})
	if err == nil || !strings.Contains(err.Error(), "apps/gone/package.json") {
		t.Fatalf("got %v", err)
	}
}
