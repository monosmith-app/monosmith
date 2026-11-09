package sandbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func repo(t *testing.T) string {
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-q", "-m", "init")
	return dir
}

func TestTryChangesTheCopyAndNeverTheRealTree(t *testing.T) {
	root := repo(t)
	// An uncommitted edit must be in the copy.
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("edited"), 0o644); err != nil {
		t.Fatal(err)
	}
	var seen string
	run := func(dir string, _ []string, args ...string) (string, int, error) {
		if args[0] == "add" {
			b, _ := os.ReadFile(filepath.Join(dir, "a.txt"))
			seen = string(b)
			return "", 0, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("x"), 0o644)
		}
		return "healthy", 0, nil
	}
	res, err := Try(Options{Root: root, Command: []string{"add", "app", "web"}, Checks: []string{"doctor"}, Run: run})
	if err != nil {
		t.Fatal(err)
	}
	if seen != "edited" {
		t.Errorf("the copy lacks the uncommitted edit: %q", seen)
	}
	if !res.OK() || len(res.Changed) != 1 || !strings.Contains(res.Changed[0], "new.txt") {
		t.Errorf("result: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(root, "new.txt")); err == nil {
		t.Error("the real tree got the file")
	}
	if b, _ := os.ReadFile(filepath.Join(root, "a.txt")); string(b) != "edited" {
		t.Error("the real edit changed")
	}
}

func TestTryReportsAFailedCheckAndSkipsChecksAfterAFailedCommand(t *testing.T) {
	root := repo(t)
	failing := func(dir string, _ []string, args ...string) (string, int, error) {
		if args[0] == "doctor" {
			return "boom", 1, nil
		}
		return "", 0, nil
	}
	res, err := Try(Options{Root: root, Command: []string{"up"}, Checks: []string{"doctor"}, Run: failing})
	if err != nil || res.OK() || res.Checks[0].Tail != "boom" {
		t.Fatalf("%+v %v", res, err)
	}
	dead := func(string, []string, ...string) (string, int, error) { return "no", 2, nil }
	res, _ = Try(Options{Root: root, Command: []string{"up"}, Checks: []string{"doctor"}, Run: dead})
	if res.CommandOK || len(res.Checks) != 0 {
		t.Errorf("%+v", res)
	}
}

func TestTryRefusesWhatItCannotTry(t *testing.T) {
	root := repo(t)
	for _, tc := range []struct {
		cmd, checks []string
		want        string
	}{
		{[]string{"build"}, nil, "one of"},
		{[]string{"add"}, []string{"deploy"}, "not a check"},
	} {
		if _, err := Try(Options{Root: root, Command: tc.cmd, Checks: tc.checks}); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%v: %v", tc.cmd, err)
		}
	}
	if _, err := Try(Options{Root: t.TempDir(), Command: []string{"add"}}); err == nil || !strings.Contains(err.Error(), "git repository") {
		t.Errorf("no repository: %v", err)
	}
}

func TestParseStatusReadsRenamesAndSpaces(t *testing.T) {
	got := parseStatus(" M a b.txt\x00R  new.txt\x00old.txt\x00?? c.txt\x00")
	if len(got) != 3 || got[0].Path != "a b.txt" || got[1].OldPath != "old.txt" || got[2].XY != "??" {
		t.Errorf("%+v", got)
	}
}
