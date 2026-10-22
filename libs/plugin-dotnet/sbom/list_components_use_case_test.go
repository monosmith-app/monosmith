package sbom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func put(t *testing.T, root, rel, body string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(abs), 0o755)
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListReadsThePackagesLock(t *testing.T) {
	root := t.TempDir()
	put(t, root, "app/packages.lock.json", `{"version":1,"dependencies":{"net10.0":{"Newtonsoft.Json":{"type":"Direct","resolved":"13.0.3"},"Lib":{"type":"Project"}}}}`)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "app"}})
	if err != nil || len(res.Components) != 1 || res.Components[0].PURL != "pkg:nuget/Newtonsoft.Json@13.0.3" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestListFallsBackToTheProjectFile(t *testing.T) {
	root := t.TempDir()
	put(t, root, "app/App.csproj", `<Project Sdk="Microsoft.NET.Sdk"><ItemGroup>
<PackageReference Include="Serilog" Version="3.1.1" />
<PackageReference Include="Dapper"><Version>2.1.0</Version></PackageReference>
<ProjectReference Include="../lib/Lib.csproj" />
</ItemGroup></Project>`)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "app"}})
	if err != nil || res.Source != "app/App.csproj" || len(res.Components) != 2 || res.Components[0].Name != "Dapper" || res.Components[1].PURL != "pkg:nuget/Serilog@3.1.1" {
		t.Fatalf("%+v %v", res, err)
	}
}
