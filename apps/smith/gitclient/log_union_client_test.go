package gitclient

import (
	"reflect"
	"testing"
)

func TestTagsMergedAtAndLogUnionSeeTheOtherBranch(t *testing.T) {
	g := newTestRepo(t)
	checkout := func(args ...string) {
		t.Helper()
		if _, err := g.Run(append([]string{"checkout", "-q"}, args...)...); err != nil {
			t.Fatal(err)
		}
	}
	write(t, g, "a", "1")
	_ = g.CommitAll("feat: first")
	if err := g.CreateTag("p@1.0.0", "p"); err != nil {
		t.Fatal(err)
	}
	checkout("-b", "dev")
	write(t, g, "b", "2")
	_ = g.CommitAll("fix: on dev")
	checkout("main")
	write(t, g, "c", "3")
	_ = g.CommitAll("feat: on main")
	if err := g.CreateTag("p@1.1.0", "p"); err != nil {
		t.Fatal(err)
	}
	checkout("dev")

	if tags, _ := g.TagsMerged("p@"); !reflect.DeepEqual(tags, []string{"p@1.0.0"}) {
		t.Fatalf("HEAD contains %v", tags)
	}
	if tags, err := g.TagsMergedAt("p@", "main"); err != nil || !reflect.DeepEqual(tags, []string{"p@1.0.0", "p@1.1.0"}) {
		t.Fatalf("main contains %v %v", tags, err)
	}
	if _, err := g.TagsMergedAt("p@", "nope"); err == nil {
		t.Fatal("an unknown ref is an error")
	}
	union, err := g.LogUnion("p@1.0.0", "main")
	if err != nil || len(union) != 2 || union[0].Subject == union[1].Subject {
		t.Fatalf("union %+v %v", union, err)
	}
	if only, _ := g.LogUnion("p@1.1.0", "main"); len(only) != 1 || only[0].Subject != "fix: on dev" {
		t.Fatalf("since the newer tag only dev's commit is left: %+v", only)
	}
	if all, _ := g.LogUnion("", "main", "c"); len(all) != 1 || all[0].Subject != "feat: on main" {
		t.Fatalf("path filter %+v", all)
	}
}
