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
	if out, ok := SetVersion("  spec.version = '0.0.0'\n", "1.2.0"); !ok || out != "  spec.version = '1.2.0'\n" {
		t.Fatalf("%q", out)
	}
	if out, ok := SetVersion(`  s.version = "0.0.0"`, "1.2.0"); !ok || out != `  s.version = "1.2.0"` {
		t.Fatalf("%q", out)
	}
	if _, ok := SetVersion("  spec.version = AcmeCore::VERSION\n", "1.2.0"); ok {
		t.Fatal("a computed version is left alone")
	}
}

func TestAPrereleaseIsSpelledTheRubyGemsWay(t *testing.T) {
	for in, want := range map[string]string{"2.0.0-beta.1": "2.0.0.pre.beta.1", "1.2.0": "1.2.0", "2.0.0-rc.2+build.5": "2.0.0.pre.rc.2"} {
		if got := GemVersion(in); got != want {
			t.Errorf("GemVersion(%q) = %q, want %q", in, got, want)
		}
	}
	if out, ok := SetVersion("  spec.version = '1.0.0'\n", "2.0.0-beta.1"); !ok || out != "  spec.version = '2.0.0.pre.beta.1'\n" {
		t.Fatalf("the gemspec carries the gem spelling: %q", out)
	}
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "packages", "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "packages", "core", "acme_core.gemspec"), []byte("spec.version = '1.0.0'"), 0o644); err != nil {
		t.Fatal(err)
	}
	var ran []string
	real := GemRunner
	defer func() { GemRunner = real }()
	GemRunner = func(_ string, argv []string) error { ran = append(ran, strings.Join(argv, " ")); return nil }
	core := protocol.Project{Name: "core", Root: "packages/core", Kind: protocol.ProjectPackage, Exports: []string{"acme_core"}}
	if res, err := Publish(ws, protocol.PublishParams{Project: core, Version: "2.0.0-beta.1", Tag: "beta"}); err != nil || !res.Published {
		t.Fatal(err, res)
	}
	if !strings.Contains(filepath.ToSlash(ran[0]), "dist/packages/core/acme_core-2.0.0.pre.beta.1.gem") || !strings.Contains(filepath.ToSlash(ran[1]), "acme_core-2.0.0.pre.beta.1.gem") {
		t.Fatalf("the packed file and the pushed file have the gem spelling: %q", ran)
	}
}

func TestPublish(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "packages", "core"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "packages", "core", "acme_core.gemspec"), []byte("spec.version = '0.0.0'"), 0o644)
	var ran []string
	real := GemRunner
	defer func() { GemRunner = real }()
	GemRunner = func(_ string, argv []string) error { ran = append(ran, strings.Join(argv, " ")); return nil }
	core := protocol.Project{Name: "core", Root: "packages/core", Kind: protocol.ProjectPackage, Exports: []string{"acme_core"}}
	res, err := Publish(ws, protocol.PublishParams{Project: core, Version: "1.0.0"})
	if err != nil || !res.Published || len(ran) != 2 {
		t.Fatal(err, res, ran)
	}
	if !strings.HasPrefix(ran[0], "gem build acme_core.gemspec --output ") || !strings.HasSuffix(filepath.ToSlash(ran[0]), "dist/packages/core/acme_core-1.0.0.gem") || !strings.HasSuffix(ran[1], " --host https://rubygems.org") {
		t.Fatalf("%q", ran)
	}
	if _, err := os.Stat(filepath.Join(ws, "dist", "packages", "core")); err != nil {
		t.Fatal("the gem is packed into dist/")
	}
}

func TestPublished(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/versions/acme_core/latest.json" {
			_, _ = w.Write([]byte(`{"version": "1.4.0"}`))
			return
		}
		_, _ = w.Write([]byte(`{"version": "unknown"}`))
	}))
	defer srv.Close()
	RubyGems = srv.URL
	core := protocol.Project{Kind: protocol.ProjectPackage, Exports: []string{"acme_core"}}
	if res, _ := Published(protocol.PublishedParams{Project: core}); res.Version != "1.4.0" {
		t.Fatalf("%+v", res)
	}
	core.Exports[0] = "other"
	if res, _ := Published(protocol.PublishedParams{Project: core}); res.Version != "" || res.Unreachable {
		t.Fatalf("%+v", res)
	}
}
