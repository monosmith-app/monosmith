package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const cliGemspec = `# frozen_string_literal: true

Gem::Specification.new do |spec|
  spec.name = 'acme_cli'
  spec.version = '1.2.0'
  spec.executables = ['acme_cli']
  spec.metadata['smith_kind'] = 'app'
  spec.metadata["smith_stack"] = "ruby"

  spec.add_dependency 'acme_core', '~> 0.0'
  spec.add_development_dependency("webmock")
end
`

func TestParseGemspec(t *testing.T) {
	g := ParseGemspec(cliGemspec)
	if g.Name != "acme_cli" || g.Version != "1.2.0" || g.Kind != "app" || g.Stack != "ruby" || g.Executables[0] != "acme_cli" || len(g.Dependencies) != 2 || g.Dependencies[0] != "acme_core" || g.Dependencies[1] != "webmock" {
		t.Fatalf("%+v", g)
	}
}

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("apps/cli/acme_cli.gemspec", cliGemspec)
	write("packages/core/acme_core.gemspec", "Gem::Specification.new do |s|\n  s.name = 'acme_core'\n  s.version = '0.0.0'\nend\n")
	write("packages/core/spec/core_spec.rb", "")
	res, err := Discover(ws, []string{"apps/cli/acme_cli.gemspec", "packages/core/acme_core.gemspec"})
	if err != nil || len(res.Projects) != 2 {
		t.Fatalf("%v %+v", err, res.Projects)
	}
	cli, core := res.Projects[0], res.Projects[1]
	if cli.Kind != protocol.ProjectApp || cli.Exports[0] != "acme_cli" || cli.PackageName != "acme_cli" || cli.Imports[0] != "acme_core" || cli.Tasks["start"].Command[3] != "exe/acme_cli" {
		t.Errorf("cli %+v", cli)
	}
	if _, ok := cli.Tasks["test"]; ok {
		t.Error("cli has no spec/")
	}
	if core.Kind != protocol.ProjectPackage || core.Tasks["test"].Command[2] != "rspec" || core.Tasks["lint"].Command[2] != "rubocop" {
		t.Errorf("core %+v", core)
	}
}
