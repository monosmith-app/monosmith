package linking

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/dependencies"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestReferenceIncludeIsRelativeWithBackslashes(t *testing.T) {
	if got := ReferenceInclude("apps/Web/Web.csproj", "libs/Core/Core.csproj"); got != `..\..\libs\Core\Core.csproj` {
		t.Fatalf("got %q", got)
	}
	want := []string{"add", "apps/Web/Web.csproj", "reference", "libs/Core/Core.csproj"}
	if got := AddArgs("apps/Web/Web.csproj", "libs/Core/Core.csproj"); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func lab(t *testing.T, web string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range map[string]string{
		"apps/Web/Web.csproj":   web,
		"libs/Core/Core.csproj": "<Project Sdk=\"Microsoft.NET.Sdk\"></Project>",
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestLinkRunsDotnetAddReferenceUnlessItIsThere(t *testing.T) {
	var ran [][]string
	saved := dependencies.Dotnet
	dependencies.Dotnet = func(_ string, argv []string) error { ran = append(ran, argv); return nil }
	defer func() { dependencies.Dotnet = saved }()

	params := protocol.LinkParams{Project: protocol.Project{Name: "web", Root: "apps/Web"}, Library: protocol.Project{Name: "core", Root: "libs/Core"}}
	res, err := Link(lab(t, "<Project Sdk=\"Microsoft.NET.Sdk\"></Project>"), params)
	if err != nil || len(ran) != 1 || res.Manifest != "apps/Web/Web.csproj" || len(res.Files) != 1 {
		t.Fatalf("ran %v res %+v err %v", ran, res, err)
	}

	ran = nil
	linked := "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <ItemGroup>\n    <ProjectReference Include=\"..\\..\\libs\\Core\\Core.csproj\" />\n  </ItemGroup>\n</Project>"
	root := lab(t, linked)
	res, err = Link(root, params)
	if err != nil || len(ran) != 0 || len(res.Notes) != 1 {
		t.Fatalf("an existing reference must not run dotnet: %v %+v %v", ran, res, err)
	}
	un, err := Unlink(root, protocol.UnlinkParams{Project: params.Project, Library: params.Library})
	if err != nil || len(ran) != 1 || ran[0][0] != "remove" || len(un.Files) != 1 {
		t.Fatalf("unlink ran %v %+v %v", ran, un, err)
	}
	un, err = Unlink(lab(t, "<Project Sdk=\"Microsoft.NET.Sdk\"></Project>"), protocol.UnlinkParams{Project: params.Project, Library: params.Library})
	if err != nil || len(un.Files) != 0 {
		t.Fatalf("nothing linked must report no files: %+v %v", un, err)
	}
}

func TestLinkNamesTheFolderWithoutAProjectFile(t *testing.T) {
	_, err := Link(t.TempDir(), protocol.LinkParams{Project: protocol.Project{Name: "gone", Root: "apps/gone"}, Library: protocol.Project{Root: "libs/x"}})
	if err == nil || !strings.Contains(err.Error(), "apps/gone") {
		t.Fatalf("got %v", err)
	}
}
