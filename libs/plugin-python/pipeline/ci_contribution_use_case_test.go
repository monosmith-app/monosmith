package pipeline

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCIPointsTheVenvsPipAtThePrivateIndexBeforeInstalling(t *testing.T) {
	feed := &protocol.Registry{Name: "libs", Auth: protocol.AuthAzureArtifacts, PyPI: "https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/simple/", PyPIUpload: "https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/upload/"}
	r := CI(protocol.CIParams{Projects: []protocol.Project{{Name: "sdk", Root: "packages/sdk", Kind: protocol.ProjectPackage, Registries: &protocol.Registries{Install: feed, Publish: feed}}}})
	if len(r.Setup) != 5 || r.Setup[2].Name != "Point pip at the private index" {
		t.Fatalf("setup: %+v", r.Setup)
	}
	if got := strings.Join(r.Setup[2].Command[1:], " "); got != "-m pip config --site set global.index-url "+feed.PyPI {
		t.Errorf("command %q", got)
	}
	for _, c := range r.Credentials {
		if c.Env == "PYPI_TOKEN" {
			t.Error("published to Azure Artifacts: no PyPI token")
		}
	}
	if len(r.Authenticate) != 2 {
		t.Errorf("sign-ins: %+v", r.Authenticate)
	}
}

func TestCIWithoutARegistryIsUnchanged(t *testing.T) {
	r := CI(protocol.CIParams{Projects: []protocol.Project{{Name: "sdk", Root: "packages/sdk", Kind: protocol.ProjectPackage}}})
	if len(r.Setup) != 4 || len(r.Authenticate) != 0 || len(r.Credentials) != 1 || r.Credentials[0].Env != "PYPI_TOKEN" {
		t.Errorf("result %+v", r)
	}
}

func TestCIShowsThePythonVersionAndKeepsTheResolversConflictReportOnProjectInstalls(t *testing.T) {
	r := CI(protocol.CIParams{Projects: []protocol.Project{{Name: "sdk", Root: "packages/sdk", Kind: protocol.ProjectPackage}}})
	var version, toolchain, project []string
	for _, s := range r.Setup {
		switch s.Name {
		case "Show the Python version":
			version = s.Command
		case "Install the Python toolchain":
			toolchain = s.Command
		case "Install sdk":
			project = s.Command
		}
	}
	if len(version) != 2 || version[1] != "--version" {
		t.Errorf("version step: %v", version)
	}
	if !strings.Contains(strings.Join(toolchain, " "), "--quiet") {
		t.Errorf("the toolchain install stays quiet: %v", toolchain)
	}
	if strings.Contains(strings.Join(project, " "), "--quiet") {
		t.Errorf("a project install must not hide the resolver's report: %v", project)
	}
}

func TestCIAddsThePythonPolicyStepOnlyWhenOptedIn(t *testing.T) {
	projects := []protocol.Project{{Name: "api", Root: "apps/api"}}
	if got := CI(protocol.CIParams{Projects: projects}).Audit; len(got) != 1 {
		t.Fatalf("off by default: %+v", got)
	}
	got := CI(protocol.CIParams{Projects: projects, Audit: protocol.AuditSettings{Policies: []string{"python"}}}).Audit
	want := protocol.Step{Name: "python policy (requirements)", Command: []string{"{smith}", "py-policy", "requirements-dev.txt", "apps/api/requirements.txt"}, Blocking: true, CacheFiles: []string{"requirements-dev.txt", "apps/api/requirements.txt"}}
	if len(got) != 2 || !reflect.DeepEqual(got[1], want) {
		t.Fatalf("%+v", got)
	}
	warn := protocol.AuditSettings{Policies: []string{"python"}, Warn: []string{"python"}}
	if CI(protocol.CIParams{Projects: projects, Audit: warn}).Audit[1].Blocking {
		t.Error("warn makes it report only")
	}
}

func TestCIPythonPolicyExpectsThePrivateIndex(t *testing.T) {
	feed := &protocol.Registry{Name: "libs", PyPI: "https://pkgs.example/simple/"}
	projects := []protocol.Project{{Name: "api", Root: "apps/api", Registries: &protocol.Registries{Install: feed}}}
	got := CI(protocol.CIParams{Projects: projects, Audit: protocol.AuditSettings{Policies: []string{"python"}}}).Audit[1].Command
	if got[2] != "--index=https://pkgs.example/simple/" {
		t.Errorf("%v", got)
	}
}

func TestCIDeclaresThePipCacheKeyedOnTheDependencyFiles(t *testing.T) {
	got := CI(protocol.CIParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}}}).Stores
	want := []protocol.PackageStore{{Name: "pip", Path: "~/.cache/pip", Env: "PIP_CACHE_DIR", Files: []string{"requirements-dev.txt", "apps/api/pyproject.toml"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
}
