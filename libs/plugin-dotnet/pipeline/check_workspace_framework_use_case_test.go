package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCheckWorkspaceFrameworkOffersToShareTheFramework(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		path := filepath.Join(ws, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	csproj := "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>net10.0</TargetFramework>\n  </PropertyGroup>\n</Project>\n"
	write("libs/core/Acme.Core.csproj", csproj)
	write("libs/core/tests/Acme.Core.Tests.csproj", csproj)
	write("Directory.Build.props", "<Project>\n  <PropertyGroup>\n    <Nullable>enable</Nullable>\n  </PropertyGroup>\n</Project>\n")
	f, ok := CheckWorkspaceFramework(ws, []protocol.Project{{Name: "core", Root: "libs/core"}})
	if !ok || !f.Warning || f.Fix == nil || f.Fix.Safe || len(f.Fix.Files) != 3 {
		t.Fatalf("want an unsafe fix editing the props and two projects, got %+v", f)
	}
	if f.Fix.Files[0].Path != "Directory.Build.props" || !strings.Contains(f.Fix.Files[0].Content, "<TargetFramework>net10.0</TargetFramework>") || strings.Contains(f.Fix.Files[1].Content, "TargetFramework") {
		t.Fatalf("files %+v", f.Fix.Files)
	}
	write("libs/core/Acme.Core.csproj", "<Project Sdk=\"Microsoft.NET.Sdk\" />")
	write("libs/core/tests/Acme.Core.Tests.csproj", "<Project Sdk=\"Microsoft.NET.Sdk\" />")
	if _, ok := CheckWorkspaceFramework(ws, []protocol.Project{{Name: "core", Root: "libs/core"}}); ok {
		t.Fatal("projects that inherit need no finding")
	}
}
