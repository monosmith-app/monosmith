package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "pyproject.toml", "[tool.ruff]\nline-length = 100\n")
	write(t, ws, "apps/api/pyproject.toml", "[project]\nname = \"acme-api\"\nversion = \"0.0.0\"\ndependencies = [\"Acme-Core>=0.1\", \"requests\"]\n\n[tool.smith]\nkind = \"app\"\n")
	write(t, ws, "apps/api/src/acme_api/__main__.py", "import os, sys\nfrom acme_utils.x import y\nfrom . import local\n")
	write(t, ws, "apps/api/build/lib/stale.py", "import stale\n")
	write(t, ws, "packages/sdk/pyproject.toml", "[project]\nname = \"acme.sdk\"\nversion = \"1.2.0\"\n")

	res, err := Discover(ws, []string{"pyproject.toml", "apps/api/pyproject.toml", "packages/sdk/pyproject.toml"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 2 {
		t.Fatalf("projects %+v", res.Projects)
	}
	api, sdk := res.Projects[0], res.Projects[1]
	if api.Kind != protocol.ProjectApp || api.Exports[0] != "acme_api" {
		t.Fatalf("api %+v", api)
	}
	if !reflect.DeepEqual(api.Imports, []string{"acme_core", "acme_utils", "os", "requests", "sys"}) {
		t.Fatalf("api imports %v", api.Imports)
	}
	if start := api.Tasks["start"].Command; start[len(start)-1] != "acme_api" || !strings.Contains(start[0], ".venv") {
		t.Fatalf("start %v", start)
	}
	if check := sdk.Tasks["deps-check"]; strings.Join(check.Command, " ") != "{smith} py-deps-check packages/sdk" || !check.Cache || sdk.Tasks["lint"].DependsOn[0] != "deps-check" {
		t.Fatalf("a package checks its imports before lint: %+v", check)
	}
	if _, ok := api.Tasks["deps-check"]; ok {
		t.Fatal("an app has no deps-check")
	}
	if sdk.Kind != protocol.ProjectPackage || sdk.Exports[0] != "acme_sdk" || sdk.Tasks["build"].Outputs[0] != "../../dist/packages/sdk" || sdk.Version != "1.2.0" {
		t.Fatalf("sdk %+v", sdk)
	}
	if lint := api.Tasks["lint"].Command; len(lint) != 5 || lint[3] != "check" {
		t.Fatalf("the root pyproject.toml configures ruff, so smith passes no --config: %v", lint)
	}
}

func TestModuleNames(t *testing.T) {
	got := ModuleNames("import a.b, c\n    from d.e import f\nfrom .rel import g\n# import nope\n")
	if !reflect.DeepEqual(got, []string{"a", "c", "d"}) {
		t.Fatalf("got %v", got)
	}
}

func TestToolsRunOnSmithsDefaultsUnlessTheRepositoryConfiguresThem(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "libs/typed/pyproject.toml", "[project]\nname = \"typed\"\n\n[tool.mypy]\nstrict = true\n\n[tool.pytest.ini_options]\naddopts = \"-x\"\n")
	write(t, ws, "libs/plain/pyproject.toml", "[project]\nname = \"plain\"\n")
	res, err := Discover(ws, []string{"libs/plain/pyproject.toml", "libs/typed/pyproject.toml"})
	if err != nil {
		t.Fatal(err)
	}
	plain, typed := res.Projects[0].Tasks, res.Projects[1].Tasks
	if cmd := strings.Join(plain["typecheck"].Command[1:], " "); cmd != "-m mypy --config-file {workspaceRoot}/.smith/defaults/mypy.ini ." {
		t.Fatalf("typecheck %s", cmd)
	}
	if cmd := strings.Join(plain["test"].Command[1:], " "); cmd != "-m pytest -q -p no:cacheprovider -c {workspaceRoot}/.smith/defaults/pytest.ini --rootdir=." {
		t.Fatalf("test %s", cmd)
	}
	for task, file := range map[string]string{"lint": "ruff.toml", "typecheck": "mypy.ini", "test": "pytest.ini"} {
		in := plain[task].Inputs
		if in[len(in)-1] != "{workspaceRoot}/.smith/defaults/"+file {
			t.Errorf("%s: the default's content is an input: %v", task, in)
		}
	}
	if cmd := typed["typecheck"].Command; len(cmd) != 4 || cmd[2] != "mypy" || cmd[3] != "." {
		t.Fatalf("a project that configures mypy gets no --config-file: %v", cmd)
	}
	if cmd := typed["test"].Command; len(cmd) != 6 {
		t.Fatalf("a project that configures pytest gets no -c: %v", cmd)
	}
	if strings.Contains(typed["typecheck"].Description, "smith's default") {
		t.Fatalf("description %q", typed["typecheck"].Description)
	}
}

func TestTheProjectsDescriptionIsTheProjectsDescription(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/api/pyproject.toml", "[project]\nname = \"acme-api\"\nversion = \"0.0.0\"\ndescription = \"Serves the API.\"\n")
	res, err := Discover(ws, []string{"apps/api/pyproject.toml"})
	if err != nil || len(res.Projects) != 1 || res.Projects[0].Description != "Serves the API." {
		t.Fatalf("%+v %v", res, err)
	}
}
