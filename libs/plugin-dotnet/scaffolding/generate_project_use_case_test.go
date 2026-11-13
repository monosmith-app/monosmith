package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerateInheritsTheFrameworkFromTheWorkspace(t *testing.T) {
	for name, tc := range map[string]struct {
		props    string
		explicit bool
	}{
		"new workspace":            {"", false},
		"props sets the framework": {"<Project><PropertyGroup>\n<TargetFramework>net9.0</TargetFramework></PropertyGroup></Project>", false},
		"props sets none":          {"<Project>\n</Project>", true},
	} {
		ws := t.TempDir()
		if tc.props != "" {
			if err := os.WriteFile(filepath.Join(ws, "Directory.Build.props"), []byte(tc.props), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "acme"}}); err != nil {
			t.Fatal(err)
		}
		for _, rel := range []string{"libs/core/Acme.Core.csproj", "libs/core/tests/Acme.Core.Tests.csproj"} {
			raw, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
			if got := strings.Contains(string(raw), "<TargetFramework>"); got != tc.explicit {
				t.Errorf("%s: %s sets its own framework = %v, want %v\n%s", name, rel, got, tc.explicit, raw)
			}
		}
		if tc.props == "" {
			if props, _ := os.ReadFile(filepath.Join(ws, "Directory.Build.props")); !strings.Contains(string(props), "<TargetFramework>net10.0</TargetFramework>") {
				t.Errorf("%s: Directory.Build.props:\n%s", name, props)
			}
		}
	}
}

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	info := protocol.WorkspaceInfo{Name: "acme"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web-api", Directory: "apps/web-api", Workspace: info}); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "core", Directory: "packages/core", Workspace: info, Stack: "dotnet"}); err != nil {
		t.Fatal(err)
	}
	props, _ := os.ReadFile(filepath.Join(ws, "Directory.Build.props"))
	if !strings.Contains(string(props), "<ArtifactsPath>$(MSBuildThisFileDirectory)dist/dotnet</ArtifactsPath>") {
		t.Fatalf("Directory.Build.props:\n%s", props)
	}
	manifests := []string{"apps/web-api/Acme.WebApi.csproj", "apps/web-api/tests/Acme.WebApi.Tests.csproj", "packages/core/Acme.Core.csproj", "packages/core/tests/Acme.Core.Tests.csproj"}
	res, err := discovery.Discover(ws, manifests)
	if err != nil || len(res.Projects) != 2 {
		t.Fatalf("%v %+v", err, res.Projects)
	}
	api, core := res.Projects[0], res.Projects[1]
	if api.Kind != protocol.ProjectApp || api.Exports[0] != "Acme.WebApi" || api.Tasks["test"].Command[2] != "tests/Acme.WebApi.Tests.csproj" {
		t.Errorf("api %+v", api)
	}
	if core.Kind != protocol.ProjectPackage || core.Stack != "dotnet" || core.Exports[0] != "Acme.Core" {
		t.Errorf("core %+v", core)
	}
	if program, _ := os.ReadFile(filepath.Join(ws, "apps", "web-api", "Program.cs")); !strings.Contains(string(program), "using Acme.WebApi.Greeting;") {
		t.Errorf("Program.cs:\n%s", program)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "packages/core", Workspace: info}); err == nil {
		t.Fatal("an existing directory is refused")
	}
}
