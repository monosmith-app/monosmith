package pluginmarket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// asGitHub makes the test server count as a GitHub host for the test.
func asGitHub(t *testing.T, srv *httptest.Server) {
	t.Helper()
	u, _ := url.Parse(srv.URL)
	githubHosts[u.Hostname()] = true
	t.Cleanup(func() { delete(githubHosts, u.Hostname()) })
}

func TestFetchIndexReadsAPublicIndexWithoutAToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("a public index was asked for with a token")
		}
		_, _ = w.Write([]byte(sampleIndexJSON))
	}))
	defer srv.Close()
	asGitHub(t, srv)
	asked := false
	idx, err := FetchIndex(context.Background(), srv.Client(), srv.URL+"/index.json", func() string { asked = true; return "t" })
	if err != nil || len(idx.Plugins) != 2 || asked {
		t.Fatalf("%v %v %v", idx, err, asked)
	}
}

func TestFetchIndexUsesATokenOnlyWhenGitHubRefusesAndOnlyOnGitHub(t *testing.T) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Header.Get("Authorization"))
		if r.Header.Get("Authorization") != "Bearer secret" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(sampleIndexJSON))
	}))
	defer srv.Close()
	token := func() string { return "secret" }

	// Another host never receives the token.
	if _, err := FetchIndex(context.Background(), srv.Client(), srv.URL, token); err == nil {
		t.Fatal("a non-GitHub host got the token")
	}
	if len(calls) != 1 || calls[0] != "" {
		t.Fatalf("calls: %q", calls)
	}

	asGitHub(t, srv)
	calls = nil
	if _, err := FetchIndex(context.Background(), srv.Client(), srv.URL, token); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || calls[0] != "" || calls[1] != "Bearer secret" {
		t.Fatalf("calls: %q", calls)
	}
}

func TestFetchIndexOfAPrivateRepositoryWithNoTokenSaysWhatToDo(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	asGitHub(t, srv)
	_, err := FetchIndex(context.Background(), srv.Client(), srv.URL, func() string { return "" })
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"private", "GITHUB_TOKEN", "gh auth login", LocationEnv} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error lacks %q: %v", want, err)
		}
	}
}

func TestFetchIndexRefusesAWrongTokenAndABadIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/bad") {
			_, _ = w.Write([]byte(`{"schema":9}`))
			return
		}
		http.Error(w, "no", http.StatusForbidden)
	}))
	defer srv.Close()
	asGitHub(t, srv)
	if _, err := FetchIndex(context.Background(), srv.Client(), srv.URL, func() string { return "x" }); err == nil || !strings.Contains(err.Error(), "even with your token") {
		t.Fatalf("got %v", err)
	}
	if _, err := FetchIndex(context.Background(), srv.Client(), srv.URL+"/bad", nil); err == nil || !strings.Contains(err.Error(), "schema 9") {
		t.Fatalf("got %v", err)
	}
}

func TestFetchIndexReadsAFileAndRefusesOtherSchemes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.json")
	if err := os.WriteFile(path, []byte(sampleIndexJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if idx, err := FetchIndex(context.Background(), http.DefaultClient, path, nil); err != nil || len(idx.Plugins) != 2 {
		t.Fatalf("%v %v", idx, err)
	}
	for _, loc := range []string{filepath.Join(t.TempDir(), "missing.json"), "ftp://example.com/index.json", "http://example.com/index.json"} {
		if _, err := FetchIndex(context.Background(), http.DefaultClient, loc, nil); err == nil {
			t.Errorf("%s: expected an error", loc)
		}
	}
}
