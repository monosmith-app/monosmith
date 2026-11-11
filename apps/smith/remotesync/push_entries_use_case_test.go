package remotesync

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

const testKey = "team-key"

// fakeStore imitates a plain HTTP object store: PUT, GET and HEAD under a
// path, a bearer token it insists on, and a log of the writes.
type fakeStore struct {
	mu      sync.Mutex
	objects map[string][]byte
	puts    []string
	token   string
	delay   time.Duration
}

func newFakeStore(t *testing.T) (*fakeStore, *httptest.Server) {
	f := &fakeStore{objects: map[string][]byte{}, token: "team-token"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.delay > 0 {
			time.Sleep(f.delay)
		}
		if r.Header.Get("Authorization") != "Bearer "+f.token {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		name := strings.TrimPrefix(r.URL.Path, "/team/")
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			f.objects[name] = body
			f.puts = append(f.puts, name)
		case http.MethodGet, http.MethodHead:
			body, ok := f.objects[name]
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
	return f, srv
}

// plan points the remote settings at srv, with the signing key and the token set.
func plan(t *testing.T, srv *httptest.Server) session.RemotePlan {
	t.Setenv("SMITH_CACHE_KEY", testKey)
	t.Setenv("SMITH_CACHE_TOKEN", "team-token")
	return session.RemotePlan{URL: srv.URL + "/team", Active: true}
}

// localWith makes a local cache holding n task results, each with one output file.
func localWith(t *testing.T, n int) *taskcache.LocalStorage {
	t.Helper()
	root := t.TempDir()
	store, err := taskcache.NewLocalStorage(filepath.Join(root, "cache"))
	if err != nil {
		t.Fatal(err)
	}
	for i := range n {
		project := "p" + string(rune('a'+i))
		dist := filepath.Join(root, project, "dist")
		if err := os.MkdirAll(dist, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dist, "out.js"), []byte("bundle "+project), 0o644); err != nil {
			t.Fatal(err)
		}
		key := strings.Repeat(string(rune('a'+i)), 64)
		if err := taskcache.Save(store, root, key, project+":build", project, []string{"dist"}, []byte("log "+project), time.Second, nil); err != nil {
			t.Fatal(err)
		}
	}
	return store
}

func connect(t *testing.T, p session.RemotePlan) Remote {
	t.Helper()
	r, err := Connect(p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestPushUploadsWhatTheRemoteLacksSignedAsARunSignsIt(t *testing.T) {
	fake, srv := newFakeStore(t)
	local := localWith(t, 2)
	remote := connect(t, plan(t, srv))
	var steps []PushStep
	result, err := Push(context.Background(), local, remote, PushOptions{Progress: func(s PushStep) { steps = append(steps, s) }})
	if err != nil {
		t.Fatal(err)
	}
	if result.Entries != 2 || result.Uploaded != 2 || result.Present != 0 || result.Failed != 0 || len(steps) != 2 {
		t.Fatalf("result = %+v, steps = %d", result, len(steps))
	}
	// Two entries, each with a log and an output: six objects.
	if len(fake.puts) != 6 {
		t.Fatalf("puts = %v", fake.puts)
	}
	// Blobs go before their entry.
	last := map[string]int{}
	for i, name := range fake.puts {
		last[name] = i
	}
	for name, i := range last {
		if strings.HasPrefix(name, "ac/") {
			for other, j := range last {
				if strings.HasPrefix(other, "cas/") && j > i && strings.Contains(string(fake.objects[name]), other[4:]) {
					t.Errorf("blob %s went after its entry %s", other, name)
				}
			}
		}
	}
	// Every object reads back through a Tiered cache: the signature is a run's.
	tiered := &remotecache.Tiered{Local: localWith(t, 0), Remote: remote.Store, Secret: []byte(testKey)}
	key := strings.Repeat("a", 64)
	if entry, err := taskcache.Lookup(tiered, key); err != nil || entry.Task != "pa:build" {
		t.Fatalf("lookup through the remote = %+v, %v", entry, err)
	}
}

func TestPushNeverUploadsWhatTheRemoteHolds(t *testing.T) {
	fake, srv := newFakeStore(t)
	local := localWith(t, 2)
	remote := connect(t, plan(t, srv))
	if _, err := Push(context.Background(), local, remote, PushOptions{}); err != nil {
		t.Fatal(err)
	}
	fake.puts = nil
	result, err := Push(context.Background(), local, remote, PushOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Present != 2 || result.Uploaded != 0 || len(fake.puts) != 0 {
		t.Fatalf("second push = %+v, puts = %v", result, fake.puts)
	}
}

func TestPushDryRunWritesNothing(t *testing.T) {
	fake, srv := newFakeStore(t)
	local := localWith(t, 3)
	remote := connect(t, plan(t, srv))
	result, err := Push(context.Background(), local, remote, PushOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Uploaded != 3 || result.Blobs != 6 || len(fake.puts) != 0 {
		t.Fatalf("dry run = %+v, puts = %v", result, fake.puts)
	}
	for _, s := range result.Steps {
		if s.Outcome != StepWouldUpload {
			t.Errorf("step %+v", s)
		}
	}
}

func TestPushStopsWhenTheRemoteRefusesTheCredentials(t *testing.T) {
	fake, srv := newFakeStore(t)
	fake.token = "another-token"
	local := localWith(t, 3)
	_, err := Push(context.Background(), local, connect(t, plan(t, srv)), PushOptions{})
	if !remotecache.Refused(err) {
		t.Fatalf("err = %v, want a refusal", err)
	}
}

func TestPushRefusesAReadOnlyRemote(t *testing.T) {
	_, srv := newFakeStore(t)
	p := plan(t, srv)
	p.ReadOnly = true
	if _, err := Push(context.Background(), localWith(t, 1), connect(t, p), PushOptions{}); err != ErrReadOnly {
		t.Fatalf("err = %v", err)
	}
}

func TestCountUnpushedCountsAndCaps(t *testing.T) {
	_, srv := newFakeStore(t)
	local := localWith(t, 5)
	remote := connect(t, plan(t, srv))
	got, err := CountUnpushed(context.Background(), local, remote.Store, CountLimits{})
	if err != nil || got != (Pending{Count: 5, Checked: 5}) {
		t.Fatalf("count = %+v, %v", got, err)
	}
	got, err = CountUnpushed(context.Background(), local, remote.Store, CountLimits{MaxEntries: 2})
	if err != nil || got != (Pending{Count: 2, Checked: 2, AtLeast: true}) {
		t.Fatalf("capped count = %+v, %v", got, err)
	}
	if _, err := Push(context.Background(), local, remote, PushOptions{}); err != nil {
		t.Fatal(err)
	}
	got, err = CountUnpushed(context.Background(), local, remote.Store, CountLimits{})
	if err != nil || got != (Pending{Count: 0, Checked: 5}) {
		t.Fatalf("count after push = %+v, %v", got, err)
	}
}

func TestCountUnpushedSaysAtLeastWhenTimeRunsOut(t *testing.T) {
	fake, srv := newFakeStore(t)
	fake.delay = 300 * time.Millisecond
	local := localWith(t, 4)
	remote := connect(t, plan(t, srv))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	got, err := CountUnpushed(ctx, local, remote.Store, CountLimits{Parallel: 1})
	if err != nil || !got.AtLeast {
		t.Fatalf("count = %+v, %v", got, err)
	}
	if took := time.Since(start); took > 250*time.Millisecond {
		t.Errorf("the count waited %s for a slow remote", took)
	}
}
