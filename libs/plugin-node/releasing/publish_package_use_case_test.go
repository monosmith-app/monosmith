package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func fakeNpm(t *testing.T, published string) *[]string {
	t.Helper()
	var calls []string
	npm.output = func(_ string, argv []string) (string, error) {
		calls = append(calls, strings.Join(argv, " "))
		if argv[1] == "view" {
			return published, nil
		}
		return "https://registry.example/\n", nil
	}
	npm.run = func(_ string, argv []string) error { calls = append(calls, strings.Join(argv, " ")); return nil }
	return &calls
}

func TestPublishSkipsExistingVersion(t *testing.T) {
	calls := fakeNpm(t, "1.0.0\n")
	res, err := Publish("/ws", protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk", Exports: []string{"@acme/sdk"}}, Version: "1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Published || res.Skipped != "already published" || res.Registry != "https://registry.example/" {
		t.Fatalf("result %+v", res)
	}
	for _, c := range *calls {
		if strings.Contains(c, "publish") {
			t.Fatal("must not publish an existing version")
		}
	}
}

func TestPublishNewVersion(t *testing.T) {
	calls := fakeNpm(t, "")
	res, err := Publish("/ws", protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk", Exports: []string{"@acme/sdk"}}, Version: "1.1.0"})
	if err != nil || !res.Published {
		t.Fatalf("res %+v err %v", res, err)
	}
	if last := (*calls)[len(*calls)-1]; last != "npm publish --access public" {
		t.Fatalf("last call %q", last)
	}
}

func TestPublishGoesToThePrivateRegistrySmithTomlNames(t *testing.T) {
	calls := fakeNpm(t, "")
	feed := "https://pkgs.dev.azure.com/acme/_packaging/libs/npm/registry/"
	project := protocol.Project{Root: "packages/sdk", Exports: []string{"@acme/sdk"}, Registries: &protocol.Registries{Publish: &protocol.Registry{Name: "libs", Auth: protocol.AuthAzureArtifacts, NPM: feed}}}
	res, err := Publish("/ws", protocol.PublishParams{Project: project, Version: "1.1.0"})
	if err != nil || !res.Published || res.Registry != feed {
		t.Fatalf("res %+v err %v", res, err)
	}
	want := []string{"npm view @acme/sdk@1.1.0 version --registry " + feed, "npm publish --access public --registry " + feed}
	if strings.Join(*calls, "\n") != strings.Join(want, "\n") {
		t.Errorf("calls:\n%s", strings.Join(*calls, "\n"))
	}
}

func TestPublishNeverSendsAnAppToNpm(t *testing.T) {
	calls := fakeNpm(t, "")
	res, err := Publish("/ws", protocol.PublishParams{Project: protocol.Project{Root: "apps/driver", Kind: protocol.ProjectApp, Exports: []string{"@acme/driver"}}, Version: "1.0.0"})
	if err != nil || res.Published || res.Skipped != "released by its tag" || len(*calls) != 0 {
		t.Fatalf("an app releases by its tag and assets, not npm: %+v %v %v", res, *calls, err)
	}

	// An app whose manifest is not private looks meant for npm: the skip says how to get there.
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "cli"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "cli", "package.json"), []byte(`{"name": "acme-cli", "bin": "main.js"}`), 0o644)
	res, err = Publish(ws, protocol.PublishParams{Project: protocol.Project{Root: "apps/cli", Kind: protocol.ProjectApp, Exports: []string{"acme-cli"}}, Version: "1.0.0"})
	if err != nil || res.Published || !strings.Contains(res.Skipped, `"kind": "package"`) || len(*calls) != 0 {
		t.Errorf("a public app: %+v %v", res, err)
	}
}

func TestPublishDryRunTouchesNothing(t *testing.T) {
	calls := fakeNpm(t, "")
	res, err := Publish("/ws", protocol.PublishParams{Project: protocol.Project{Root: "p", Exports: []string{"x"}}, Version: "1.0.0", DryRun: true})
	if err != nil || res.Skipped != "dry run" || len(*calls) != 1 {
		t.Fatalf("res %+v calls %v err %v", res, *calls, err)
	}
}

func TestAnExtensionAppBuildsNoArchives(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "apps", "ext"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "apps", "ext", "package.json"), []byte(`{"name":"ext","engines":{"vscode":"^1.90.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Name: "ext", Root: "apps/ext", Kind: protocol.ProjectApp}, Version: "1.0.0", Assets: &protocol.AssetSpec{Platforms: []string{"linux/amd64"}}})
	if err != nil || len(res.Assets) != 0 || !strings.Contains(res.Skipped, "vscode plugin") {
		t.Fatalf("result %+v: %v", res, err)
	}
}
