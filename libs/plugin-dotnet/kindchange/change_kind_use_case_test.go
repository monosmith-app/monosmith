package kindchange

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func edited(r protocol.ChangeKindResult) map[string]string {
	out := map[string]string{}
	for _, f := range r.Files {
		out[f.Path] = f.Content
	}
	return out
}

const coreProject = `<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net9.0</TargetFramework>
    <RootNamespace>Acme.Core</RootNamespace>
    <AssemblyName>Acme.Core</AssemblyName>
    <IsPackable>false</IsPackable>
    <Version>0.0.0</Version>
    <SmithKind>lib</SmithKind>
  </PropertyGroup>

  <ItemGroup>
    <ProjectReference Include="..\util\Acme.Util.csproj" />
  </ItemGroup>

</Project>
`

var workspace = map[string]string{
	"Acme.sln": "Project(\"{FAE04EC0}\") = \"Acme.Core\", \"libs\\core\\Acme.Core.csproj\", \"{1}\"\nEndProject\n" +
		"Project(\"{FAE04EC0}\") = \"Acme.Core.Tests\", \"libs\\core\\tests\\Greeting\\Acme.Core.Tests.csproj\", \"{2}\"\nEndProject\n" +
		"Project(\"{FAE04EC0}\") = \"Acme.Util\", \"libs\\util\\Acme.Util.csproj\", \"{3}\"\nEndProject\n",
	"libs/core/Acme.Core.csproj": coreProject,
	"libs/core/tests/Greeting/Acme.Core.Tests.csproj": "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <ItemGroup>\n" +
		"    <ProjectReference Include=\"../../Acme.Core.csproj\" />\n    <ProjectReference Include=\"../../../util/Acme.Util.csproj\" />\n  </ItemGroup>\n</Project>\n",
	"libs/util/Acme.Util.csproj": "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <IsPackable>false</IsPackable>\n    <SmithKind>lib</SmithKind>\n  </PropertyGroup>\n</Project>\n",
	"apps/cli/Acme.Cli.csproj": "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <OutputType>Exe</OutputType>\n    <SmithKind>app</SmithKind>\n  </PropertyGroup>\n  <ItemGroup>\n" +
		"    <ProjectReference Include=\"..\\..\\libs\\core\\Acme.Core.csproj\" />\n  </ItemGroup>\n</Project>\n",
}

func projects() []protocol.Project {
	return []protocol.Project{
		{Name: "core", Root: "libs/core", Language: "dotnet"},
		{Name: "util", Root: "libs/util", Language: "dotnet"},
		{Name: "cli", Root: "apps/cli", Language: "dotnet"},
	}
}

func TestLibBecomesPackage(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: projects()[0], To: protocol.ProjectPackage, NewRoot: "packages/core", Projects: projects(),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	core := got["libs/core/Acme.Core.csproj"]
	if strings.Contains(core, "IsPackable") || !strings.Contains(core, "<SmithKind>package</SmithKind>") {
		t.Fatalf("the project is packable and its kind changes:\n%s", core)
	}
	if !strings.Contains(core, `Include="..\..\libs\util\Acme.Util.csproj"`) {
		t.Fatalf("its own reference still reaches util from the new directory, with its slashes:\n%s", core)
	}
	tests := got["libs/core/tests/Greeting/Acme.Core.Tests.csproj"]
	if !strings.Contains(tests, `"../../Acme.Core.csproj"`) || !strings.Contains(tests, `"../../../../libs/util/Acme.Util.csproj"`) {
		t.Fatalf("the tests project keeps its reference to core and recomputes the one to util:\n%s", tests)
	}
	if !strings.Contains(got["apps/cli/Acme.Cli.csproj"], `"..\..\packages\core\Acme.Core.csproj"`) {
		t.Fatalf("dependents follow:\n%s", got["apps/cli/Acme.Cli.csproj"])
	}
	sln := got["Acme.sln"]
	if !strings.Contains(sln, `"packages\core\Acme.Core.csproj"`) || !strings.Contains(sln, `"packages\core\tests\Greeting\Acme.Core.Tests.csproj"`) || !strings.Contains(sln, `"libs\util\Acme.Util.csproj"`) {
		t.Fatalf("the solution lists the new directory and keeps util:\n%s", sln)
	}
	if _, touched := got["libs/util/Acme.Util.csproj"]; touched {
		t.Fatal("a project that does not reference core is not edited")
	}
	joined := strings.Join(r.Warnings, "\n")
	if !strings.Contains(joined, "a description") || !strings.Contains(joined, "a license") {
		t.Fatalf("nuget.org's requirements are named: %v", r.Warnings)
	}
}

func TestAppToLibKeepsTheProgramAndSaysSo(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: projects()[2], To: protocol.ProjectLib, NewRoot: "libs/cli", Projects: projects(),
	})
	if err != nil {
		t.Fatal(err)
	}
	cli := edited(r)["apps/cli/Acme.Cli.csproj"]
	if !strings.Contains(cli, "<IsPackable>false</IsPackable>") || !strings.Contains(cli, "<SmithKind>lib</SmithKind>") {
		t.Fatalf("a lib is not packable:\n%s", cli)
	}
	if !strings.Contains(cli, strings.ReplaceAll(`"../core/Acme.Core.csproj"`, "/", string(rune(92)))) {
		t.Fatalf("its reference to core is rewritten for the new directory:\n%s", cli)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "program") {
		t.Fatalf("the program is named: %v", r.Warnings)
	}
}

func TestLibToPackageAndBackLeavesTheProjectByteForByte(t *testing.T) {
	dir := writeFiles(t, map[string]string{"libs/core/Acme.Core.csproj": coreProject})
	project := protocol.Project{Name: "core", Root: "libs/core", Language: "dotnet"}
	for _, to := range []protocol.ProjectKind{protocol.ProjectPackage, protocol.ProjectLib} {
		r, err := Change(dir, protocol.ChangeKindParams{Project: project, To: to, NewRoot: "libs/core"})
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range r.Files {
			if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(e.Path)), []byte(e.Content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "libs/core/Acme.Core.csproj")); string(got) != coreProject {
		t.Fatalf("round trip changed the project:\n%s", got)
	}
}
