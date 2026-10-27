package dependencies

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const centralProps = `<Project>
  <PropertyGroup>
    <ManagePackageVersionsCentrally>true</ManagePackageVersionsCentrally>
  </PropertyGroup>
  <ItemGroup>
    <PackageVersion Include="Newtonsoft.Json" Version="13.0.1" />
    <PackageVersion Include="Serilog" Version="[3.0,4.0)" />
  </ItemGroup>
</Project>
`

const apiProject = `<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" />
    <PackageReference Include="Dapper" Version="2.0.123" PrivateAssets="all" />
    <PackageReference Include="Polly">
      <Version>7.2.3</Version>
    </PackageReference>
    <PackageReference Include="Floating" Version="1.*" />
    <PackageReference Include="Overridden" VersionOverride="2.0.0" />
  </ItemGroup>
</Project>
`

const apiTests = `<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="xunit" Version="2.4.2" />
    <PackageReference Include="Dapper" Version="2.1.35" />
  </ItemGroup>
</Project>
`

var api = protocol.Project{Name: "Api", Root: "src/Api", Tasks: map[string]protocol.Task{
	"test": {Command: []string{"dotnet", "test", "../../tests/Api.Tests/Api.Tests.csproj", "--nologo"}},
}}

func writeWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	for file, body := range map[string]string{
		"Directory.Packages.props":         centralProps,
		"src/Api/Api.csproj":               apiProject,
		"tests/Api.Tests/Api.Tests.csproj": apiTests,
	} {
		full := filepath.Join(ws, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func TestDeclaredListsProjectsTestsAndCentralVersions(t *testing.T) {
	res, err := Declared(writeWorkspace(t), protocol.DeclaredParams{Projects: []protocol.Project{api}})
	if err != nil {
		t.Fatal(err)
	}
	type row struct {
		Project, File, Name, Spec, Section string
		Rewritable                         bool
	}
	var got []row
	for _, d := range res.Declarations {
		got = append(got, row{d.Project, d.File, d.Name, d.Spec, d.Section, d.Rewritable})
	}
	want := []row{
		{"", "Directory.Packages.props", "Newtonsoft.Json", "13.0.1", "PackageVersion", true},
		{"", "Directory.Packages.props", "Serilog", "[3.0,4.0)", "PackageVersion", false},
		{"Api", "src/Api/Api.csproj", "Dapper", "2.0.123", "PackageReference", true},
		{"Api", "src/Api/Api.csproj", "Floating", "1.*", "PackageReference", false},
		{"Api", "src/Api/Api.csproj", "Polly", "7.2.3", "PackageReference", true},
		{"Api", "tests/Api.Tests/Api.Tests.csproj", "Dapper", "2.1.35", "PackageReference", true},
		{"Api", "tests/Api.Tests/Api.Tests.csproj", "xunit", "2.4.2", "PackageReference", true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("declarations:\n got %+v\nwant %+v", got, want)
	}
}

func TestRewriteChangesOnlyTheVersionText(t *testing.T) {
	ws := writeWorkspace(t)
	declared, _ := Declared(ws, protocol.DeclaredParams{Projects: []protocol.Project{api}})
	var changes []protocol.Rewrite
	for _, d := range declared.Declarations {
		switch d.Name {
		case "Newtonsoft.Json", "Dapper", "Polly":
			changes = append(changes, protocol.Rewrite{Declaration: d, Spec: "9.9.9"})
		}
	}
	res, err := Rewrite(ws, protocol.RewriteParams{Changes: changes})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 3 {
		t.Errorf("files: %v", res.Files)
	}
	read := func(rel string) string {
		raw, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		return string(raw)
	}
	project := read("src/Api/Api.csproj")
	for _, want := range []string{
		`<PackageReference Include="Dapper" Version="9.9.9" PrivateAssets="all" />`,
		"<Version>9.9.9</Version>",
		`<PackageReference Include="Floating" Version="1.*" />`,
	} {
		if !strings.Contains(project, want) {
			t.Errorf("project lacks %q:\n%s", want, project)
		}
	}
	if props := read("Directory.Packages.props"); !strings.Contains(props, `<PackageVersion Include="Newtonsoft.Json" Version="9.9.9" />`) || !strings.Contains(props, `Version="[3.0,4.0)"`) {
		t.Errorf("central versions:\n%s", props)
	}
	if tests := read("tests/Api.Tests/Api.Tests.csproj"); !strings.Contains(tests, `<PackageReference Include="Dapper" Version="9.9.9" />`) || !strings.Contains(tests, `Version="2.4.2"`) {
		t.Errorf("tests project:\n%s", tests)
	}
}

func TestSetVersionIgnoresCaseAndRefusesAGonePackage(t *testing.T) {
	src := `<PackageReference Include="newtonsoft.json" Version="1.0" />`
	if out, ok := SetVersion(src, ElementReference, "Newtonsoft.Json", "2.0"); !ok || !strings.Contains(out, `Version="2.0"`) {
		t.Fatalf("%v %s", ok, out)
	}
	if _, ok := SetVersion(src, ElementVersion, "Newtonsoft.Json", "2.0"); ok {
		t.Error("a reference is not a central version")
	}
	_, err := Rewrite(writeWorkspace(t), protocol.RewriteParams{Changes: []protocol.Rewrite{{
		Declaration: protocol.Declaration{File: "src/Api/Api.csproj", Name: "Gone", Section: ElementReference}, Spec: "1.0",
	}}})
	if err == nil || !strings.Contains(err.Error(), "no longer declares Gone") {
		t.Fatalf("%v", err)
	}
}

func TestLatestTakesTheNewestStableRelease(t *testing.T) {
	real := FetchJSON
	t.Cleanup(func() { FetchJSON = real })
	var asked []string
	FetchJSON = func(url string, v any) error {
		asked = append(asked, url)
		switch {
		case strings.Contains(url, "private"):
			return errors.New(url + ": 404 Not Found")
		case strings.Contains(url, "offline"):
			return errors.New("no route")
		}
		return json.Unmarshal([]byte(`{"versions":["1.0.0","13.0.3","14.0.0-beta.1"]}`), v)
	}
	res, err := Latest("", protocol.LatestParams{Names: []string{"Newtonsoft.Json", "private.pkg", "offline.pkg"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Latest["Newtonsoft.Json"] != "13.0.3" || len(res.Latest) != 1 || res.Unreachable["offline.pkg"] != "no route" || len(res.Unreachable) != 1 {
		t.Fatalf("%+v", res)
	}
	found := false
	for _, u := range asked {
		found = found || u == FlatContainer+"/newtonsoft.json/index.json"
	}
	if !found {
		t.Errorf("ids are lower-cased for nuget.org: %v", asked)
	}
}

func TestInstallAndUninstallRunDotnet(t *testing.T) {
	ws := writeWorkspace(t)
	var ran []string
	real := Dotnet
	t.Cleanup(func() { Dotnet = real })
	Dotnet = func(dir string, argv []string) error {
		ran = append(ran, filepath.Base(dir)+": "+strings.Join(argv, " "))
		return nil
	}
	_, _ = Install(ws, protocol.InstallParams{Projects: []protocol.Project{api}})
	_, _ = Install(ws, protocol.InstallParams{Project: &api, Packages: []string{"Serilog@3.1.1", "Polly"}})
	_, _ = Uninstall(ws, protocol.UninstallParams{Project: api, Packages: []string{"Polly"}})
	want := []string{
		"Api: restore",
		filepath.Base(ws) + ": add src/Api/Api.csproj package Serilog --version 3.1.1",
		filepath.Base(ws) + ": add src/Api/Api.csproj package Polly",
		filepath.Base(ws) + ": remove src/Api/Api.csproj package Polly",
	}
	if !reflect.DeepEqual(ran, want) {
		t.Fatalf("dotnet ran:\n%v", ran)
	}
}
