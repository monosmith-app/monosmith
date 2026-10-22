package sbom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func put(t *testing.T, root, rel, body string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(abs), 0o755)
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListReadsGemfileLockSpecsAndSkipsPathGems(t *testing.T) {
	root := t.TempDir()
	put(t, root, "Gemfile.lock", "PATH\n  remote: .\n  specs:\n    mine (0.1.0)\n\nGEM\n  remote: https://rubygems.org/\n  specs:\n    rack (3.0.8)\n      webrick (>= 1)\n    webrick (1.8.1)\n\nDEPENDENCIES\n  rack\n")
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "app"}})
	if err != nil || res.Source != "Gemfile.lock" || len(res.Components) != 2 || res.Components[0].PURL != "pkg:gem/rack@3.0.8" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestListFallsBackToGemfileAndGemspec(t *testing.T) {
	root := t.TempDir()
	put(t, root, "Gemfile", "source 'https://rubygems.org'\ngemspec\ngem 'rake', '~> 13.0'\n")
	put(t, root, "mine.gemspec", "Gem::Specification.new do |spec|\n  spec.add_dependency \"json\", \">= 2\"\nend\n")
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || len(res.Components) != 2 || res.Components[0].Name != "json" {
		t.Fatalf("%+v %v", res, err)
	}
}
