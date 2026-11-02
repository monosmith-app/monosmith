package linking

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/dependencies"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAddArgsPointCargoAtTheLibraryFolder(t *testing.T) {
	want := []string{"add", "--path", "../../libs/core", "--manifest-path", "apps/web/Cargo.toml"}
	if got := AddArgs("apps/web", "libs/core"); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
	if got := PathTo("libs", "libs/core"); got != "./core" {
		t.Fatalf("got %q", got)
	}
}

func lab(t *testing.T, web string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range map[string]string{
		"apps/web/Cargo.toml":  web,
		"libs/core/Cargo.toml": "[package]\nname = \"shop-core\"\nversion = \"0.1.0\"\n",
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

func TestLinkRunsCargoAddUnlessTheDependencyIsThere(t *testing.T) {
	var ran [][]string
	saved := dependencies.Cargo
	dependencies.Cargo = func(_ string, argv []string) error { ran = append(ran, argv); return nil }
	defer func() { dependencies.Cargo = saved }()

	params := protocol.LinkParams{Project: protocol.Project{Name: "web", Root: "apps/web"}, Library: protocol.Project{Name: "core", Root: "libs/core", Exports: []string{"shop-core"}}}
	root := lab(t, "[package]\nname = \"web\"\nversion = \"0.1.0\"\n")
	res, err := Link(root, params)
	if err != nil || len(ran) != 1 || strings.Join(res.Files, ",") != "apps/web/Cargo.toml,Cargo.lock" {
		t.Fatalf("ran %v res %+v err %v", ran, res, err)
	}
	if res.Line != `shop-core = { path = "../../libs/core" }` {
		t.Fatalf("line %q", res.Line)
	}

	ran = nil
	root = lab(t, "[package]\nname = \"web\"\nversion = \"0.1.0\"\n\n[dependencies]\nshop-core = { path = \"../../libs/core\" }\n")
	res, err = Link(root, params)
	if err != nil || len(ran) != 0 || len(res.Notes) != 1 {
		t.Fatalf("an existing link must not run cargo: ran %v res %+v err %v", ran, res, err)
	}
	un, err := Unlink(root, protocol.UnlinkParams{Project: params.Project, Library: params.Library})
	if err != nil || len(ran) != 1 || ran[0][0] != "remove" || len(un.Files) == 0 {
		t.Fatalf("unlink ran %v %+v %v", ran, un, err)
	}
	un, err = Unlink(lab(t, "[package]\nname = \"web\"\n"), protocol.UnlinkParams{Project: params.Project, Library: params.Library})
	if err != nil || len(un.Files) != 0 {
		t.Fatalf("nothing linked must report no files: %+v %v", un, err)
	}
}

func TestLinkNamesTheMissingManifest(t *testing.T) {
	_, err := Link(t.TempDir(), protocol.LinkParams{Project: protocol.Project{Name: "gone", Root: "apps/gone"}, Library: protocol.Project{Root: "libs/x", Exports: []string{"x"}}})
	if err == nil || !strings.Contains(err.Error(), "apps/gone/Cargo.toml") {
		t.Fatalf("got %v", err)
	}
}
