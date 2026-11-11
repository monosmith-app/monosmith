package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func dotnetWorkspace(t *testing.T, files map[string]string) string {
	t.Helper()
	ws := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

const csproj = "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup>\n</Project>\n"

var appProject = []protocol.Project{{Name: "app", Root: "src/App"}}

func TestDotnetBuildDirMergesTheArtifactsBlockAndKeepsTheOtherLines(t *testing.T) {
	ws := dotnetWorkspace(t, map[string]string{
		"src/App/App.csproj":    csproj,
		"Directory.Build.props": "<Project>\n  <PropertyGroup>\n    <Nullable>enable</Nullable>\n  </PropertyGroup>\n</Project>\n",
	})
	findings, ok := CheckBuildDir(ws, appProject)
	if !ok || len(findings) != 1 || findings[0].OK || findings[0].Warning {
		t.Fatalf("%+v", findings)
	}
	f := findings[0]
	if !strings.Contains(f.Detail, "app builds into bin/ and obj/ of each project, not into dist/dotnet/") {
		t.Fatalf("names where it builds now and where it should: %s", f.Detail)
	}
	move, leave := f.Fix.Choices[0].Options[0].Fix, f.Fix.Choices[0].Options[1].Fix
	if len(leave.Files) != 0 || len(move.Files) != 1 || move.Files[0].Path != "Directory.Build.props" {
		t.Fatalf("%+v", f.Fix)
	}
	want := "<Project>\n  <PropertyGroup>\n    <Nullable>enable</Nullable>\n  </PropertyGroup>\n  <!-- Every build goes to the workspace's dist/ (smith). -->\n  <PropertyGroup>\n    <UseArtifactsOutput>true</UseArtifactsOutput>\n    <ArtifactsPath>$(MSBuildThisFileDirectory)dist/dotnet</ArtifactsPath>\n  </PropertyGroup>\n</Project>\n"
	if move.Files[0].Content != want {
		t.Fatalf("got\n%s", move.Files[0].Content)
	}
}

func TestDotnetBuildDirWritesTheFileWhenThereIsNone(t *testing.T) {
	ws := dotnetWorkspace(t, map[string]string{"src/App/App.csproj": csproj})
	findings, _ := CheckBuildDir(ws, appProject)
	file := findings[0].Fix.Choices[0].Options[0].Fix.Files[0]
	if file.Path != "Directory.Build.props" || !strings.Contains(file.Content, "<ArtifactsPath>$(MSBuildThisFileDirectory)dist/dotnet</ArtifactsPath>") {
		t.Fatalf("%+v", file)
	}
}

func TestDotnetBuildDirUpdatesAnotherArtifactsPath(t *testing.T) {
	ws := dotnetWorkspace(t, map[string]string{
		"src/App/App.csproj":    csproj,
		"Directory.Build.props": "<Project>\n  <PropertyGroup>\n    <ArtifactsPath>$(MSBuildThisFileDirectory)out</ArtifactsPath>\n    <Nullable>enable</Nullable>\n  </PropertyGroup>\n</Project>\n",
	})
	findings, _ := CheckBuildDir(ws, appProject)
	if !strings.Contains(findings[0].Detail, "builds into out/") {
		t.Fatalf("%s", findings[0].Detail)
	}
	got := findings[0].Fix.Choices[0].Options[0].Fix.Files[0].Content
	if !strings.Contains(got, "<ArtifactsPath>$(MSBuildThisFileDirectory)dist/dotnet</ArtifactsPath>") || !strings.Contains(got, "<Nullable>enable</Nullable>") {
		t.Fatalf("%s", got)
	}
}

func TestDotnetBuildDirPassesWhenArtifactsPathIsInDist(t *testing.T) {
	ws := dotnetWorkspace(t, map[string]string{
		"src/App/App.csproj":    csproj,
		"Directory.Build.props": "<Project><PropertyGroup><ArtifactsPath>$(MSBuildThisFileDirectory)dist/dotnet</ArtifactsPath></PropertyGroup></Project>",
	})
	if f, _ := CheckBuildDir(ws, appProject); len(f) != 1 || !f[0].OK {
		t.Fatalf("%+v", f)
	}
}

func TestDotnetBuildDirRefusesWithAWarningAndTheReason(t *testing.T) {
	for name, tc := range map[string]struct {
		files map[string]string
		want  string
	}{
		"an explicit OutputPath": {
			files: map[string]string{"src/App/App.csproj": strings.Replace(csproj, "<TargetFramework>", "<OutputPath>out\\</OutputPath><TargetFramework>", 1)},
			want:  "src/App/App.csproj sets OutputPath or BaseOutputPath itself",
		},
		"a Dockerfile that copies bin": {
			files: map[string]string{"src/App/App.csproj": csproj, "src/App/Dockerfile": "COPY bin/Release/net8.0/publish/ .\n"},
			want:  "src/App/Dockerfile names the bin/ or obj/ folder of app",
		},
		"a reference by path": {
			files: map[string]string{
				"src/App/App.csproj":     csproj,
				"src/Other/Other.csproj": "<Project><ItemGroup><Reference Include=\"App\"><HintPath>..\\App\\bin\\Release\\App.dll</HintPath></Reference></ItemGroup></Project>",
			},
			want: "references a built file of app by its path",
		},
		"a props file that is not XML": {
			files: map[string]string{"src/App/App.csproj": csproj, "Directory.Build.props": "<Project><PropertyGroup>"},
			want:  "not valid XML",
		},
	} {
		t.Run(name, func(t *testing.T) {
			ws := dotnetWorkspace(t, tc.files)
			projects := append([]protocol.Project{}, appProject...)
			if _, ok := tc.files["src/Other/Other.csproj"]; ok {
				projects = append(projects, protocol.Project{Name: "other", Root: "src/Other"})
			}
			findings, _ := CheckBuildDir(ws, projects)
			f := findings[0]
			if f.OK || !f.Warning || f.Fix != nil || !strings.Contains(f.Remedy, tc.want) {
				t.Fatalf("a warning with the reason and no fix: %+v", f)
			}
		})
	}
}

func TestDotnetBuildDirLeavesAChainedPropsFileUnchecked(t *testing.T) {
	ws := dotnetWorkspace(t, map[string]string{
		"src/App/App.csproj":    csproj,
		"Directory.Build.props": "<Project><Import Project=\"$([MSBuild]::GetPathOfFileAbove('Directory.Build.props', '$(MSBuildThisFileDirectory)../'))\" /></Project>",
	})
	if f, ok := CheckBuildDir(ws, appProject); ok {
		t.Fatalf("nothing is guessed: %+v", f)
	}
}

func TestDotnetBuildDirMovesAnArtifactsPathInDistThatIsNotTheSharedWorkFolder(t *testing.T) {
	ws := dotnetWorkspace(t, map[string]string{
		"src/App/App.csproj":    csproj,
		"Directory.Build.props": "<Project><PropertyGroup><ArtifactsPath>$(MSBuildThisFileDirectory)dist/apps</ArtifactsPath></PropertyGroup></Project>",
	})
	f, _ := CheckBuildDir(ws, appProject)
	if len(f) != 1 || f[0].OK || f[0].Warning || f[0].Fix == nil || !strings.Contains(f[0].Detail, "shipped folder stays empty") {
		t.Fatalf("%+v", f)
	}
}

func TestDotnetBuildDirRefusesTheSharedArtifactsFolderAsATaskOutput(t *testing.T) {
	ws := dotnetWorkspace(t, map[string]string{
		"src/App/App.csproj":    csproj,
		"Directory.Build.props": "<Project><PropertyGroup><ArtifactsPath>$(MSBuildThisFileDirectory)dist/dotnet</ArtifactsPath></PropertyGroup></Project>",
	})
	app := []protocol.Project{{Name: "app", Root: "src/App", Tasks: map[string]protocol.Task{
		"build": {Outputs: []string{"../../dist/dotnet/bin"}, Ship: []string{"../../dist/dotnet/bin/App/debug/*"}},
	}}}
	f, _ := CheckBuildDir(ws, app)
	if len(f) != 2 || !f[0].OK || f[1].OK || f[1].Problem != "caches dist/dotnet/bin, a work folder other projects share" {
		t.Fatalf("%+v", f)
	}
}
