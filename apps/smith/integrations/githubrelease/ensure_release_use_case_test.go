package githubrelease

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type fakeGitHub struct {
	mu       sync.Mutex
	releases map[string]*Release
	uploads  []string
	content  map[int64]string
	deleted  []int64
	push     bool
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.Header.Get("Authorization") != "Bearer tok" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repos/o/r/releases/tags/"):
		rel, ok := f.releases[strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/tags/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(rel)
	case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/releases":
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		rel := &Release{ID: 1, HTMLURL: "https://example/" + body["tag_name"], UploadURL: "http://" + r.Host + "/upload/" + body["tag_name"] + "{?name,label}"}
		f.releases[body["tag_name"]] = rel
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(rel)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/upload/"):
		raw, _ := io.ReadAll(r.Body)
		name := r.URL.Query().Get("name")
		rel := f.releases[strings.TrimPrefix(r.URL.Path, "/upload/")]
		id := int64(len(f.content) + 100)
		if f.content == nil {
			f.content = map[int64]string{}
			id = 100
		}
		f.content[id] = string(raw)
		rel.Assets = append(rel.Assets, Asset{ID: id, Name: name, Size: int64(len(raw))})
		f.uploads = append(f.uploads, name+"="+string(raw))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repos/o/r/releases/assets/"):
		var id int64
		_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/assets/"), "%d", &id)
		_, _ = w.Write([]byte(f.content[id]))
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/repos/o/r/releases/assets/"):
		var id int64
		_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/assets/"), "%d", &id)
		f.deleted = append(f.deleted, id)
		for _, rel := range f.releases {
			kept := rel.Assets[:0]
			for _, a := range rel.Assets {
				if a.ID != id {
					kept = append(kept, a)
				}
			}
			rel.Assets = kept
		}
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r":
		_, _ = fmt.Fprintf(w, `{"permissions":{"push":%t}}`, f.push)
	default:
		w.WriteHeader(http.StatusTeapot)
	}
}

func TestEnsureCreatesOnceAndUploadsOnlyMissingAssets(t *testing.T) {
	fake := &fakeGitHub{releases: map[string]*Release{}}
	srv := httptest.NewServer(fake)
	defer srv.Close()
	c := &Client{Token: "tok", API: srv.URL, HTTP: srv.Client()}
	dir := t.TempDir()
	a, b := filepath.Join(dir, "app_1.0.0_linux_amd64.zip"), filepath.Join(dir, "checksums.txt")
	_ = os.WriteFile(a, []byte("A"), 0o644)
	_ = os.WriteFile(b, []byte("B"), 0o644)
	repo := Repository{Owner: "o", Name: "r"}

	out, err := Ensure(c, repo, Request{Tag: "app@1.0.0", Title: "app 1.0.0", Notes: "n", Assets: []string{a}})
	if err != nil || !out.Created || len(out.Uploaded) != 1 {
		t.Fatalf("first run: %+v %v", out, err)
	}
	out, err = Ensure(c, repo, Request{Tag: "app@1.0.0", Assets: []string{a, b}})
	if err != nil || out.Created || len(out.Uploaded) != 1 || out.Uploaded[0] != "checksums.txt" {
		t.Fatalf("a re-run uploads only what is missing: %+v %v", out, err)
	}
	if len(fake.uploads) != 2 || fake.uploads[0] != "app_1.0.0_linux_amd64.zip=A" {
		t.Fatalf("uploads %v", fake.uploads)
	}
}

func TestEnsureReportsAPIErrors(t *testing.T) {
	srv := httptest.NewServer(&fakeGitHub{releases: map[string]*Release{}})
	defer srv.Close()
	c := &Client{Token: "wrong", API: srv.URL, HTTP: srv.Client()}
	if _, err := Ensure(c, Repository{Owner: "o", Name: "r"}, Request{Tag: "x@1.0.0"}); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("want a 401 error, got %v", err)
	}
}

func TestEnsureVerifyKeepsAnIdenticalAssetAndReplacesADifferentOne(t *testing.T) {
	fake := &fakeGitHub{releases: map[string]*Release{}}
	srv := httptest.NewServer(fake)
	defer srv.Close()
	c := &Client{Token: "tok", API: srv.URL, HTTP: srv.Client()}
	repo := Repository{Owner: "o", Name: "r"}
	dir := t.TempDir()
	same, other := filepath.Join(dir, "same.tgz"), filepath.Join(dir, "other.tgz")
	_ = os.WriteFile(same, []byte("1111"), 0o644)
	_ = os.WriteFile(other, []byte("2222"), 0o644)
	if _, err := Ensure(c, repo, Request{Tag: "t", Assets: []string{same, other}, Verify: true}); err != nil {
		t.Fatal(err)
	}
	// The same size, other bytes: only the content tells them apart.
	_ = os.WriteFile(other, []byte("3333"), 0o644)
	out, err := Ensure(c, repo, Request{Tag: "t", Assets: []string{same, other}, Verify: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Replaced) != 1 || out.Replaced[0] != "other.tgz" || len(out.Uploaded) != 1 || len(fake.deleted) != 1 {
		t.Fatalf("only the differing asset is replaced: %+v deleted %v", out, fake.deleted)
	}
	out, err = Ensure(c, repo, Request{Tag: "t", Assets: []string{same, other}, Verify: true})
	if err != nil || len(out.Replaced) != 0 || len(out.Uploaded) != 0 {
		t.Fatalf("a third run changes nothing: %+v %v", out, err)
	}
}

func TestCanPushReadsThePermissionsOfTheRepository(t *testing.T) {
	fake := &fakeGitHub{releases: map[string]*Release{}}
	srv := httptest.NewServer(fake)
	defer srv.Close()
	c := &Client{Token: "tok", API: srv.URL, HTTP: srv.Client()}
	if ok, err := c.CanPush(Repository{Owner: "o", Name: "r"}); err != nil || ok {
		t.Fatalf("a read-only token cannot push: %v %v", ok, err)
	}
	fake.push = true
	if ok, err := c.CanPush(Repository{Owner: "o", Name: "r"}); err != nil || !ok {
		t.Fatalf("a token with push can: %v %v", ok, err)
	}
}

func TestTheTokenComesFromTheEnvironmentThenFromGh(t *testing.T) {
	real := GhToken
	defer func() { GhToken = real }()
	GhToken = func() string { return "from-gh" }
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	c, err := FromEnvironment()
	if err != nil || c.Token != "from-gh" {
		t.Fatalf("gh supplies the token when the environment has none: %v", err)
	}
	t.Setenv("GITHUB_TOKEN", "from-env")
	if c, _ := FromEnvironment(); c.Token != "from-env" {
		t.Fatal("the environment wins over gh")
	}
	GhToken = func() string { return "" }
	t.Setenv("GITHUB_TOKEN", "")
	if _, err := FromEnvironment(); err != ErrNoToken {
		t.Fatalf("no token anywhere: %v", err)
	}
}
