package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-go/linting"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "go.mod", "module example.com/root\n")
	write(t, ws, "apps/api/go.mod", "module example.com/acme/apps/api\n\ngo 1.24\n")
	write(t, ws, "apps/api/main.go", "package main\nimport (\n\t\"fmt\"\n\t\"example.com/acme/libs/core/greeting\"\n)\nfunc main(){fmt.Println(greeting.X)}\n")
	write(t, ws, "apps/api/testdata/x.go", "package x\nimport \"example.com/ignored\"\n")
	write(t, ws, "apps/api/nested/go.mod", "module example.com/nested\n")
	write(t, ws, "apps/api/nested/n.go", "package n\nimport \"example.com/nested-only\"\n")
	write(t, ws, "packages/sdk/go.mod", "// smith:kind package\nmodule example.com/acme/packages/sdk\n")

	res, err := Discover(ws, []string{"go.mod", "apps/api/go.mod", "packages/sdk/go.mod"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 2 {
		t.Fatalf("projects %+v", res.Projects)
	}
	api, sdk := res.Projects[0], res.Projects[1]
	if api.Kind != protocol.ProjectApp || api.Exports[0] != "example.com/acme/apps/api" {
		t.Fatalf("api %+v", api)
	}
	if !reflect.DeepEqual(api.Imports, []string{"example.com/acme/libs/core/greeting", "fmt"}) {
		t.Fatalf("api imports %v", api.Imports)
	}
	if api.Tasks["build"].Outputs[0] != "../../dist/apps/api" || api.Tasks["start"].Command[1] != "run" {
		t.Fatalf("api tasks %+v", api.Tasks)
	}
	if sdk.Kind != protocol.ProjectPackage || sdk.TagPrefix != "packages/sdk/v" {
		t.Fatalf("sdk %+v", sdk)
	}
	if check := sdk.Tasks["deps-check"]; !reflect.DeepEqual(check.Command, []string{"{smith}", "go-deps-check", "packages/sdk"}) || !check.Cache || sdk.Tasks["lint"].DependsOn[0] != "deps-check" {
		t.Fatalf("a package checks its imports before lint: %+v", check)
	}
	if _, ok := api.Tasks["deps-check"]; ok {
		t.Fatal("an app has no deps-check")
	}
	if _, ok := sdk.Tasks["start"]; ok {
		t.Fatal("a package has no start task")
	}
	lint := api.Tasks["lint"].Command
	if lint[0] != "golangci-lint" || lint[len(lint)-2] != "--config" {
		t.Fatalf("a module without a golangci config lints with smith's default: %v", lint)
	}
}

func TestARepositoryGolangciConfigReplacesTheDefault(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, ".golangci.yml", "version: \"2\"\n")
	write(t, ws, "apps/api/go.mod", "module example.com/acme/apps/api\n")
	write(t, ws, "apps/api/main.go", "package main\nfunc main(){}\n")
	res, err := Discover(ws, []string{"apps/api/go.mod"})
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range res.Projects[0].Tasks["lint"].Command {
		if a == "--config" {
			t.Fatalf("smith passes no --config over the repository's own: %v", res.Projects[0].Tasks["lint"].Command)
		}
	}
}

func TestParseModule(t *testing.T) {
	if _, err := ParseModule("go 1.24\n"); err == nil {
		t.Fatal("expected an error")
	}
	m, err := ParseModule("// smith:kind lib\nmodule \"example.com/q\"\n")
	if err != nil || m.Path != "example.com/q" || m.Kind != "lib" {
		t.Fatalf("%+v %v", m, err)
	}
}

func TestANativeModuleIsTaggedAndBuildsWithCgo(t *testing.T) {
	mod, err := ParseModule("// smith:kind app\n// smith:native\nmodule example.com/probe\n")
	if err != nil || !mod.Native || mod.Kind != "app" {
		t.Fatalf("%+v %v", mod, err)
	}
	if plain, _ := ParseModule("module example.com/x\n"); plain.Native {
		t.Fatal("a module without the comment is not native")
	}
	for name, task := range Tasks("probe", "apps/probe", protocol.ProjectApp, true, linting.Lint{}) {
		if task.Env["CGO_ENABLED"] != "1" {
			t.Errorf("%s does not build with cgo", name)
		}
	}
}
