package scaffold

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// mentionsRepo makes a temp git repository with the given tracked files.
func mentionsRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	return dir
}

func TestFindMentionsGroupsByKindAndSkipsNoise(t *testing.T) {
	dir := mentionsRepo(t, map[string]string{
		".github/workflows/release.yml": "name: r\nrun: build ./apps/mvd-tui\n",
		"mvd.code-workspace":            "{\n\"label\": \"mvd-tui:qa\"\n}\n",
		"docs/guide.md":                 "intro\nSee mvd-tui here.\n",
		"apps/mvd-tui/main.go":          "// mvd-tui itself\n",
		"package-lock.json":             "mvd-tui\n",
		"other/readme.md":               "apps/mvd-tui-extra is another thing\n",
		"smith.toml":                    "mvd-tui\n",
		"image.bin":                     "mvd-tui\x00\x01",
	})
	got, rootOnly := FindMentions(dir, "mvd-tui", "apps/mvd-tui", []string{"smith.toml"})
	if rootOnly {
		t.Fatal("a long name is matched as a word")
	}
	var lines []string
	for _, m := range got {
		lines = append(lines, m.Kind+" "+m.File+":"+string(rune('0'+m.Line)))
	}
	want := ".github/workflows/release.yml:2,mvd.code-workspace:2,docs/guide.md:2,other/readme.md:1"
	if strings.Join(trimKind(lines), ",") != want {
		t.Fatalf("got %v", lines)
	}
	if got[0].Kind != MentionWorkflow || got[1].Kind != MentionEditor || got[2].Kind != MentionOther {
		t.Fatalf("%+v", got)
	}
}

// trimKind drops the kind prefix; the order is what the test checks.
func trimKind(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l[strings.Index(l, " ")+1:]
	}
	return out
}

func TestFindMentionsShortNameMatchesTheRootOnly(t *testing.T) {
	dir := mentionsRepo(t, map[string]string{
		"docs/a.md":      "the web app\n",
		".vscode/t.json": "cd apps/web && go run .\n",
	})
	got, rootOnly := FindMentions(dir, "web", "apps/web", nil)
	if !rootOnly || len(got) != 1 || got[0].File != ".vscode/t.json" || got[0].Kind != MentionEditor {
		t.Fatalf("%v %+v", rootOnly, got)
	}
}

func TestFindMentionsOutsideARepositoryIsEmpty(t *testing.T) {
	got, _ := FindMentions(t.TempDir(), "mvd-tui", "apps/mvd-tui", nil)
	if len(got) != 0 {
		t.Fatalf("%+v", got)
	}
}

func TestWriteMentionsBoundsTheListAndSaysSmithKeepsOut(t *testing.T) {
	r := Removal{Project: "mvd-tui", Root: "apps/mvd-tui"}
	for i := 0; i < 25; i++ {
		r.Mentions = append(r.Mentions, Mention{File: "docs/f" + string(rune('a'+i)) + ".md", Line: 1, Kind: MentionOther})
	}
	var b bytes.Buffer
	WriteMentions(&b, r)
	out := b.String()
	if !strings.Contains(out, "Still mentions mvd-tui") || !strings.Contains(out, "and 5 more") || !strings.Contains(out, "does not change") {
		t.Fatal(out)
	}
	if strings.Count(out, "docs/f") != 20 {
		t.Fatal(out)
	}
	b.Reset()
	WriteMentions(&b, Removal{Project: "x"})
	if b.Len() != 0 {
		t.Fatal(b.String())
	}
}
