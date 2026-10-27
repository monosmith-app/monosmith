package dependencies

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const rootPubspec = `name: workspace
environment:
  sdk: ^3.5.0
workspace:
  - apps/web
dev_dependencies:
  lints: ^4.0.0
`

const webPubspec = `name: web
dependencies:
  flutter:
    sdk: flutter
  http: ^1.1.0 # network
  intl: "^0.19.0"
  collection:
    hosted: https://pub.dev
    version: ^1.18.0
  local_pkg:
    path: ../local
  fork:
    git: https://example.com/fork.git
  loose: any
  ranged: '>=1.0.0 <2.0.0'
dev_dependencies:
  test: ^1.24.0
`

const toolPubspec = `name: tool
dependencies:
  http: ^1.2.0
`

var projects = []protocol.Project{{Name: "web", Root: "apps/web"}, {Name: "tool", Root: "apps/tool"}}

func writeWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	for file, body := range map[string]string{"pubspec.yaml": rootPubspec, "apps/web/pubspec.yaml": webPubspec, "apps/tool/pubspec.yaml": toolPubspec} {
		full := filepath.Join(ws, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func TestDeclaredListsHostedPackagesOfEveryPubspec(t *testing.T) {
	res, err := Declared(writeWorkspace(t), protocol.DeclaredParams{Projects: projects})
	if err != nil {
		t.Fatal(err)
	}
	type row struct {
		Project, Name, Spec, Section string
		Rewritable                   bool
	}
	var got []row
	for _, d := range res.Declarations {
		got = append(got, row{d.Project, d.Name, d.Spec, d.Section, d.Rewritable})
	}
	want := []row{
		{"tool", "http", "^1.2.0", "dependencies", true},
		{"web", "collection", "^1.18.0", "dependencies", true},
		{"web", "http", "^1.1.0", "dependencies", true},
		{"web", "intl", "^0.19.0", "dependencies", true},
		{"web", "loose", "any", "dependencies", false},
		{"web", "ranged", ">=1.0.0 <2.0.0", "dependencies", false},
		{"web", "test", "^1.24.0", "dev_dependencies", true},
		{"", "lints", "^4.0.0", "dev_dependencies", true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("declarations:\n got %+v\nwant %+v", got, want)
	}
}

func TestRewriteChangesOnlyTheConstraint(t *testing.T) {
	ws := writeWorkspace(t)
	declared, _ := Declared(ws, protocol.DeclaredParams{Projects: projects})
	var changes []protocol.Rewrite
	for _, d := range declared.Declarations {
		switch d.Name {
		case "http", "intl", "collection", "test":
			changes = append(changes, protocol.Rewrite{Declaration: d, Spec: "^9.9.9"})
		}
	}
	res, err := Rewrite(ws, protocol.RewriteParams{Changes: changes})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 2 {
		t.Errorf("files: %v", res.Files)
	}
	raw, _ := os.ReadFile(filepath.Join(ws, "apps", "web", "pubspec.yaml"))
	for _, want := range []string{"  http: ^9.9.9 # network", `  intl: "^9.9.9"`, "    version: ^9.9.9", "  test: ^9.9.9", "  loose: any", "    sdk: flutter"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("web pubspec lacks %q:\n%s", want, raw)
		}
	}
	tool, _ := os.ReadFile(filepath.Join(ws, "apps", "tool", "pubspec.yaml"))
	if !strings.Contains(string(tool), "  http: ^9.9.9") {
		t.Errorf("tool pubspec:\n%s", tool)
	}
}

func TestRewriteRefusesAPackageThatIsGone(t *testing.T) {
	_, err := Rewrite(writeWorkspace(t), protocol.RewriteParams{Changes: []protocol.Rewrite{{
		Declaration: protocol.Declaration{File: "apps/tool/pubspec.yaml", Name: "gone", Section: SectionDependencies}, Spec: "^1.0.0",
	}}})
	if err == nil || !strings.Contains(err.Error(), "no longer declares gone") {
		t.Fatalf("%v", err)
	}
}

func TestLatestAsksPubDev(t *testing.T) {
	real := FetchJSON
	t.Cleanup(func() { FetchJSON = real })
	FetchJSON = func(url string, v any) error {
		switch {
		case strings.HasSuffix(url, "/private"):
			return errors.New(url + ": 404 Not Found")
		case strings.HasSuffix(url, "/offline"):
			return errors.New("no route")
		}
		return json.Unmarshal([]byte(`{"latest":{"version":"1.5.0"}}`), v)
	}
	res, err := Latest("", protocol.LatestParams{Names: []string{"http", "private", "offline"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Latest["http"] != "1.5.0" || len(res.Latest) != 1 || res.Unreachable["offline"] != "no route" || len(res.Unreachable) != 1 {
		t.Fatalf("%+v", res)
	}
}
