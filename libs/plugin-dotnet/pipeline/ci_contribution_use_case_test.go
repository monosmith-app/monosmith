package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCI(t *testing.T) {
	ws := t.TempDir()
	core := protocol.Project{Name: "core", Root: "packages/core", Kind: protocol.ProjectPackage, Tasks: map[string]protocol.Task{"test": {Command: []string{"dotnet", "test", "tests/Acme.Core.Tests.csproj"}}}}
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{core}})
	if res.Toolchains[0].Version != SDKVersion || len(res.Setup) != 2 || res.Setup[1].Command[2] != "tests/Acme.Core.Tests.csproj" || res.Credentials[0].Env != "NUGET_API_KEY" {
		t.Fatalf("%+v", res)
	}
	_ = os.WriteFile(filepath.Join(ws, "global.json"), []byte(`{"sdk": {"version": "9.0.300"}}`), 0o644)
	if v := Version(ws); v != "9.0" {
		t.Fatalf("global.json pins %s", v)
	}
}

func TestCIAuditsEachProjectForVulnerablePackages(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "api", "Api.csproj"), []byte("<Project/>"), 0o644)
	api := protocol.Project{Name: "api", Root: "apps/api", Tasks: map[string]protocol.Task{"test": {Command: []string{"dotnet", "test", "../../tests/Api.Tests.csproj"}}}}
	got := CI(ws, protocol.CIParams{Projects: []protocol.Project{api}}).Audit
	if len(got) != 1 || got[0].Dir != "apps/api" || got[0].Blocking {
		t.Fatalf("one report-only step per project: %+v", got)
	}
	want := []string{"apps/api/Api.csproj", "tests/Api.Tests.csproj", "apps/api/packages.lock.json", "Directory.Packages.props", "Directory.Build.props"}
	if !reflect.DeepEqual(got[0].CacheFiles, want) {
		t.Errorf("the verdict is remembered per project file, lockfile and central versions: %v", got[0].CacheFiles)
	}
	if script := got[0].Command[2]; !strings.Contains(script, "dotnet list package --vulnerable --include-transitive") || !strings.Contains(script, "exit 1") {
		t.Errorf("dotnet only prints vulnerabilities, so the step must fail on them: %s", script)
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "", errors.New("missing") }
	res := Doctor(t.TempDir(), protocol.DoctorParams{Projects: []protocol.Project{{}}})
	if len(res.Findings) != 1 || res.Findings[0].OK {
		t.Fatalf("%+v", res)
	}
}

func TestCIDeclaresTheNuGetDownloadStore(t *testing.T) {
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{{Name: "api", Root: "src/api"}}}).Stores
	want := []protocol.PackageStore{{Name: "nuget", Path: "~/.nuget/packages", Env: "NUGET_PACKAGES", Files: []string{"**/*.csproj", "**/*.fsproj"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
}
