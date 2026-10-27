package dependencies

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-php/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const rootComposer = `{
    "name": "acme/workspace",
    "require-dev": {
        "phpunit/phpunit": "^10.5"
    }
}
`

const apiComposer = `{
    "name": "acme/api",
    "require": {
        "php": ">=8.2",
        "ext-json": "*",
        "monolog/monolog": "^3.4.0",
        "guzzlehttp/guzzle": "^7.8 || ^8.0",
        "acme/core": "dev-main"
    },
    "require-dev": {
        "phpunit/phpunit": "^10.4",
        "phpstan/phpstan": "~1.10"
    }
}
`

const toolComposer = `{
    "name": "acme/tool",
    "require": {
        "monolog/monolog": "^3.5.0"
    }
}
`

var projects = []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "tool", Root: "apps/tool"}}

func writeWorkspace(t *testing.T, lock bool) string {
	t.Helper()
	ws := t.TempDir()
	files := map[string]string{"composer.json": rootComposer, "apps/api/composer.json": apiComposer, "apps/tool/composer.json": toolComposer}
	if lock {
		files["composer.lock"] = "{}"
	}
	for file, body := range files {
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

func TestDeclaredListsPackagistPackagesOfEveryComposerJSON(t *testing.T) {
	res, err := Declared(writeWorkspace(t, false), protocol.DeclaredParams{Projects: projects})
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
		{"api", "acme/core", "dev-main", "require", false},
		{"api", "guzzlehttp/guzzle", "^7.8 || ^8.0", "require", false},
		{"api", "monolog/monolog", "^3.4.0", "require", true},
		{"api", "phpstan/phpstan", "~1.10", "require-dev", true},
		{"api", "phpunit/phpunit", "^10.4", "require-dev", true},
		{"tool", "monolog/monolog", "^3.5.0", "require", true},
		{"", "phpunit/phpunit", "^10.5", "require-dev", true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("declarations:\n got %+v\nwant %+v", got, want)
	}
}

func TestRewriteChangesOnlyTheConstraintAndUpdatesTheLock(t *testing.T) {
	ws := writeWorkspace(t, true)
	var ran []string
	real := scaffolding.Runner
	t.Cleanup(func() { scaffolding.Runner = real })
	scaffolding.Runner = func(_ string, argv []string) error {
		ran = append(ran, strings.Join(argv, " "))
		return nil
	}
	declared, _ := Declared(ws, protocol.DeclaredParams{Projects: projects})
	var changes []protocol.Rewrite
	for _, d := range declared.Declarations {
		if d.Rewritable && (d.Name == "monolog/monolog" || d.Name == "phpunit/phpunit") {
			changes = append(changes, protocol.Rewrite{Declaration: d, Spec: "^9.9.9"})
		}
	}
	res, err := Rewrite(ws, protocol.RewriteParams{Changes: changes})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 4 || res.Files[len(res.Files)-1] != "composer.lock" {
		t.Errorf("files: %v", res.Files)
	}
	if want := []string{"composer update --no-interaction --no-progress --no-install --with-dependencies monolog/monolog phpunit/phpunit"}; !reflect.DeepEqual(ran, want) {
		t.Errorf("composer ran %v", ran)
	}
	raw, _ := os.ReadFile(filepath.Join(ws, "apps", "api", "composer.json"))
	for _, want := range []string{`"monolog/monolog": "^9.9.9"`, `"phpunit/phpunit": "^9.9.9"`, `"phpstan/phpstan": "~1.10"`, `"guzzlehttp/guzzle": "^7.8 || ^8.0"`, "\n        \"php\": \">=8.2\""} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("api composer.json lacks %q:\n%s", want, raw)
		}
	}
}

func TestRewriteLeavesTheLockAloneWhenThereIsNone(t *testing.T) {
	ws := writeWorkspace(t, false)
	real := scaffolding.Runner
	t.Cleanup(func() { scaffolding.Runner = real })
	scaffolding.Runner = func(string, []string) error { t.Error("composer must not run without a lockfile"); return nil }
	_, err := Rewrite(ws, protocol.RewriteParams{Changes: []protocol.Rewrite{{
		Declaration: protocol.Declaration{File: "apps/tool/composer.json", Name: "monolog/monolog", Section: SectionRequire}, Spec: "^4.0.0",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(filepath.Join(ws, "apps", "tool", "composer.json")); !strings.Contains(string(raw), `"monolog/monolog": "^4.0.0"`) {
		t.Errorf("%s", raw)
	}
}

func TestRewriteRefusesAPackageThatIsGone(t *testing.T) {
	_, err := Rewrite(writeWorkspace(t, false), protocol.RewriteParams{Changes: []protocol.Rewrite{{
		Declaration: protocol.Declaration{File: "apps/tool/composer.json", Name: "gone/pkg", Section: SectionRequire}, Spec: "^1.0",
	}}})
	if err == nil || !strings.Contains(err.Error(), "no longer declares gone/pkg") {
		t.Fatalf("%v", err)
	}
}

func TestLatestTakesTheNewestRelease(t *testing.T) {
	real := FetchJSON
	t.Cleanup(func() { FetchJSON = real })
	FetchJSON = func(url string, v any) error {
		switch {
		case strings.Contains(url, "private/"):
			return errors.New(url + ": 404 Not Found")
		case strings.Contains(url, "offline/"):
			return errors.New("no route")
		}
		return json.Unmarshal([]byte(`{"packages":{"monolog/monolog":[{"version":"4.0.0-beta1"},{"version":"dev-main"},{"version":"v3.9.0"},{"version":"3.8.1"}]}}`), v)
	}
	res, err := Latest("", protocol.LatestParams{Names: []string{"monolog/monolog", "private/pkg", "offline/pkg"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Latest["monolog/monolog"] != "3.9.0" || len(res.Latest) != 1 || res.Unreachable["offline/pkg"] != "no route" || len(res.Unreachable) != 1 {
		t.Fatalf("%+v", res)
	}
}
