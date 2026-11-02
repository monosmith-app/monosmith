package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/npmpropagation"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func fakeNpm(t *testing.T, published string) *[]string {
	t.Helper()
	registries.forget()
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

func TestAReleaseChannelPublishesUnderItsDistTagAndLooksUpWithoutIt(t *testing.T) {
	calls := fakeNpm(t, "")
	res, err := Publish("/ws", protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk", Exports: []string{"@acme/sdk"}}, Version: "2.0.0-beta.1", Tag: "beta"})
	if err != nil || !res.Published {
		t.Fatalf("res %+v err %v", res, err)
	}
	want := []string{"npm view @acme/sdk@2.0.0-beta.1 version", "npm publish --access public --tag beta"}
	if got := (*calls)[len(*calls)-2:]; strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("calls:\n%s", strings.Join(*calls, "\n"))
	}
}

func TestPublishKeepsTheChannelTagWhenPlatformPackagesAreNotVisible(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "packages", "cli")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"@acme/cli","optionalDependencies":{"@acme/cli-linux":"2.0.0-beta.1"}}`), 0o644)
	real := awaitPlatforms
	defer func() { awaitPlatforms = real }()
	for _, c := range []struct {
		name    string
		ready   bool
		wantTag string
		note    bool
	}{{"visible", true, "--tag next", false}, {"late", false, "--tag pending", true}} {
		t.Run(c.name, func(t *testing.T) {
			calls := fakeNpm(t, "")
			awaitPlatforms = func(_ string, _ []string, _ time.Duration, _ []string, _ string) npmpropagation.Outcome {
				if c.ready {
					return npmpropagation.Outcome{Ready: true}
				}
				return npmpropagation.Outcome{Missing: []string{"@acme/cli-linux"}, Waited: time.Minute}
			}
			res, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Root: "packages/cli", Exports: []string{"@acme/cli"}}, Version: "2.0.0-beta.1", Tag: "next"})
			if err != nil || !res.Published {
				t.Fatalf("res %+v err %v", res, err)
			}
			if last := (*calls)[len(*calls)-1]; !strings.Contains(last, c.wantTag) {
				t.Errorf("last call %q, want %q", last, c.wantTag)
			}
			if len(res.Notes) != 0 {
				t.Errorf("notes %v: the note comes with the final phase", res.Notes)
			}
			if got := len(res.PendingMoves) == 1 && res.PendingMoves[0].Command() == "npm dist-tag add @acme/cli@2.0.0-beta.1 next" && res.PendingMoves[0].Awaits[0] == "@acme/cli-linux"; got != c.note {
				t.Errorf("pending moves %v: the host needs what to await and the tag", res.PendingMoves)
			}
		})
	}
}

func TestAPublishNeverWaitsAndAStableReleaseTakesThePendingTagToo(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "packages", "cli")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"@acme/cli","optionalDependencies":{"@acme/cli-linux":"2.0.0"}}`), 0o644)
	real := awaitPlatforms
	defer func() { awaitPlatforms = real }()
	waited := time.Hour
	awaitPlatforms = func(_ string, _ []string, wait time.Duration, _ []string, _ string) npmpropagation.Outcome {
		waited = wait
		return npmpropagation.Outcome{Missing: []string{"@acme/cli-linux"}}
	}
	calls := fakeNpm(t, "")
	res, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Root: "packages/cli", Exports: []string{"@acme/cli"}}, Version: "2.0.0", PropagationWait: "90s"})
	if err != nil || !res.Published {
		t.Fatalf("res %+v err %v", res, err)
	}
	if last := (*calls)[len(*calls)-1]; !strings.Contains(last, "--tag pending") {
		t.Errorf("last call %q: latest must not move", last)
	}
	if waited != 0 || res.PendingMoves[0].Command() != "npm dist-tag add @acme/cli@2.0.0 latest" {
		t.Errorf("waited %v, pending %v: a publish probes once and returns the move", waited, res.PendingMoves)
	}
}

func TestFinalizeWaitsThenMovesTheTagAndLeavesTheLateOnesPending(t *testing.T) {
	real := awaitPlatforms
	defer func() { awaitPlatforms = real }()
	var waits []time.Duration
	awaitPlatforms = func(_ string, _ []string, wait time.Duration, names []string, _ string) npmpropagation.Outcome {
		waits = append(waits, wait)
		if names[0] == "@acme/late-linux" {
			return npmpropagation.Outcome{Missing: names, Waited: 15 * time.Minute}
		}
		return npmpropagation.Outcome{Ready: true}
	}
	calls := fakeNpm(t, "")
	moves := []protocol.TagMove{
		{Package: "@acme/cli", Version: "2.0.0-beta.1", Tag: "next", Awaits: []string{"@acme/cli-linux"}},
		{Package: "@acme/late", Version: "2.0.0-beta.1", Tag: "next", Awaits: []string{"@acme/late-linux"}},
	}
	res, err := Publish("/ws", protocol.PublishParams{Project: protocol.Project{Root: "packages/cli"}, Finalize: moves, PropagationWait: "15m"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(*calls, "\n"); got != "npm dist-tag add @acme/cli@2.0.0-beta.1 next\nnpm dist-tag rm @acme/cli pending" {
		t.Errorf("calls: %s", got)
	}
	if len(res.Moved) != 1 || len(res.PendingMoves) != 1 || res.PendingMoves[0].Package != "@acme/late" || len(res.Notes) != 1 || !strings.Contains(res.Notes[0], "npm dist-tag add @acme/late@2.0.0-beta.1 next") {
		t.Errorf("result %+v", res)
	}
	if waits[0] != 15*time.Minute {
		t.Errorf("waits %v", waits)
	}
	if _, err := Publish("/ws", protocol.PublishParams{Finalize: moves, PropagationWait: "soon"}); err == nil {
		t.Error("a wait that is no duration must be refused")
	}
}
