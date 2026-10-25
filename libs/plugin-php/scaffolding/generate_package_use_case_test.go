package scaffolding

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-php/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	var ran []string
	real := Runner
	defer func() { Runner = real }()
	Runner = func(_ string, argv []string) error { ran = append(ran, strings.Join(argv, " ")); return nil }
	info := protocol.WorkspaceInfo{Name: "acme-corp"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "core", Directory: "packages/core", Workspace: info, Stack: "php"}); err != nil {
		t.Fatal(err)
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web-cli", Directory: "apps/web-cli", Workspace: info})
	if err != nil || !strings.Contains(strings.Join(res.Files, ","), "composer.json,") {
		t.Fatal(err, res.Files)
	}
	if ran[3] != "composer config repositories.apps-web-cli path apps/web-cli" || ran[4] != "composer require acme-corp/web-cli:@dev --no-update --no-interaction" || len(ran) != 6 {
		t.Fatalf("%q", ran)
	}
	manifest, _ := os.ReadFile(filepath.Join(ws, "apps/web-cli/composer.json"))
	if !strings.Contains(string(manifest), `"AcmeCorp\\WebCli\\": "src/"`) || !strings.Contains(string(manifest), `"bin": ["bin/web-cli"]`) {
		t.Fatalf("composer.json:\n%s", manifest)
	}
	bin, _ := os.ReadFile(filepath.Join(ws, "apps/web-cli/bin/web-cli"))
	if !strings.Contains(string(bin), "use AcmeCorp\\WebCli\\Greeting\\GreetUseCase;") || !strings.Contains(string(bin), "__DIR__ . '/../../../vendor/autoload.php'") {
		t.Fatalf("bin:\n%s", bin)
	}
	if ignore, _ := os.ReadFile(filepath.Join(ws, ".gitignore")); strings.Count(string(ignore), "/vendor/") != 1 {
		t.Fatalf(".gitignore:\n%s", ignore)
	}
	found, err := discovery.Discover(ws, []string{"composer.json", "packages/core/composer.json", "apps/web-cli/composer.json"})
	if err != nil || len(found.Projects) != 2 {
		t.Fatalf("%v %+v", err, found.Projects)
	}
	core, cli := found.Projects[0], found.Projects[1]
	if core.Kind != protocol.ProjectPackage || core.Stack != "php" || core.Exports[0] != "acme-corp/core" || core.Tasks["test"].Command == nil {
		t.Errorf("core %+v", core)
	}
	if cli.Kind != protocol.ProjectApp || cli.Tasks["start"].Command[1] != "bin/web-cli" {
		t.Errorf("cli %+v", cli)
	}
	ran = nil
	removed, err := Remove(ws, protocol.RemoveParams{Project: cli})
	if err != nil || strings.Join(ran, "|") != "composer remove acme-corp/web-cli --no-update --no-interaction|composer config --unset repositories.apps-web-cli|composer update --no-interaction --no-progress" {
		t.Fatal(err, ran)
	}
	if strings.Join(removed.Files, ",") != "composer.json,composer.lock" {
		t.Errorf("the lockfile is rewritten too: %v", removed.Files)
	}
}

func TestComposerOutputIsHiddenUnlessItFailsOrVerbose(t *testing.T) {
	t.Setenv(toolrun.VerboseEnv, "")
	var log bytes.Buffer
	previous := toolrun.Log
	toolrun.Log = &log
	t.Cleanup(func() { toolrun.Log = previous })
	if err := runComposer(t.TempDir(), []string{"go", "version"}); err != nil {
		t.Fatal(err)
	}
	if text := log.String(); !strings.Contains(text, "Running go version") || strings.Contains(text, "go version go") {
		t.Fatalf("the tool's output must not stream when it succeeds: %q", text)
	}
}
