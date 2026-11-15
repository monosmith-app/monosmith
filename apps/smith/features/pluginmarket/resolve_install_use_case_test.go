package pluginmarket

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestResolverTurnsANameIntoAnArchiveAndItsDigest(t *testing.T) {
	t.Chdir(t.TempDir())
	srv := indexServer(t)
	resolve := NewResolver(http.DefaultClient, srv.URL, nil)

	r, ok, err := resolve(context.Background(), "widget@1.0.0")
	if err != nil || !ok {
		t.Fatalf("%v %v", ok, err)
	}
	if r.Name != "widget@1.0.0" || r.Source != "https://example.com/widget-1.0.0.smith-plugin.zip" || r.SHA256 != digestA || r.Origin != srv.URL {
		t.Fatalf("got %+v", r)
	}
	if r, _, _ := resolve(context.Background(), "widget"); r.Name != "widget@1.10.0" {
		t.Fatalf("the newest was not chosen: %+v", r)
	}
}

func TestResolverLeavesSourcesAloneAndNeverReadsTheIndexForThem(t *testing.T) {
	t.Chdir(t.TempDir())
	resolve := NewResolver(http.DefaultClient, "http://127.0.0.1:1/unreachable", nil)
	for _, arg := range []string{"./widget", "github:acme/widget@v1", "https://example.com/w.smith-plugin.zip", "w.zip"} {
		if _, ok, err := resolve(context.Background(), arg); ok || err != nil {
			t.Errorf("%s: %v %v", arg, ok, err)
		}
	}
}

func TestResolverErrorsPointToSearchAndToTheOtherSources(t *testing.T) {
	t.Chdir(t.TempDir())
	srv := indexServer(t)
	_, _, err := NewResolver(http.DefaultClient, srv.URL, nil)(context.Background(), "gadget")
	if err == nil || !strings.Contains(err.Error(), "smith plugin search") || !strings.Contains(err.Error(), "github:owner/repo@tag") {
		t.Fatalf("got %v", err)
	}
}
