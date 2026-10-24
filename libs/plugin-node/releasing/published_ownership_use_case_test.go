package releasing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPublishedDoesNotTrustABareNameOnThePublicRegistry(t *testing.T) {
	asked := false
	capture = func(_ string, argv []string) (string, string, error) {
		if argv[1] == "config" {
			return "https://registry.npmjs.org/\n", "", nil
		}
		asked = true
		return "0.0.2\n", "", nil
	}
	for _, name := range []string{"web", "@acme/web"} {
		asked = false
		got, err := Published(t.TempDir(), protocol.PublishedParams{Project: protocol.Project{Exports: []string{name}}})
		if err != nil {
			t.Fatal(err)
		}
		bare := name == "web"
		if bare && (asked || got.Version != "") {
			t.Errorf("%s: a stranger's package must not be reported as ours: %+v", name, got)
		}
		if !bare && got.Version != "0.0.2" {
			t.Errorf("%s: a scoped name is asked about: %+v", name, got)
		}
	}
}

func TestPublishedTrustsABareNameOnAPrivateRegistry(t *testing.T) {
	capture = func(_ string, argv []string) (string, string, error) {
		if argv[1] == "config" {
			return "https://feed.example/npm/\n", "", nil
		}
		return "1.2.3\n", "", nil
	}
	got, _ := Published(t.TempDir(), protocol.PublishedParams{Project: protocol.Project{Exports: []string{"web"}}})
	if got.Version != "1.2.3" {
		t.Errorf("a bare name on the team's own feed is theirs: %+v", got)
	}
}

func TestPublishedSkipsAPrivatePackage(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "pkg", "package.json"), []byte(`{"name":"@acme/internal","private":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	capture = func(_ string, argv []string) (string, string, error) {
		if argv[1] == "config" {
			return "https://registry.example/\n", "", nil
		}
		return "9.9.9\n", "", nil
	}
	got, _ := Published(ws, protocol.PublishedParams{Project: protocol.Project{Root: "pkg", Exports: []string{"@acme/internal"}}})
	if got.Version != "" {
		t.Errorf("a private package is never published: %+v", got)
	}
}

func TestLatestAsksAboutAnyDependencyEvenInsideAPrivateWorkspace(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"name":"demo","private":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	capture = func(_ string, argv []string) (string, string, error) {
		if argv[1] == "config" {
			return "https://registry.npmjs.org/\n", "", nil
		}
		return "3.0.1\n", "", nil
	}
	if got := Latest(ws, "is-odd"); got.Version != "3.0.1" {
		t.Errorf("a dependency's latest version is whoever's it is: %+v", got)
	}
}
