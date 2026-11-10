package releasing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSetVersion(t *testing.T) {
	if out, ok := SetVersion("{\n    \"name\": \"acme/core\",\n    \"version\": \"0.0.0\"\n}\n", "1.2.0"); !ok || out != "{\n    \"name\": \"acme/core\",\n    \"version\": \"1.2.0\"\n}\n" {
		t.Fatalf("%q", out)
	}
	if out, ok := SetVersion("{\n    \"name\": \"acme/core\",\n    \"type\": \"library\"\n}\n", "1.2.0"); !ok || out != "{\n    \"name\": \"acme/core\",\n    \"version\": \"1.2.0\",\n    \"type\": \"library\"\n}\n" {
		t.Fatalf("added %q", out)
	}
}

func TestAPrereleaseIsWrittenAsSemverAndPublishedByItsTag(t *testing.T) {
	if out, ok := SetVersion("{\n    \"name\": \"acme/core\",\n    \"version\": \"1.0.0\"\n}\n", "2.0.0-beta.3"); !ok || out != "{\n    \"name\": \"acme/core\",\n    \"version\": \"2.0.0-beta.3\"\n}\n" {
		t.Fatalf("Composer reads the semver prerelease as beta stability: %q", out)
	}
	res, err := Publish(protocol.PublishParams{Project: protocol.Project{Kind: protocol.ProjectPackage}, Version: "2.0.0-beta.3", Tag: "beta"})
	if err != nil || res.Published || res.Skipped == "" {
		t.Fatal("Packagist takes the prerelease from the repository's tag, as it takes a release", err, res)
	}
}

func TestPublishIsTheTag(t *testing.T) {
	res, err := Publish(protocol.PublishParams{Project: protocol.Project{Kind: protocol.ProjectPackage}})
	if err != nil || res.Published || res.Skipped == "" {
		t.Fatal(err, res)
	}
}

func TestPublished(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/p2/acme/core.json" {
			_, _ = w.Write([]byte(`{"packages": {"acme/core": [{"version": "v2.0.0-RC1"}, {"version": "1.10.0"}, {"version": "v1.9.0"}, {"version": "dev-main"}]}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	Packagist = srv.URL
	core := protocol.Project{Kind: protocol.ProjectPackage, Exports: []string{"acme/core"}}
	if res, _ := Published(protocol.PublishedParams{Project: core}); res.Version != "1.10.0" {
		t.Fatalf("%+v", res)
	}
	core.Exports[0] = "acme/other"
	if res, _ := Published(protocol.PublishedParams{Project: core}); res.Version != "" || res.Unreachable {
		t.Fatalf("%+v", res)
	}
}
