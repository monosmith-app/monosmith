package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

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

func TestProjectIDIsAValidDotNetIdentifier(t *testing.T) {
	for _, c := range []struct{ workspace, name, want string }{
		{"acme", "web-api", "Acme.WebApi"},
		{"jato.auto.packages", "core", "JatoAutoPackages.Core"},
		{"my shop", "api", "MyShop.Api"},
		{"3d", "2fa", "X3d.X2fa"},
		{"", "core", "Core"},
	} {
		if got := ProjectID(c.workspace, c.name); got != c.want {
			t.Errorf("ProjectID(%q, %q) = %q, want %q", c.workspace, c.name, got, c.want)
		}
	}
}
