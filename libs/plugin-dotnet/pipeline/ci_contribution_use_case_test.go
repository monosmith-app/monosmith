package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
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
