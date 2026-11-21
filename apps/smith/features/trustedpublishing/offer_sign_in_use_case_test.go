package trustedpublishing

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

// workspaceWith writes a smith.toml with body in a new folder and returns the
// folder and the loaded configuration.
func workspaceWith(t *testing.T, body string) (string, workspace.Config) {
	t.Helper()
	dir := t.TempDir()
	raw := "[workspace]\nname = \"w\"\n" + body
	if err := os.WriteFile(filepath.Join(dir, workspace.FileName), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := workspace.Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	return dir, cfg
}

func readConfig(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, workspace.FileName))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

var place = Place{Owner: "acme", Repository: "kit", Workflow: "ci.yml", Branches: []string{"main", "next"}}

func TestOfferAsksExplainsAndRecordsTrustedPublishing(t *testing.T) {
	dir, cfg := workspaceWith(t, "")
	var out bytes.Buffer
	recorded, err := Offer(Question{Root: dir, Config: &cfg, Secrets: []string{"NUGET_API_KEY"}, Place: place, Ask: true,
		In: bufio.NewReader(strings.NewReader("\n")), Out: &out})
	if err != nil || len(recorded) != 1 {
		t.Fatalf("got %v %v, want NuGet recorded", recorded, err)
	}
	for _, want := range []string{
		"NuGet: the pipeline publishes with the secret NUGET_API_KEY",
		"owner acme, repository kit, workflow file ci.yml, no environment",
		"secret NUGET_USER",
		"1) trusted publishing", "recommended",
		"2) API key secret", "keep the repository secret NUGET_API_KEY",
		"wrote smith.toml: [release] nuget_trusted_publishing = true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the output lacks %q:\n%s", want, out.String())
		}
	}
	if !strings.Contains(readConfig(t, dir), "nuget_trusted_publishing = true") {
		t.Errorf("smith.toml does not record the answer:\n%s", readConfig(t, dir))
	}
	if !cfg.Release.NuGetTrustedOn() {
		t.Error("the configuration in memory does not take the answer")
	}
}

func TestOfferRecordsTheKeyAsFalseSoItDoesNotAskAgain(t *testing.T) {
	dir, cfg := workspaceWith(t, "")
	var out bytes.Buffer
	if _, err := Offer(Question{Root: dir, Config: &cfg, Secrets: []string{"GEM_HOST_API_KEY"}, Place: place, Ask: true,
		In: bufio.NewReader(strings.NewReader("2\n")), Out: &out}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readConfig(t, dir), "rubygems_trusted_publishing = false") {
		t.Fatalf("smith.toml does not record the no:\n%s", readConfig(t, dir))
	}
	if cfg.Release.RubyGemsTrustedOn() || !Options[1].Decided(cfg.Release) {
		t.Error("the no is not recorded as a decided false")
	}
	out.Reset()
	again, err := Offer(Question{Root: dir, Config: &cfg, Secrets: []string{"GEM_HOST_API_KEY"}, Place: place, Ask: true,
		In: bufio.NewReader(strings.NewReader("1\n")), Out: &out})
	if err != nil || len(again) != 0 || out.Len() != 0 {
		t.Errorf("a decided option was asked again: %v %v\n%s", again, err, out.String())
	}
}

func TestOfferWithoutAPersonOnlyNotes(t *testing.T) {
	dir, cfg := workspaceWith(t, "")
	before := readConfig(t, dir)
	var out bytes.Buffer
	recorded, err := Offer(Question{Root: dir, Config: &cfg, Secrets: []string{"VSCE_PAT"}, Place: place, Out: &out})
	if err != nil || len(recorded) != 0 {
		t.Fatalf("got %v %v, want nothing recorded", recorded, err)
	}
	if readConfig(t, dir) != before || cfg.Release.VSCEEntra != nil {
		t.Error("a run without a person changed the configuration")
	}
	want := "note: The VS Code Marketplace publishes with the secret VSCE_PAT.\n→ smith config set release.vsce_entra true signs in with Microsoft Entra ID instead, false keeps the key and stops this note\n"
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

func TestOfferNamesTheEntraBranchesAndPlaceholders(t *testing.T) {
	steps := Options[2].Steps(Place{Workflow: "ci.yml", Branches: []string{"main", "next"}})
	if !strings.Contains(steps[0], "organisation <owner>, repository <repository>") || !strings.Contains(steps[0], "each of the branches main, next") {
		t.Errorf("got %q", steps[0])
	}
	if one := Options[2].Steps(Place{Workflow: "ci.yml", Branches: []string{"main"}}); !strings.Contains(one[0], "entity Branch, branch main.") {
		t.Errorf("got %q", one[0])
	}
}
