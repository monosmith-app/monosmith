package graphview

import (
	"strings"
	"testing"
)

func TestDiff(t *testing.T) {
	before := []Node{
		{Name: "web", Kind: "app", Language: "node", Root: "apps/web", Stack: "node-jest", Deps: []string{"ui"}},
		{Name: "ui", Kind: "lib", Language: "node", Root: "libs/ui"},
		{Name: "old", Kind: "lib", Language: "go", Root: "libs/old"},
	}
	after := []Node{
		{Name: "web", Kind: "app", Language: "node", Root: "apps/web", Stack: "node-vitest", Deps: []string{"ui", "sdk"}},
		{Name: "ui", Kind: "package", Language: "node", Root: "libs/ui"},
		{Name: "sdk", Kind: "package", Language: "node", Root: "packages/sdk"},
	}
	d := Diff(before, after)
	d.From, d.To = "v1", "the working tree"
	want := "project graph from v1 to the working tree\n" +
		"+ sdk (node package, packages/sdk)\n" +
		"- old (go lib, libs/old)\n" +
		"+ web → sdk\n" +
		"~ web stack: node-jest → node-vitest\n" +
		"~ ui kind: lib → package\n"
	if got := DiffText(d); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if !strings.Contains(DiffText(Diff(before, before)), "no change") {
		t.Fatal("an unchanged graph says so")
	}
}
