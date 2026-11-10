package releasepack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
)

// fakeGitHub is a GitHub that holds releases and their assets in memory.
type fakeGitHub struct {
	mu       sync.Mutex
	releases map[string]*githubrelease.Release
	content  map[int64]string
}

func newFake(tags ...string) *fakeGitHub {
	f := &fakeGitHub{releases: map[string]*githubrelease.Release{}, content: map[int64]string{}}
	for i, tag := range tags {
		f.releases[tag] = &githubrelease.Release{ID: int64(i + 1), HTMLURL: "https://example/" + tag}
	}
	return f
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repos/o/r/releases/tags/"):
		rel, ok := f.releases[strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/tags/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		rel.UploadURL = "http://" + r.Host + "/upload/" + strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/tags/") + "{?name}"
		_ = json.NewEncoder(w).Encode(rel)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/upload/"):
		raw, _ := io.ReadAll(r.Body)
		rel := f.releases[strings.TrimPrefix(r.URL.Path, "/upload/")]
		id := int64(len(f.content) + 100)
		f.content[id] = string(raw)
		rel.Assets = append(rel.Assets, githubrelease.Asset{ID: id, Name: r.URL.Query().Get("name"), Size: int64(len(raw))})
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repos/o/r/releases/assets/"):
		var id int64
		_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/assets/"), "%d", &id)
		_, _ = w.Write([]byte(f.content[id]))
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/repos/o/r/releases/assets/"):
		var id int64
		_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/assets/"), "%d", &id)
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
	default:
		w.WriteHeader(http.StatusTeapot)
	}
}

func (f *fakeGitHub) assets(tag string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var names []string
	for _, a := range f.releases[tag].Assets {
		names = append(names, a.Name)
	}
	return names
}

func (f *fakeGitHub) body(tag, name string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.releases[tag].Assets {
		if a.Name == name {
			return f.content[a.ID]
		}
	}
	return ""
}

func client(srv *httptest.Server) *githubrelease.Client {
	return &githubrelease.Client{Token: "tok", API: srv.URL, HTTP: srv.Client()}
}

var repo = githubrelease.Repository{Owner: "o", Name: "r"}

func npmPackage(project, version, file string) Package {
	return Package{Project: project, Ecosystem: "npm", Version: version, Tag: project + "@" + version, Files: []File{{Name: file, SHA256: strings.Repeat("a", 64), Size: 3}}}
}
