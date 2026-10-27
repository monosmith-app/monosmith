package kindchange

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func edited(r protocol.ChangeKindResult) map[string]string {
	out := map[string]string{}
	for _, f := range r.Files {
		out[f.Path] = f.Content
	}
	return out
}

const coreGemspec = `# frozen_string_literal: true

Gem::Specification.new do |spec|
  spec.name = 'core'
  spec.version = '0.0.0'
  spec.files = Dir['lib/**/*.rb']
  spec.require_paths = ['lib']
  spec.metadata['rubygems_mfa_required'] = 'true'
  # Released by its tag, never pushed to a gem server.
  spec.metadata['allowed_push_host'] = 'none'
  spec.metadata['smith_kind'] = 'lib'
end
`

var workspace = map[string]string{
	"Gemfile":                  "source 'https://rubygems.org'\n\ngemspec path: 'libs/core'\ngemspec path: 'apps/cli'\n\ngroup :development do\nend\n",
	"Gemfile.lock":             "PATH\n  remote: libs/core\n  specs:\n    core (0.0.0)\n",
	"libs/core/core.gemspec":   coreGemspec,
	"apps/cli/cli.gemspec":     "Gem::Specification.new do |spec|\n  spec.name = 'cli'\n  spec.add_dependency 'core'\n  spec.metadata['smith_kind'] = 'app'\nend\n",
	"apps/cli/Gemfile":         "gem 'core', path: '../../libs/core'\n",
	"libs/other/other.gemspec": "Gem::Specification.new do |spec|\n  spec.name = 'other'\nend\n",
	"libs/other/Gemfile":       "gem 'core', path: '../core'\n",
}

func projects() []protocol.Project {
	return []protocol.Project{
		{Name: "core", Root: "libs/core", Language: "ruby"},
		{Name: "cli", Root: "apps/cli", Language: "ruby"},
		{Name: "other", Root: "libs/other", Language: "ruby"},
	}
}

func TestLibBecomesPackage(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{Project: projects()[0], To: protocol.ProjectPackage, NewRoot: "packages/core", Projects: projects()})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	core := got["libs/core/core.gemspec"]
	if strings.Contains(core, "allowed_push_host") || strings.Contains(core, "Released by its tag") || !strings.Contains(core, "'smith_kind'] = 'package'") {
		t.Fatalf("the push guard goes and the kind changes:\n%s", core)
	}
	if !strings.Contains(got["Gemfile"], "gemspec path: 'packages/core'") || !strings.Contains(got["Gemfile"], "gemspec path: 'apps/cli'") {
		t.Fatalf("the Gemfile lists the new directory:\n%s", got["Gemfile"])
	}
	if !strings.Contains(got["Gemfile.lock"], "remote: packages/core") {
		t.Fatalf("the lockfile follows:\n%s", got["Gemfile.lock"])
	}
	if !strings.Contains(got["apps/cli/Gemfile"], "path: '../../packages/core'") || !strings.Contains(got["libs/other/Gemfile"], "path: '../../packages/core'") {
		t.Fatalf("dependents follow:\n%v", got)
	}
	joined := strings.Join(r.Warnings, "\n")
	if !strings.Contains(joined, "description") || !strings.Contains(joined, "homepage") {
		t.Fatalf("rubygems.org's page needs are named: %v", r.Warnings)
	}
}

func TestLibToPackageAndBackLeavesTheGemspecByteForByte(t *testing.T) {
	dir := writeFiles(t, map[string]string{"libs/core/core.gemspec": coreGemspec})
	project := protocol.Project{Name: "core", Root: "libs/core", Language: "ruby"}
	for _, to := range []protocol.ProjectKind{protocol.ProjectPackage, protocol.ProjectLib} {
		r, err := Change(dir, protocol.ChangeKindParams{Project: project, To: to, NewRoot: "libs/core"})
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range r.Files {
			if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(e.Path)), []byte(e.Content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "libs/core/core.gemspec")); string(got) != coreGemspec {
		t.Fatalf("round trip changed the gemspec:\n%s", got)
	}
}

func TestAppToLibKeepsTheExecutablesAndSaysSo(t *testing.T) {
	files := map[string]string{"apps/tool/tool.gemspec": "Gem::Specification.new do |spec|\n  spec.executables = ['tool']\n  spec.metadata['smith_kind'] = 'app'\nend\n"}
	dir := writeFiles(t, files)
	r, err := Change(dir, protocol.ChangeKindParams{Project: protocol.Project{Name: "tool", Root: "apps/tool", Language: "ruby"}, To: protocol.ProjectLib, NewRoot: "libs/tool"})
	if err != nil {
		t.Fatal(err)
	}
	spec := edited(r)["apps/tool/tool.gemspec"]
	if !strings.Contains(spec, "'smith_kind'] = 'lib'") || !strings.Contains(spec, "allowed_push_host'] = 'none'") {
		t.Fatalf("a lib is kept off the gem server:\n%s", spec)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "executables") {
		t.Fatalf("the executables are named: %v", r.Warnings)
	}
}
