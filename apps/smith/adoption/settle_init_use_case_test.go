package adoption

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/adoptdecision"
	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
)

// rootProjectRepo makes a clean git repository with a project at its root.
func rootProjectRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range map[string]string{"package.json": `{"name":"solo","version":"1.0.0"}`, "index.js": "x\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "init"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func TestSettleInitRefusesBeforeAnyChangeAndNamesEveryMissingFlag(t *testing.T) {
	dir := rootProjectRepo(t)
	r := Report{Root: dir, Name: "solo", Clean: true}
	cases := []struct {
		name  string
		s     InitSettings
		wants []string
	}{
		{"move and answers missing", InitSettings{}, []string{"--move packages/solo", "--gitattributes", "--agent", "--ci"}},
		{"only the answers missing", InitSettings{Move: MoveOptions{Dir: "packages/solo"}}, []string{"--gitattributes", "--agent", "--ci"}},
		{"only --yes missing", InitSettings{
			Move:  MoveOptions{Dir: "packages/solo"},
			Flags: adoptdecision.Flags{Gitattributes: "skip", Agent: "skip", CI: []string{"none"}, CIGiven: true},
		}, []string{"--yes applies the plan"}},
	}
	for _, c := range cases {
		c.s.In, c.s.Out = bufio.NewReader(strings.NewReader("")), io.Discard
		_, err := SettleInit(r, c.s)
		if err == nil {
			t.Fatalf("%s: no refusal", c.name)
		}
		for _, want := range c.wants {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: refusal lacks %q:\n%v", c.name, want, err)
			}
		}
		if _, statErr := os.Stat(filepath.Join(dir, "packages")); statErr == nil {
			t.Errorf("%s: a folder was created before the refusal", c.name)
		}
	}
}

func TestSettleInitWithEveryAnswerSettlesTheMove(t *testing.T) {
	dir := rootProjectRepo(t)
	got, err := SettleInit(Report{Root: dir, Name: "solo", Clean: true}, InitSettings{
		Move:  MoveOptions{Dir: "packages/solo"},
		Flags: adoptdecision.Flags{Gitattributes: "skip", Agent: "skip", CI: []string{"none"}, CIGiven: true},
		Yes:   true, In: bufio.NewReader(strings.NewReader("")), Out: io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Move == nil || got.Move.Dir != "packages/solo" {
		t.Fatalf("move: %+v", got.Move)
	}
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
		t.Error("settling moved a file")
	}
}

func TestRollbackPutsTheTreeBackAfterAMove(t *testing.T) {
	dir := rootProjectRepo(t)
	r := Report{Root: dir, Name: "solo", Clean: true}
	c, plan, err := ChooseMove(r, MoveOptions{Dir: "packages/solo", Yes: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := adoptmove.Apply(dir, *plan, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
		t.Fatal("the move did not run")
	}
	_ = c
	if err := adoptmove.Rollback(dir, false); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"package.json", "index.js"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("%s did not come back: %v", f, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "packages")); err == nil {
		t.Error("the empty destination folder stayed")
	}
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	if out, _ := cmd.Output(); len(out) != 0 {
		t.Errorf("the tree is not clean:\n%s", out)
	}
}

func TestNeedsAnswerForMatchesPresent(t *testing.T) {
	plan := operationplan.Plan{Title: "t"}
	plan.Add("Files", operationplan.Line{Action: operationplan.Create, Path: "a"})
	if err := operationplan.NeedsAnswerFor(plan, operationplan.Options{Ask: true}); err == nil {
		t.Error("no terminal and no --yes must refuse")
	}
	if err := operationplan.NeedsAnswerFor(plan, operationplan.Options{Ask: true, Yes: true}); err != nil {
		t.Errorf("--yes must not refuse: %v", err)
	}
}
