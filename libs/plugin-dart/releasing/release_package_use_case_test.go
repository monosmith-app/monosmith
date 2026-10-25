package releasing

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSetVersion(t *testing.T) {
	if out, ok := SetVersion("name: acme_core\nversion: 0.0.0\n", "1.2.0"); !ok || out != "name: acme_core\nversion: 1.2.0\n" {
		t.Fatalf("%q", out)
	}
	if out, ok := SetVersion("name: acme_core\ndescription: core\n", "1.2.0"); !ok || out != "name: acme_core\nversion: 1.2.0\ndescription: core\n" {
		t.Fatalf("added %q", out)
	}
}

func writeCore(t *testing.T, pubspec string) (string, protocol.Project) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "packages", "core"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "packages", "core", "pubspec.yaml"), []byte(pubspec), 0o644)
	return ws, protocol.Project{Name: "core", Root: "packages/core", Kind: protocol.ProjectPackage, Exports: []string{"acme_core"}}
}

func TestPublish(t *testing.T) {
	var ran []string
	real := PubRunner
	defer func() { PubRunner = real }()
	PubRunner = func(_ string, argv []string) error { ran = append(ran, strings.Join(argv, " ")); return nil }
	ws, core := writeCore(t, "name: acme_core\nversion: 1.0.0\n")
	if res, err := Publish(ws, protocol.PublishParams{Project: core, Version: "1.0.0"}); err != nil || res.Published || res.Skipped == "" || res.Registry != PubDev {
		t.Fatal("pub.dev is published to only from its tagged run", err, res)
	}
	if res, err := Publish(ws, protocol.PublishParams{Project: core, Version: "1.0.0", Tagged: true}); err != nil || !res.Published || res.Registry != PubDev {
		t.Fatal(err, res)
	}
	t.Setenv("PUB_TOKEN", "t")
	ws, core = writeCore(t, "name: acme_core\nversion: 1.0.0\npublish_to: https://pub.example/\n")
	if _, err := Publish(ws, protocol.PublishParams{Project: core, Version: "1.0.0"}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(ran, "|") != "dart pub publish --force|dart pub token add https://pub.example --env-var PUB_TOKEN|dart pub publish --force" {
		t.Fatalf("%q", ran)
	}
}

func TestPublished(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/packages/acme_core" {
			_, _ = w.Write([]byte(`{"name": "acme_core", "latest": {"version": "1.4.0"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	PubDev = srv.URL
	ws, core := writeCore(t, "name: acme_core\n")
	if res, _ := Published(ws, protocol.PublishedParams{Project: core}); res.Version != "1.4.0" {
		t.Fatalf("%+v", res)
	}
	core.Exports[0] = "other"
	if res, _ := Published(ws, protocol.PublishedParams{Project: core}); res.Version != "" || res.Unreachable {
		t.Fatalf("%+v", res)
	}
}

func TestAPackageWithPublishToNoneIsNeverPlannedForAnyRegistry(t *testing.T) {
	var ran []string
	real := PubRunner
	defer func() { PubRunner = real }()
	PubRunner = func(_ string, argv []string) error { ran = append(ran, strings.Join(argv, " ")); return nil }
	ws, core := writeCore(t, "name: acme_core\nversion: 1.0.0\npublish_to: none\n")
	for _, p := range []protocol.PublishParams{{Project: core, Version: "1.0.0", DryRun: true}, {Project: core, Version: "1.0.0", Tagged: true}} {
		res, err := Publish(ws, p)
		if err != nil || res.Published || res.Registry != "" || res.Skipped == "" {
			t.Fatalf("publish_to: none was planned for a registry: %v %+v", err, res)
		}
	}
	if len(ran) != 0 {
		t.Errorf("pub ran: %q", ran)
	}
	if res, err := Published(ws, protocol.PublishedParams{Project: core}); err != nil || res.Registry != "" || res.Unreachable {
		t.Errorf("no registry is asked about it: %v %+v", err, res)
	}
}
