package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestLintRunsPHPStanOnSmithsDefaultUnlessThePackageConfiguresIt(t *testing.T) {
	def := Tasks(Shape{Kind: protocol.ProjectLib, PHPStan: []string{"-c", "{workspaceRoot}/.smith/defaults/phpstan.neon"}})["lint"]
	got := strings.Join(def.Command, " ")
	if !strings.Contains(got, "analyse --no-progress --memory-limit=1G -c {workspaceRoot}/.smith/defaults/phpstan.neon --autoload-file") || strings.Contains(got, "--level") {
		t.Fatalf("command: %s", got)
	}
	if last := def.Inputs[len(def.Inputs)-1]; last != "{workspaceRoot}/.smith/defaults/phpstan.neon" {
		t.Fatalf("the default's content is an input: %v", def.Inputs)
	}
	own := strings.Join(Tasks(Shape{Kind: protocol.ProjectLib})["lint"].Command, " ")
	if strings.Contains(own, " -c ") || strings.Contains(own, "--level") {
		t.Fatalf("the package's phpstan.neon decides, level included: %s", own)
	}
}

func TestDiscoverPassesTheDefaultPHPStanConfigToEachPackage(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("packages/core/composer.json", `{"name": "acme/core"}`)
	write("packages/own/composer.json", `{"name": "acme/own"}`)
	write("packages/own/phpstan.neon", "parameters:\n    level: 5\n")
	res, err := Discover(ws, []string{"packages/core/composer.json", "packages/own/composer.json"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(res.Projects[0].Tasks["lint"].Command, " "); !strings.Contains(got, "-c {workspaceRoot}/.smith/defaults/phpstan.neon") {
		t.Errorf("core: %s", got)
	}
	if got := strings.Join(res.Projects[1].Tasks["lint"].Command, " "); strings.Contains(got, " -c ") {
		t.Errorf("own: %s", got)
	}
}
