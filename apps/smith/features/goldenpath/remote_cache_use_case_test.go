//go:build e2e

package goldenpath

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestRemoteCache shares task results through a remote store: one machine
// builds and fills it, another (an empty local cache) replays every task from
// it, and objects signed with another key are refused.
func TestRemoteCache(t *testing.T) {
	var mu sync.Mutex
	objects := map[string][]byte{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			objects[r.URL.Path] = body
		default:
			body, ok := objects[r.URL.Path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if r.Method == http.MethodGet {
				_, _ = w.Write(body)
			}
		}
	}))
	defer srv.Close()

	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: append(identity(), "SMITH_CACHE_KEY=team-secret")}
	w.smith(true, "new", "acme", "--ci", "github")
	w.dir = filepath.Join(base, "acme")
	w.smith(true, "add", "lib", "core", "--lang", "go")
	w.smith(true, "config", "set", "cache.remote", srv.URL+"/cache")

	if out := w.smith(true, "build"); !strings.Contains(out, "(0 from cache)") {
		t.Fatalf("first build:\n%s", out)
	}
	mu.Lock()
	stored := len(objects)
	mu.Unlock()
	if stored == 0 {
		t.Fatal("the build's results must reach the remote cache")
	}

	// Another machine: nothing cached locally, everything from the remote.
	_ = os.RemoveAll(filepath.Join(w.dir, ".smith", "cache"))
	if out := w.smith(true, "build"); !strings.Contains(out, "from cache") || strings.Contains(out, "(0 from cache)") {
		t.Fatalf("a build on an empty local cache must replay from the remote:\n%s", out)
	}

	// A machine with another key trusts nothing it finds there.
	_ = os.RemoveAll(filepath.Join(w.dir, ".smith", "cache"))
	w.env = append(identity(), "SMITH_CACHE_KEY=someone-else", "SMITH_CACHE_READ_ONLY=1")
	if out := w.smith(true, "build"); !strings.Contains(out, "(0 from cache)") || !strings.Contains(out, "was refused") {
		t.Fatalf("objects signed with another key must be refused:\n%s", out)
	}
}
