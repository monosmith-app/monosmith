package taskrun

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/runevents"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestMain lets the test binary act as the task command: with SMITH_HELPER set
// it writes out/<file> containing its argument, or fails when asked to.
func TestMain(m *testing.M) {
	if mode := os.Getenv("SMITH_HELPER"); mode != "" {
		arg := os.Args[len(os.Args)-1]
		fmt.Println("helper ran with", arg)
		if arg == "fail" {
			os.Exit(3)
		}
		if arg == "needs-drop" {
			// The project is two levels below the workspace root.
			if _, err := os.Stat(filepath.Join("..", "..", "dist", "drop")); err != nil {
				os.Exit(6)
			}
		}
		if arg == "sleep" {
			time.Sleep(time.Minute)
		}
		if arg == "serve" {
			http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("up")) })
			fmt.Println("serving on", os.Getenv("SERVE_ADDR"))
			os.Exit(map[bool]int{true: 1, false: 0}[http.ListenAndServe(os.Getenv("SERVE_ADDR"), nil) != nil])
		}
		if arg == "probe" {
			resp, err := http.Get(os.Getenv("E2E_BASE_URL"))
			if err != nil || resp.StatusCode != 200 {
				fmt.Println("probe failed:", err)
				os.Exit(4)
			}
			_ = os.MkdirAll("out", 0o755)
			_ = os.WriteFile(filepath.Join("out", "result.txt"), []byte("probed "+os.Getenv("E2E_BASE_URL")), 0o644)
			os.Exit(0)
		}
		_ = os.MkdirAll("out", 0o755)
		_ = os.WriteFile(filepath.Join("out", "result.txt"), []byte(arg), 0o644)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

type fixture struct {
	root  string
	pg    *projectgraph.Graph
	files []string
	store taskcache.Storage
}

func newFixture(t *testing.T, libArg string) *fixture {
	t.Helper()
	root := t.TempDir()
	for _, f := range []string{"libs/core/src.txt", "apps/web/src.txt"} {
		p := filepath.Join(root, filepath.FromSlash(f))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte("v1"), 0o644)
	}
	self, _ := os.Executable()
	task := func(arg string, deps ...string) protocol.Task {
		return protocol.Task{Command: []string{self, "-test.run=^$", arg}, Env: map[string]string{"SMITH_HELPER": "1"}, Outputs: []string{"out"}, Cache: true, DependsOn: deps}
	}
	core := &projectgraph.Project{Project: protocol.Project{Name: "core", Root: "libs/core", Tasks: map[string]protocol.Task{"build": task(libArg)}}, Plugin: "test"}
	web := &projectgraph.Project{Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{"build": task("web", "^build")}}, Plugin: "test", Deps: []string{"core"}}
	pg, err := projectgraph.New([]*projectgraph.Project{core, web})
	if err != nil {
		t.Fatal(err)
	}
	store, _ := taskcache.NewLocalStorage(filepath.Join(root, ".cache"))
	return &fixture{root: root, pg: pg, files: []string{"apps/web/src.txt", "libs/core/src.txt"}, store: store}
}

func (f *fixture) run(t *testing.T) (Summary, string) {
	t.Helper()
	g, err := taskgraph.Build(f.pg, []string{"web"}, []string{"build"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	s := Run(context.Background(), g, Options{Root: f.root, Files: f.files, Projects: f.pg, Storage: f.store, Parallel: 2, Out: &out, ToolVersion: "test"})
	return s, out.String()
}

func TestRunCachesAndInvalidates(t *testing.T) {
	f := newFixture(t, "core")
	s, out := f.run(t)
	if !s.OK() || s.Count(Succeeded) != 2 {
		t.Fatalf("first run: %s\n%s", s, out)
	}
	if !strings.Contains(out, "web:build │ helper ran with web") {
		t.Fatalf("output must be prefixed:\n%s", out)
	}

	_ = os.RemoveAll(filepath.Join(f.root, "apps", "web", "out"))
	s, out = f.run(t)
	if s.Count(Cached) != 2 {
		t.Fatalf("second run must be fully cached: %s\n%s", s, out)
	}
	if body, _ := os.ReadFile(filepath.Join(f.root, "apps", "web", "out", "result.txt")); string(body) != "web" {
		t.Fatalf("outputs must be restored, got %q", body)
	}
	if !strings.Contains(out, "web:build │ helper ran with web") {
		t.Fatalf("a cache hit replays the log:\n%s", out)
	}

	// Changing the dependency's sources invalidates the dependent too.
	_ = os.WriteFile(filepath.Join(f.root, "libs", "core", "src.txt"), []byte("v2"), 0o644)
	s, _ = f.run(t)
	if s.Count(Succeeded) != 2 {
		t.Fatalf("a changed dependency must re-run both: %s", s)
	}
}

func TestFailureSkipsDependents(t *testing.T) {
	f := newFixture(t, "fail")
	s, out := f.run(t)
	if s.OK() || s.Count(Failed) != 1 || s.Count(Skipped) != 1 {
		t.Fatalf("summary %s\n%s", s, out)
	}
	if !strings.Contains(s.String(), "web:build because core:build did not succeed") {
		t.Fatalf("summary text %q", s.String())
	}
	if ok, _ := f.store.Has(taskcache.AC, s.Results[0].Key); ok {
		t.Fatal("a failure must not be cached")
	}
}

func TestCancelStopsTheProcess(t *testing.T) {
	f := newFixture(t, "sleep")
	g, _ := taskgraph.Build(f.pg, []string{"core"}, []string{"build"})
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	began := time.Now()
	s := Run(ctx, g, Options{Root: f.root, Files: f.files, Projects: f.pg, Out: &bytes.Buffer{}})
	if time.Since(began) > 20*time.Second || s.Results[0].Status != Failed {
		t.Fatalf("cancel must stop the task promptly: %+v after %s", s.Results, time.Since(began))
	}
}

func TestKeyOfMatchesTheRunAndExplainsAMiss(t *testing.T) {
	f := newFixture(t, "core")
	s, out := f.run(t)
	if !s.OK() {
		t.Fatalf("%s\n%s", s, out)
	}
	g, _ := taskgraph.Build(f.pg, []string{"web"}, []string{"build"})
	opts := Options{Root: f.root, Files: f.files, Projects: f.pg, ToolVersion: "test"}
	key, _ := KeyOf(g, opts, "web:build")
	if key != s.Results[1].Key || s.Results[1].Task != "web:build" {
		t.Fatalf("KeyOf %s, the run used %+v", key, s.Results)
	}

	_ = os.WriteFile(filepath.Join(f.root, "libs", "core", "src.txt"), []byte("v2"), 0o644)
	_, now := KeyOf(g, opts, "web:build")
	last, lastKey, err := taskcache.LastRun(f.store, "web:build")
	if err != nil || lastKey != key || last.Inputs == nil {
		t.Fatalf("last run %v %s %v", err, lastKey, last.Inputs)
	}
	var got []string
	for _, d := range taskcache.Diff(*last.Inputs, now) {
		got = append(got, d.String())
	}
	if strings.Join(got, "|") != "upstream   core changed|dependency core:build changed" {
		t.Fatalf("differences %q", got)
	}
}

func TestAChangedToolchainIsAMissAndWhyMissNamesIt(t *testing.T) {
	f := newFixture(t, "core")
	f.pg.Get("core").Toolchain = "node 20.11.0"
	f.pg.Get("web").Toolchain = "node 20.11.0"
	if s, out := f.run(t); s.Count(Succeeded) != 2 {
		t.Fatalf("first run: %s\n%s", s, out)
	}
	if s, out := f.run(t); s.Count(Cached) != 2 {
		t.Fatalf("the same toolchain is a hit: %s\n%s", s, out)
	}

	f.pg.Get("core").Toolchain = "node 22.3.0"
	f.pg.Get("web").Toolchain = "node 22.3.0"
	g, _ := taskgraph.Build(f.pg, []string{"web"}, []string{"build"})
	_, now := KeyOf(g, Options{Root: f.root, Files: f.files, Projects: f.pg, ToolVersion: "test"}, "core:build")
	last, _, err := taskcache.LastRun(f.store, "core:build")
	if err != nil || last.Inputs == nil {
		t.Fatalf("last run %v", err)
	}
	var got []string
	for _, d := range taskcache.Diff(*last.Inputs, now) {
		got = append(got, d.String())
	}
	if strings.Join(got, "|") != "toolchain  node 20.11.0 → 22.3.0" {
		t.Fatalf("differences %q", got)
	}
	if s, out := f.run(t); s.Count(Succeeded) != 2 {
		t.Fatalf("another toolchain re-runs both: %s\n%s", s, out)
	}
}

func TestAProjectThatReportsNoToolchainKeepsItsKeys(t *testing.T) {
	f := newFixture(t, "core")
	g, _ := taskgraph.Build(f.pg, []string{"web"}, []string{"build"})
	opts := Options{Root: f.root, Files: f.files, Projects: f.pg, ToolVersion: "test"}
	without, _ := KeyOf(g, opts, "web:build")
	f.pg.Get("web").Toolchain = "node 20.11.0"
	if with, _ := KeyOf(g, opts, "web:build"); with == without {
		t.Fatal("a reported toolchain is part of the key")
	}
	f.pg.Get("web").Toolchain = ""
	if again, _ := KeyOf(g, opts, "web:build"); again != without {
		t.Fatal("an empty toolchain leaves the key as it was")
	}
}

func TestAPackageTaskFindsDistDropAlreadyThere(t *testing.T) {
	f := newFixture(t, "core")
	self, _ := os.Executable()
	web := f.pg.Get("web")
	web.Tasks["package"] = protocol.Task{Command: []string{self, "-test.run=^$", "needs-drop"}, Env: map[string]string{"SMITH_HELPER": "1"}}
	g, err := taskgraph.Build(f.pg, []string{"web"}, []string{"package"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	s := Run(context.Background(), g, Options{Root: f.root, Files: f.files, Projects: f.pg, Storage: f.store, Out: &out, ToolVersion: "test"})
	if !s.OK() {
		t.Fatalf("a fresh workspace has no dist/drop, and npm pack does not create it: %s\n%s", s, out.String())
	}
}

type collected struct {
	mu     sync.Mutex
	events []runevents.Event
}

func (c *collected) Emit(e runevents.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, e)
}

func (c *collected) kinds() []string {
	var out []string
	for _, e := range c.events {
		out = append(out, fmt.Sprintf("%s %s %s", e.Event, e.Task, e.Status))
	}
	return out
}

func (f *fixture) runWithEvents(t *testing.T, tasks ...string) (Summary, *collected) {
	t.Helper()
	g, err := taskgraph.Build(f.pg, tasks, []string{"build"})
	if err != nil {
		t.Fatal(err)
	}
	got := &collected{}
	s := Run(context.Background(), g, Options{Root: f.root, Files: f.files, Projects: f.pg, Storage: f.store, Parallel: 2,
		Out: &bytes.Buffer{}, ToolVersion: "test", Command: "smith run build", Events: runevents.NewRunStream("r1", got)})
	return s, got
}

func TestARunEmitsItsEventsInOrder(t *testing.T) {
	f := newFixture(t, "core")
	s, got := f.runWithEvents(t, "web")
	if !s.OK() {
		t.Fatalf("%s", s)
	}
	want := []string{
		"run.started  ", "task.queued core:build ", "task.queued web:build ",
		"task.started core:build ", "task.output core:build ", "task.finished core:build passed",
		"task.started web:build ", "task.output web:build ", "task.finished web:build passed",
		"run.finished  ",
	}
	if fmt.Sprint(got.kinds()) != fmt.Sprint(want) {
		t.Fatalf("events\n got %q\nwant %q", got.kinds(), want)
	}
	for i, e := range got.events {
		if e.Run != "r1" || e.Seq != i+1 || e.Time.IsZero() {
			t.Fatalf("event %d is not stamped: %+v", i, e)
		}
	}
	if got.events[0].Command != "smith run build" || len(got.events[0].Tasks) != 2 {
		t.Fatalf("run.started %+v", got.events[0])
	}
	if out := got.events[4]; out.Line != "helper ran with core" || out.Stream != runevents.Combined {
		t.Fatalf("task.output %+v", out)
	}
	last := got.events[len(got.events)-1]
	if last.OK == nil || !*last.OK || last.Summary.Passed != 2 {
		t.Fatalf("run.finished %+v", last)
	}
	if got.events[5].CacheKey == "" || got.events[5].CacheKey != s.Results[0].Key {
		t.Fatalf("task.finished carries the cache key: %+v", got.events[5])
	}
}

func TestACachedTaskAndASkippedTaskAreReportedAsSuch(t *testing.T) {
	f := newFixture(t, "core")
	f.runWithEvents(t, "web")
	_ = os.RemoveAll(filepath.Join(f.root, "apps", "web", "out"))
	_, got := f.runWithEvents(t, "web")
	statuses := map[string]runevents.Status{}
	for _, e := range got.events {
		if e.Event == runevents.TaskFinished {
			statuses[e.Task] = e.Status
		}
	}
	if statuses["core:build"] != runevents.Cached || statuses["web:build"] != runevents.Cached {
		t.Fatalf("second run %v", statuses)
	}

	failing := newFixture(t, "fail")
	s, got := failing.runWithEvents(t, "web")
	if s.OK() {
		t.Fatal("the run must fail")
	}
	statuses = map[string]runevents.Status{}
	for _, e := range got.events {
		if e.Event == runevents.TaskFinished {
			statuses[e.Task] = e.Status
		}
	}
	if statuses["core:build"] != runevents.Failed || statuses["web:build"] != runevents.Skipped {
		t.Fatalf("failing run %v", statuses)
	}
	if last := got.events[len(got.events)-1]; last.OK == nil || *last.OK || last.Summary.Failed != 1 || last.Summary.Skipped != 1 {
		t.Fatalf("run.finished %+v", last)
	}
}

func TestASharedOutputFolderKeepsTheFilesAnotherTaskWrote(t *testing.T) {
	for name, tc := range map[string]struct {
		shared   []string
		wantKept bool
	}{
		"listed in shared_outputs": {[]string{"apps/web/out"}, true},
		"not listed":               {nil, false},
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t, "core")
			keep := filepath.Join(f.root, "apps", "web", "out", "earlier.zip")
			_ = os.MkdirAll(filepath.Dir(keep), 0o755)
			if err := os.WriteFile(keep, []byte("zip"), 0o644); err != nil {
				t.Fatal(err)
			}
			g, err := taskgraph.Build(f.pg, []string{"web"}, []string{"build"})
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			s := Run(context.Background(), g, Options{Root: f.root, Files: f.files, Projects: f.pg, Storage: f.store, Out: &out, ToolVersion: "test", SharedOutputs: tc.shared})
			if !s.OK() {
				t.Fatalf("run failed:\n%s", out.String())
			}
			_, err = os.Stat(keep)
			if kept := err == nil; kept != tc.wantKept {
				t.Errorf("earlier file kept = %v, want %v", kept, tc.wantKept)
			}
		})
	}
}

func TestABuildThatShipsLinksItsFileIntoTheShippedFolder(t *testing.T) {
	f := newFixture(t, "core")
	web := f.pg.Get("web")
	build := web.Tasks["build"]
	build.Ship = []string{"out/result.txt"}
	web.Tasks["build"] = build
	_, out := f.run(t)
	shipped := filepath.Join(f.root, "dist", "apps", "web", "result.txt")
	if b, err := os.ReadFile(shipped); err != nil || string(b) != "web" {
		t.Fatalf("shipped file %q, %v:\n%s", b, err, out)
	}
	if !strings.Contains(out, "web:build │ shipped 1 file to dist/apps/web") {
		t.Errorf("the run does not say what it shipped:\n%s", out)
	}
}
