package remotecache

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// objectServer is a plain HTTP object store: PUT, GET and HEAD by path, and
// a bearer token it insists on.
func objectServer(t *testing.T) (*httptest.Server, map[string][]byte) {
	var mu sync.Mutex
	objects := map[string][]byte{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer team-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			objects[r.URL.Path] = body
		case http.MethodDelete:
			delete(objects, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet, http.MethodHead:
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
	t.Cleanup(srv.Close)
	return srv, objects
}

func tiered(t *testing.T, remote string, secret string, readOnly bool) (*Tiered, *[]string) {
	t.Helper()
	local, err := taskcache.NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store, err := Open(remote)
	if err != nil {
		t.Fatal(err)
	}
	var warnings []string
	return &Tiered{Local: local, Remote: store, Secret: []byte(secret), ReadOnly: readOnly, Warn: func(m string) { warnings = append(warnings, m) }}, &warnings
}

func digest(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func read(t *testing.T, s taskcache.Storage, kind taskcache.Kind, key string) (string, error) {
	t.Helper()
	r, err := s.Get(kind, key)
	if err != nil {
		return "", err
	}
	defer func() { _ = r.Close() }()
	b, _ := io.ReadAll(r)
	return string(b), nil
}

func TestTieredSharesSignedEntries(t *testing.T) {
	srv, objects := objectServer(t)
	t.Setenv("SMITH_CACHE_TOKEN", "team-token")
	remote := srv.URL + "/smith-cache"
	ci, _ := tiered(t, remote, "team-key", false)
	blob := "compiled output"
	if err := ci.Put(taskcache.CAS, digest(blob), strings.NewReader(blob)); err != nil {
		t.Fatal(err)
	}
	if err := ci.Put(taskcache.AC, "aaaa", strings.NewReader(`{"entry":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := ci.Put(taskcache.Last, "bbbb", strings.NewReader("aaaa")); err != nil {
		t.Fatal(err)
	}
	if len(objects) != 2 || objects["/smith-cache/last/bbbb"] != nil {
		t.Fatalf("the remote holds the cas and ac objects, never last: %v", keys(objects))
	}
	if stored := objects["/smith-cache/ac/aaaa"]; len(stored) != SignatureSize+len(`{"entry":1}`) {
		t.Fatalf("an object is its signature and its content: %d bytes", len(stored))
	}

	// A developer's empty cache reads what CI wrote, verified.
	dev, warnings := tiered(t, remote, "team-key", true)
	if ok, _ := dev.Has(taskcache.CAS, digest(blob)); !ok {
		t.Fatal("the blob comes from the remote")
	}
	if got, err := read(t, dev, taskcache.AC, "aaaa"); err != nil || got != `{"entry":1}` {
		t.Fatalf("ac %q %v", got, err)
	}
	if ok, _ := dev.Local.Has(taskcache.AC, "aaaa"); !ok {
		t.Fatal("a verified remote entry is kept locally")
	}
	_ = dev.Put(taskcache.AC, "cccc", strings.NewReader("x"))
	if objects["/smith-cache/ac/cccc"] != nil || len(*warnings) != 0 {
		t.Fatalf("a read-only cache does not write: %v", *warnings)
	}

	// Changed on the server, or signed with another key: a miss, said once.
	objects["/smith-cache/ac/aaaa"][SignatureSize] = '['
	other, warnings := tiered(t, remote, "team-key", false)
	if _, err := read(t, other, taskcache.AC, "aaaa"); err != taskcache.ErrMiss {
		t.Fatalf("a tampered entry must miss: %v", err)
	}
	stranger, _ := tiered(t, remote, "another-key", false)
	if ok, _ := stranger.Has(taskcache.CAS, digest(blob)); ok {
		t.Fatal("another key's objects are refused")
	}
	if len(*warnings) != 1 || !strings.Contains((*warnings)[0], "refused") {
		t.Fatalf("warnings %v", *warnings)
	}
}

func TestTieredSurvivesAnUnreachableRemote(t *testing.T) {
	t.Setenv("SMITH_CACHE_TOKEN", "")
	cache, warnings := tiered(t, "http://127.0.0.1:1/cache", "k", false)
	if err := cache.Put(taskcache.AC, "aaaa", strings.NewReader("x")); err != nil {
		t.Fatal("a remote that cannot be written must not fail the run")
	}
	if got, err := read(t, cache, taskcache.AC, "aaaa"); err != nil || got != "x" {
		t.Fatal("the local copy still serves")
	}
	if _, err := read(t, cache, taskcache.AC, "zzzz"); err != taskcache.ErrMiss {
		t.Fatalf("an unreachable remote is a miss: %v", err)
	}
	if len(*warnings) != 2 {
		t.Fatalf("one warning per kind of problem: %v", *warnings)
	}
}

func TestOpenRefusesWhatItCannotUse(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AZURE_STORAGE_SAS_TOKEN", "")
	for remote, want := range map[string]string{"s3://bucket/p": "AWS_ACCESS_KEY_ID", "azblob://acct/c": "AZURE_STORAGE_SAS_TOKEN", "ftp://x": "https://, s3:// or azblob://"} {
		if _, err := Open(remote); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: %v", remote, err)
		}
	}
}

func keys(m map[string][]byte) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
