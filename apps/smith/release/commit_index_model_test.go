package release

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

func indexRepo(t *testing.T) *gitclient.Git {
	t.Helper()
	g := gitclient.New(t.TempDir())
	if err := g.Init(); err != nil {
		t.Skip("git is not available: ", err)
	}
	for _, kv := range [][2]string{{"user.email", "t@example.com"}, {"user.name", "T"}, {"commit.gpgsign", "false"}, {"tag.gpgsign", "false"}} {
		_, _ = g.Run("config", kv[0], kv[1])
	}
	return g
}

func commitFile(t *testing.T, g *gitclient.Git, rel, message string) {
	t.Helper()
	p := filepath.Join(g.Dir, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	old, _ := os.ReadFile(p)
	if err := os.WriteFile(p, append(old, 'x'), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := g.CommitAll(message); err != nil {
		t.Fatal(err)
	}
}

func hashes(cs []gitclient.Commit) []string {
	var out []string
	for _, c := range cs {
		out = append(out, c.Hash)
	}
	return out
}

// The index must answer what git log answers, for each folder and tag.
func TestTheCommitIndexAnswersWhatGitLogAnswers(t *testing.T) {
	g := indexRepo(t).WithRefCache()
	commitFile(t, g, "a/x.txt", "feat(a): one")
	commitFile(t, g, "b/y.txt", "feat(b): one")
	if err := g.CreateTag("a@1.0.0", "a"); err != nil {
		t.Fatal(err)
	}
	commitFile(t, g, "a/x.txt", "fix(a): two")
	if err := g.CreateTag("b@1.0.0", "b"); err != nil {
		t.Fatal(err)
	}
	commitFile(t, g, "b/y.txt", "fix(b): two")
	commitFile(t, g, "a/z/deep.txt", "feat(a): three")

	index := newCommitIndex(g, []string{"a@1.0.0", "b@1.0.0"})
	if index == nil {
		t.Fatal("no index")
	}
	for _, c := range []struct {
		tag   string
		roots []string
	}{
		{"a@1.0.0", []string{"a"}}, {"a@1.0.0", []string{"b"}}, {"b@1.0.0", []string{"a", "b"}},
		{"b@1.0.0", []string{"a/z"}}, {"a@1.0.0", []string{""}},
	} {
		if !index.answers(c.tag) {
			t.Fatalf("%s: the index cannot answer", c.tag)
		}
		want, err := g.Log(c.tag, c.roots...)
		if err != nil {
			t.Fatal(err)
		}
		if got := index.log(c.tag, c.roots); !reflect.DeepEqual(hashes(got), hashes(want)) {
			t.Errorf("%s %v: index %v, git %v", c.tag, c.roots, hashes(got), hashes(want))
		}
	}
	// No tag yet: every commit of the folder.
	all := newCommitIndex(g, []string{""})
	want, _ := g.Log("", "b")
	if got := all.log("", []string{"b"}); !reflect.DeepEqual(hashes(got), hashes(want)) {
		t.Errorf("no tag: index %v, git %v", hashes(got), hashes(want))
	}
}

// A cached tag list is forgotten when a tag is created.
func TestTheRefCacheSeesATagCreatedThroughIt(t *testing.T) {
	g := indexRepo(t).WithRefCache()
	commitFile(t, g, "a/x.txt", "feat(a): one")
	if tags, _ := g.Tags("a@"); len(tags) != 0 {
		t.Fatalf("tags %v", tags)
	}
	if err := g.CreateTag("a@1.0.0", "a"); err != nil {
		t.Fatal(err)
	}
	if tags, _ := g.Tags("a@"); !reflect.DeepEqual(tags, []string{"a@1.0.0"}) {
		t.Fatalf("tags after creating one: %v", tags)
	}
	if g.TagCommit("a@1.0.0") == "" {
		t.Fatal("the tag has no commit")
	}
}
