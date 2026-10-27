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

const gemfile = `source 'https://rubygems.org'

gemspec path: 'apps/web'

gem 'rake', '~> 13.0'
gem "puma"

group :development, :test do
  gem 'rspec', '~> 3.12'
  gem 'rubocop', '~> 1.50', require: false
end

gem 'rack', '>= 2.0', '< 4'
`

const webSpec = `Gem::Specification.new do |spec|
  spec.name = 'web'
  spec.add_dependency 'rack', '~> 3.0.1'
  spec.add_runtime_dependency("sinatra", ">= 3.0")
  spec.add_dependency 'zeitwerk'
  spec.add_development_dependency 'rspec', '~> 3.13.0'
end
`

var projects = []protocol.Project{{Name: "web", Root: "apps/web"}}

func writeWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	for file, body := range map[string]string{"Gemfile": gemfile, "apps/web/web.gemspec": webSpec} {
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

func TestDeclaredListsGemsWithAConstraint(t *testing.T) {
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
		{"", "rack", ">= 2.0, < 4", "dependencies", false},
		{"", "rake", "~> 13.0", "dependencies", true},
		{"", "rspec", "~> 3.12", "development_dependencies", true},
		{"", "rubocop", "~> 1.50", "development_dependencies", true},
		{"web", "rack", "~> 3.0.1", "dependencies", true},
		{"web", "rspec", "~> 3.13.0", "development_dependencies", true},
		{"web", "sinatra", ">= 3.0", "dependencies", true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("declarations:\n got %+v\nwant %+v", got, want)
	}
}

func TestConstraintKeepsTheTildePrecision(t *testing.T) {
	for _, c := range []struct{ old, wanted, want string }{
		{"~> 1.2", "~>1.4.1", "~> 1.4"},
		{"~> 1.2.3", "~>1.4.1", "~> 1.4.1"},
		{"~> 3", "~>3.2.1", "~> 3"},
		{">= 1.0", ">=2.1.0", ">= 2.1.0"},
		{"1.2.3", "1.4.1", "1.4.1"},
	} {
		if got := Constraint(c.old, c.wanted); got != c.want {
			t.Errorf("Constraint(%q, %q) = %q, want %q", c.old, c.wanted, got, c.want)
		}
	}
}

func TestRewriteChangesOnlyTheConstraintString(t *testing.T) {
	ws := writeWorkspace(t)
	declared, _ := Declared(ws, protocol.DeclaredParams{Projects: projects})
	var changes []protocol.Rewrite
	for _, d := range declared.Declarations {
		if d.Rewritable && (d.Name == "rspec" || d.Name == "rubocop" || (d.Name == "rack" && d.Project == "web")) {
			changes = append(changes, protocol.Rewrite{Declaration: d, Spec: "~>9.9.9"})
		}
	}
	res, err := Rewrite(ws, protocol.RewriteParams{Changes: changes})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 2 {
		t.Errorf("files: %v", res.Files)
	}
	read := func(rel string) string {
		raw, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		return string(raw)
	}
	g := read("Gemfile")
	for _, want := range []string{"gem 'rspec', '~> 9.9'", "gem 'rubocop', '~> 9.9', require: false", "gem 'rake', '~> 13.0'", "gem 'rack', '>= 2.0', '< 4'"} {
		if !strings.Contains(g, want) {
			t.Errorf("Gemfile lacks %q:\n%s", want, g)
		}
	}
	s := read("apps/web/web.gemspec")
	for _, want := range []string{"spec.add_dependency 'rack', '~> 9.9.9'", "spec.add_development_dependency 'rspec', '~> 9.9.9'", `spec.add_runtime_dependency("sinatra", ">= 3.0")`} {
		if !strings.Contains(s, want) {
			t.Errorf("gemspec lacks %q:\n%s", want, s)
		}
	}
}

func TestRewriteRefusesAGemThatIsGone(t *testing.T) {
	_, err := Rewrite(writeWorkspace(t), protocol.RewriteParams{Changes: []protocol.Rewrite{{
		Declaration: protocol.Declaration{File: "Gemfile", Name: "gone", Spec: "~> 1.0"}, Spec: "~> 2.0",
	}}})
	if err == nil || !strings.Contains(err.Error(), "no longer declares gone") {
		t.Fatalf("%v", err)
	}
}

func TestLatestAsksRubyGems(t *testing.T) {
	real := FetchJSON
	t.Cleanup(func() { FetchJSON = real })
	FetchJSON = func(url string, v any) error {
		switch {
		case strings.Contains(url, "/private/"):
			return errors.New(url + ": 404 Not Found")
		case strings.Contains(url, "/offline/"):
			return errors.New("no route")
		}
		return json.Unmarshal([]byte(`{"version":"3.1.0"}`), v)
	}
	res, err := Latest("", protocol.LatestParams{Names: []string{"rack", "private", "offline"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Latest["rack"] != "3.1.0" || len(res.Latest) != 1 || res.Unreachable["offline"] != "no route" || len(res.Unreachable) != 1 {
		t.Fatalf("%+v", res)
	}
}
