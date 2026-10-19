package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestBadServeNamesWhatCannotWork(t *testing.T) {
	web := &projectgraph.Project{Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{
		"start":    {Command: []string{"npm", "start"}},
		"e2e":      {Command: []string{"npx", "playwright", "test"}, Serve: "start", ServeURL: "http://127.0.0.1:5173", ServeTimeout: "60s"},
		"e2e-gone": {Serve: "web:missing"},
		"e2e-url":  {ServeURL: "http://127.0.0.1:1"},
		"e2e-bad":  {Serve: "start", ServeURL: "localhost:5173", ServeTimeout: "soon"},
		"e2e-time": {ServeTimeout: "5s"},
		"e2e-proj": {Serve: "api:start"},
	}}}
	pg, err := projectgraph.New([]*projectgraph.Project{web})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(BadServe(pg), "\n")
	for _, want := range []string{
		`web:e2e-bad: serve_url "localhost:5173" is not an http(s) URL`,
		`web:e2e-bad: serve_timeout "soon" is not a duration like 60s`,
		`web:e2e-gone: serve "web:missing": web has no "missing" task`,
		`web:e2e-proj: serve "api:start": there is no project "api"`,
		`web:e2e-time sets serve_timeout without serve`,
		`web:e2e-url sets serve_url without serve`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "web:e2e:") || strings.Contains(got, "web:e2e ") {
		t.Errorf("a valid serve is not reported:\n%s", got)
	}
}
