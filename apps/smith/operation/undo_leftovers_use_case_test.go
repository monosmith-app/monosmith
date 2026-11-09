package operation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUndoNamesTheGeneratedFoldersItLeftAndRemovesThemOnRequest(t *testing.T) {
	g := scratch(t, true)
	execute(t, "web")
	put(t, g.Dir, ".gitignore", ".smith/\nnode_modules/\ndist/\n")
	if err := g.CommitAll("chore: ignore generated folders"); err != nil {
		t.Fatal(err)
	}
	// What an install and a build leave in the project the operation made.
	put(t, g.Dir, "libs/web/node_modules/zod/index.js", "x\n")
	put(t, g.Dir, "libs/web/dist/main.js", "x\n")
	// A second project holds a file of the person's own: never listed.
	put(t, g.Dir, "libs/mine/notes.txt", "mine\n")
	put(t, g.Dir, "libs/mine/node_modules/a/b.js", "x\n")

	c, err := Prepare(g.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Undo(g.Dir, c); err != nil {
		t.Fatal(err)
	}
	var removed []string
	for _, l := range c.Changes {
		removed = append(removed, l.Path)
	}
	left := FindLeftovers(g.Dir, append(removed, "libs/mine/notes.txt"))
	if len(left) != 1 || left[0].Dir != "libs/web" {
		t.Fatalf("leftovers %+v", left)
	}
	if got := strings.Join(left[0].Entries, ","); !strings.Contains(got, "node_modules") || !strings.Contains(got, "dist") {
		t.Errorf("entries %q", got)
	}
	if err := RemoveLeftovers(g.Dir, left); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(g.Dir, "libs", "web")); err == nil {
		t.Error("the folder stayed")
	}
	if _, err := os.Stat(filepath.Join(g.Dir, "libs", "mine", "notes.txt")); err != nil {
		t.Error("the person's file was touched")
	}
}

func TestLeftoverLineHasNoStrayComma(t *testing.T) {
	cases := map[string]Leftover{
		"  apps/web/  (node_modules/)":        {Dir: "apps/web", Entries: []string{"", "node_modules/"}},
		"  apps/web/  (dist/, node_modules/)": {Dir: "apps/web", Entries: []string{"dist/", "node_modules/"}},
		"  apps/web/":                         {Dir: "apps/web", Entries: []string{""}},
	}
	for want, l := range cases {
		if got := leftoverLine(l); got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}
