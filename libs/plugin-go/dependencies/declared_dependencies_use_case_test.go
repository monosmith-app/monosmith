package dependencies

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const apiMod = `module example.com/api

go 1.26

require github.com/google/uuid v1.5.0

require (
	github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/spf13/cobra v1.8.0 // direct
	golang.org/x/text v0.14.0 // indirect
	example.com/fork v0.0.0-20240101120000-abcdef123456
)

replace example.com/fork => ../fork

tool (
	golang.org/x/tools/cmd/stringer
)
`

func TestParseRequirementsReadsLinesAndBlocks(t *testing.T) {
	got := ParseRequirements(apiMod)
	want := []Requirement{
		{"github.com/google/uuid", "v1.5.0", false},
		{"github.com/Azure/go-autorest", "v14.2.0+incompatible", false},
		{"github.com/spf13/cobra", "v1.8.0", false},
		{"golang.org/x/text", "v0.14.0", true},
		{"example.com/fork", "v0.0.0-20240101120000-abcdef123456", false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestDeclaredListsDirectRequirementsWithoutTheV(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "api", "go.mod"), []byte(apiMod), 0o644)
	res, err := Declared(ws, protocol.DeclaredParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "ghost", Root: "apps/ghost"}}})
	if err != nil {
		t.Fatal(err)
	}
	type row struct {
		Name, Spec string
		Rewritable bool
	}
	var got []row
	for _, d := range res.Declarations {
		if d.Project != "api" || d.File != "apps/api/go.mod" || d.Section != SectionRequire {
			t.Errorf("%+v", d)
		}
		got = append(got, row{d.Name, d.Spec, d.Rewritable})
	}
	want := []row{
		{"example.com/fork", "0.0.0-20240101120000-abcdef123456", false},
		{"github.com/Azure/go-autorest", "14.2.0+incompatible", true},
		{"github.com/google/uuid", "1.5.0", true},
		{"github.com/spf13/cobra", "1.8.0", true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestRewriteRunsGoGetPerProject(t *testing.T) {
	var ran []string
	real := GoRunner
	t.Cleanup(func() { GoRunner = real })
	GoRunner = func(dir string, argv []string) error {
		ran = append(ran, filepath.Base(dir)+": "+strings.Join(argv, " "))
		return nil
	}
	decl := func(project, module string) protocol.Declaration {
		return protocol.Declaration{Project: project, File: "apps/" + project + "/go.mod", Name: module, Section: SectionRequire}
	}
	res, err := Rewrite("ws", protocol.RewriteParams{Changes: []protocol.Rewrite{
		{Declaration: decl("api", "github.com/google/uuid"), Spec: "1.6.0"},
		{Declaration: decl("web", "github.com/google/uuid"), Spec: "1.6.0"},
		{Declaration: decl("api", "github.com/spf13/cobra"), Spec: "1.9.1"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"api: go get github.com/google/uuid@v1.6.0 github.com/spf13/cobra@v1.9.1", "web: go get github.com/google/uuid@v1.6.0"}
	if !reflect.DeepEqual(ran, want) {
		t.Fatalf("ran %v", ran)
	}
	if len(res.Files) != 4 || res.Files[1] != "apps/api/go.sum" {
		t.Errorf("files: %v", res.Files)
	}
}

func TestLatestAsksTheModuleProxy(t *testing.T) {
	real := FetchJSON
	t.Cleanup(func() { FetchJSON = real })
	var urls []string
	FetchJSON = func(url string, v any) error {
		urls = append(urls, url)
		if strings.Contains(url, "offline") {
			return os.ErrDeadlineExceeded
		}
		return json.Unmarshal([]byte(`{"Version":"v1.6.0"}`), v)
	}
	res, err := Latest("", protocol.LatestParams{Names: []string{"github.com/Azure/go-autorest", "example.com/offline"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Latest["github.com/Azure/go-autorest"] != "1.6.0" || res.Unreachable["example.com/offline"] == "" || len(res.Tools) != 0 {
		t.Fatalf("%+v", res)
	}
	found := false
	for _, u := range urls {
		found = found || u == "https://proxy.golang.org/github.com/!azure/go-autorest/@latest"
	}
	if !found {
		t.Errorf("upper-case letters are escaped for the proxy: %v", urls)
	}
}
