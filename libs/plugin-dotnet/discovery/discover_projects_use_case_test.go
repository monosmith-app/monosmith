package discovery

import (
	"os"
	"path/filepath"
	"reflect"
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
	write("apps/api/Acme.Api.csproj", `<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup><TargetFramework>net10.0</TargetFramework><OutputType>Exe</OutputType><Version>1.2.0</Version></PropertyGroup>
  <ItemGroup><ProjectReference Include="..\..\packages\core\Acme.Core.csproj" /></ItemGroup>
</Project>`)
	write("packages/core/Acme.Core.csproj", `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup><TargetFramework>net10.0</TargetFramework><PackageId>Acme.Core</PackageId><SmithStack>dotnet</SmithStack></PropertyGroup>
</Project>`)
	write("packages/core/tests/Acme.Core.Tests.csproj", `<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup><PackageReference Include="Microsoft.NET.Test.Sdk" Version="17.0.0" /></ItemGroup>
  <ItemGroup><ProjectReference Include="../Acme.Core.csproj" /><ProjectReference Include="../../../libs/fixtures/Acme.Fixtures.csproj" /></ItemGroup>
</Project>`)
	write("libs/fixtures/Acme.Fixtures.csproj", `<Project Sdk="Microsoft.NET.Sdk"><PropertyGroup><TargetFramework>net10.0</TargetFramework></PropertyGroup></Project>`)
	res, err := Discover(ws, []string{"apps/api/Acme.Api.csproj", "libs/fixtures/Acme.Fixtures.csproj", "packages/core/Acme.Core.csproj", "packages/core/tests/Acme.Core.Tests.csproj"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 3 {
		t.Fatalf("the tests project belongs to core: %+v", res.Projects)
	}
	api, fixtures, core := res.Projects[0], res.Projects[1], res.Projects[2]
	if api.Kind != protocol.ProjectApp || api.Version != "1.2.0" || !reflect.DeepEqual(api.Imports, []string{"Acme.Core"}) || api.Tasks["start"].Command[1] != "run" {
		t.Errorf("api %+v", api)
	}
	if fixtures.Kind != protocol.ProjectLib || fixtures.Exports[0] != "Acme.Fixtures" || fixtures.PackageName != "Acme.Fixtures" {
		t.Errorf("fixtures %+v", fixtures)
	}
	if core.Kind != protocol.ProjectPackage || core.Stack != "dotnet" || core.Tasks["test"].Command[2] != "tests/Acme.Core.Tests.csproj" || !reflect.DeepEqual(core.Imports, []string{"Acme.Fixtures"}) {
		t.Errorf("core %+v", core)
	}
	if pack := core.Tasks["pack"]; pack.Outputs[0] != "../../dist/packages/core" {
		t.Errorf("pack %+v", pack)
	}
	if _, ok := api.Tasks["test"]; ok {
		t.Error("a project without a tests project has no test task")
	}
}
