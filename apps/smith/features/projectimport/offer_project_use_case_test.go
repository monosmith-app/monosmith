package projectimport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/features/manifestoffer"
)

var dotnetClaims = []adoptmove.Claim{{Language: "dotnet", Manifests: []string{"*.csproj"}, Sources: []string{"*.cs"}}}

func TestChooseNestedRefusesWithThePathThatAnswers(t *testing.T) {
	one := []manifestoffer.Nested{{Path: "src/Lib", Manifest: "Lib.csproj", Language: "dotnet"}}
	_, err := chooseNested(one, Options{Source: "../lib"}, false, nil)
	if err == nil || !strings.Contains(err.Error(), "smith import ../lib --path src/Lib") {
		t.Fatalf("got %v", err)
	}
	two := append(one, manifestoffer.Nested{Path: "test/Lib.Tests", Manifest: "Lib.Tests.csproj", Language: "dotnet"})
	_, err = chooseNested(two, Options{Source: "../lib"}, false, nil)
	if err == nil || !strings.Contains(err.Error(), "so it is a workspace") || !strings.Contains(err.Error(), "smith adopt") {
		t.Fatalf("got %v", err)
	}
}

func TestOfferRefusesAFolderOfSourcesWithoutTheFlag(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Sorter.cs"), []byte("class A {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := Source{Arg: dir, Dir: dir}
	if _, _, err := offer(&src, dotnetClaims, Options{Source: dir}); err == nil || !strings.Contains(err.Error(), "--manifest dotnet") {
		t.Fatalf("got %v", err)
	}
	c, write, err := offer(&src, dotnetClaims, Options{Source: dir, Manifest: "dotnet"})
	if err != nil || write != "dotnet" || c.Ecosystem != "dotnet" || c.Name == "" {
		t.Fatalf("got %+v %q %v", c, write, err)
	}
	if _, _, err := offer(&src, dotnetClaims, Options{Source: dir, Manifest: "none"}); err == nil || !strings.Contains(err.Error(), "nothing was imported") {
		t.Fatalf("--manifest none gave %v", err)
	}
}

func TestLayoutLeftNamesTheSourcesOutsideSrc(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "libs", "tools"), 0o755); err != nil {
		t.Fatal(err)
	}
	p := Plan{Name: "tools", Folder: "libs/tools", WriteManifest: "node"}
	if left := LayoutLeft(root, p); len(left) != 1 || !strings.Contains(left[0].Problem, "point at src/") {
		t.Fatalf("got %+v", left)
	}
	if left := LayoutLeft(root, Plan{Name: "tools", Folder: "libs/tools", WriteManifest: "python"}); left != nil {
		t.Errorf("python fits its manifest to the layout: %+v", left)
	}
}
