package taskrun

import (
	"bytes"
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// serveFixture is a project web with a start task (an HTTP server) and an e2e
// task that probes E2E_BASE_URL, serving start.
type serveFixture struct {
	root, addr, url string
	pg              *projectgraph.Graph
}

func newServeFixture(t *testing.T, serveCommandArg string) *serveFixture {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "apps", "web"), 0o755)
	self, _ := os.Executable()
	start := protocol.Task{Command: []string{self, "-test.run=^$", serveCommandArg}, Env: map[string]string{"SMITH_HELPER": "1", "SERVE_ADDR": addr}}
	e2e := protocol.Task{Command: []string{self, "-test.run=^$", "probe"}, Env: map[string]string{"SMITH_HELPER": "1"}, Outputs: []string{"out"},
		Serve: "start", ServeURL: "http://" + addr + "/", ServeTimeout: "20s"}
	web := &projectgraph.Project{Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{"start": start, "e2e": e2e}}, Plugin: "test"}
	pg, err := projectgraph.New([]*projectgraph.Project{web})
	if err != nil {
		t.Fatal(err)
	}
	return &serveFixture{root: root, addr: addr, url: "http://" + addr + "/", pg: pg}
}

func (f *serveFixture) run(t *testing.T) (Summary, string) {
	t.Helper()
	g, err := taskgraph.Build(f.pg, []string{"web"}, []string{"e2e"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	s := Run(context.Background(), g, Options{Root: f.root, Files: nil, Projects: f.pg, Parallel: 2, Out: &out, ToolVersion: "test"})
	return s, out.String()
}

func (f *serveFixture) listening() bool {
	c, err := net.Dial("tcp", f.addr)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

func TestAServedAppIsUpWhileTheTaskRunsAndStoppedAfter(t *testing.T) {
	f := newServeFixture(t, "serve")
	s, out := f.run(t)
	if !s.OK() {
		t.Fatalf("%s\n%s", s, out)
	}
	body, _ := os.ReadFile(filepath.Join(f.root, "apps", "web", "out", "result.txt"))
	if string(body) != "probed "+f.url {
		t.Errorf("the task gets E2E_BASE_URL: %q", body)
	}
	if !strings.Contains(out, "web:start │ serving on") || !strings.Contains(out, "web:start ■ stopped") {
		t.Errorf("the app's output is shown, and it is stopped:\n%s", out)
	}
	if f.listening() {
		t.Error("the app is still running after the task")
	}
}

func TestTheAppIsStoppedWhenTheTaskFails(t *testing.T) {
	f := newServeFixture(t, "serve")
	f.pg.Get("web").Tasks["e2e"] = protocol.Task{Command: []string{os.Args[0], "-test.run=^$", "fail"}, Env: map[string]string{"SMITH_HELPER": "1"},
		Serve: "start", ServeURL: f.url, ServeTimeout: "20s"}
	s, out := f.run(t)
	if s.OK() || s.Count(Failed) != 1 {
		t.Fatalf("%s\n%s", s, out)
	}
	if f.listening() {
		t.Error("the app is still running after a failed task")
	}
}

func TestATaskDoesNotRunWhenTheAppExitsBeforeItAnswers(t *testing.T) {
	f := newServeFixture(t, "fail")
	s, out := f.run(t)
	if s.OK() || !strings.Contains(out, "before "+f.url+" answered") {
		t.Fatalf("%s\n%s", s, out)
	}
	if _, err := os.Stat(filepath.Join(f.root, "apps", "web", "out", "result.txt")); err == nil {
		t.Error("the suite ran without its app")
	}
}

func TestATaskFailsWhenTheAppNeverAnswers(t *testing.T) {
	f := newServeFixture(t, "sleep")
	e2e := f.pg.Get("web").Tasks["e2e"]
	e2e.ServeTimeout = "1s"
	f.pg.Get("web").Tasks["e2e"] = e2e
	s, out := f.run(t)
	if s.OK() || !strings.Contains(out, "did not answer with a 2xx within 1s") {
		t.Fatalf("%s\n%s", s, out)
	}
}

func TestServeNamesATaskThatMustExist(t *testing.T) {
	f := newServeFixture(t, "serve")
	e2e := f.pg.Get("web").Tasks["e2e"]
	e2e.Serve = "web:missing"
	f.pg.Get("web").Tasks["e2e"] = e2e
	if _, err := taskgraph.Build(f.pg, []string{"web"}, []string{"e2e"}); err == nil || !strings.Contains(err.Error(), `web has no "missing" task`) {
		t.Errorf("%v", err)
	}
}

func TestTheServedAppAndItsUrlAreInTheKey(t *testing.T) {
	f := newServeFixture(t, "serve")
	g, _ := taskgraph.Build(f.pg, []string{"web"}, []string{"e2e"})
	opts := Options{Root: f.root, Projects: f.pg, ToolVersion: "test"}
	before, _ := KeyOf(g, opts, "web:e2e")
	e2e := f.pg.Get("web").Tasks["e2e"]
	e2e.ServeURL = "http://127.0.0.1:1/"
	f.pg.Get("web").Tasks["e2e"] = e2e
	g, _ = taskgraph.Build(f.pg, []string{"web"}, []string{"e2e"})
	if after, _ := KeyOf(g, opts, "web:e2e"); after == before {
		t.Error("another url is another key")
	}
}
