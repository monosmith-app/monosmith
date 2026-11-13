package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("composer.json", `{"name": "acme/workspace", "require": {"acme/core": "@dev"}}`)
	write("apps/cli/composer.json", `{
  "name": "acme/cli", "version": "1.2.0", "bin": ["bin/cli"],
  "require": {"php": ">=8.2", "ext-json": "*", "acme/core": "*"},
  "require-dev": {"mockery/mockery": "^1.6"},
  "extra": {"smith": {"kind": "app", "stack": "php"}}
}`)
	write("packages/core/composer.json", `{"name": "acme/core", "version": "0.0.0"}`)
	write("packages/core/tests/CoreTest.php", "")
	res, err := Discover(ws, []string{"composer.json", "apps/cli/composer.json", "packages/core/composer.json"})
	if err != nil || len(res.Projects) != 2 {
		t.Fatalf("the root is not a project: %v %+v", err, res.Projects)
	}
	cli, core := res.Projects[0], res.Projects[1]
	if cli.Kind != protocol.ProjectApp || cli.Stack != "php" || cli.Version != "1.2.0" || len(cli.Imports) != 2 || cli.Imports[0] != "acme/core" || cli.Tasks["start"].Command[1] != "bin/cli" {
		t.Errorf("cli %+v", cli)
	}
	if _, ok := cli.Tasks["test"]; ok || cli.Tasks["lint"].Command[len(cli.Tasks["lint"].Command)-1] != "src" {
		t.Errorf("cli has no tests/ %+v", cli.Tasks)
	}
	if core.Kind != protocol.ProjectPackage || core.Exports[0] != "acme/core" || core.PackageName != "acme/core" || core.Tasks["test"].Command[1] != "{workspaceRoot}/vendor/bin/phpunit" {
		t.Errorf("core %+v", core)
	}
}
