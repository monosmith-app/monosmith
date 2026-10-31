package operation

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/spf13/cobra"
)

// scratch makes a git work tree with a smith.toml and one commit, and makes it
// the current directory. commit says what commit_operations is.
func scratch(t *testing.T, commit bool) *gitclient.Git {
	t.Helper()
	dir := t.TempDir()
	g := gitclient.New(dir)
	if err := g.Init(); err != nil {
		t.Skip("git is not available: ", err)
	}
	for _, kv := range [][2]string{{"user.email", "t@example.com"}, {"user.name", "T"}, {"commit.gpgsign", "false"}} {
		_, _ = g.Run("config", kv[0], kv[1])
	}
	toml := "[workspace]\nname = \"scratch\"\n"
	if commit {
		toml += "commit_operations = true\n"
	}
	put(t, dir, "smith.toml", toml)
	put(t, dir, ".gitignore", ".smith/\n")
	put(t, dir, "README.md", "hello\n")
	if err := g.CommitAll("chore: start"); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("SMITH_WORKSPACE_ROOT", dir)
	return g
}

func put(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fake is a mutating command: it writes libs/<arg>/index.txt and edits README.md.
func fake() *cobra.Command {
	cmd := &cobra.Command{Use: "add", Args: cobra.ExactArgs(1), RunE: func(c *cobra.Command, args []string) error {
		if err := os.MkdirAll(filepath.Join("libs", args[0]), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join("libs", args[0], "index.txt"), []byte("x\n"), 0o644); err != nil {
			return err
		}
		f, err := os.OpenFile("README.md", os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = f.WriteString(args[0] + "\n")
		return err
	}}
	cmd.Flags().Bool("dry-run", false, "")
	return Wrap(cmd, Spec{Describe: func(_ *cobra.Command, args []string) string { return "add lib " + args[0] }})
}

func execute(t *testing.T, args ...string) (string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := fake()
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	return out.String(), errOut.String()
}

func commits(t *testing.T, g *gitclient.Git) []string {
	t.Helper()
	out, err := g.Run("log", "--format=%s")
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(out, "\n")
}

func TestOneOperationMakesOneConventionalCommit(t *testing.T) {
	g := scratch(t, true)
	out, _ := execute(t, "billing")
	got := commits(t, g)
	if len(got) != 2 || got[0] != "chore: add lib billing" {
		t.Fatalf("commits: %q", got)
	}
	if !strings.Contains(out, "Committed") {
		t.Fatalf("no commit line:\n%s", out)
	}
	log, err := ReadLog(g.Dir)
	if err != nil || len(log.Operations) != 1 || log.Operations[0].Command != "add lib billing" {
		t.Fatalf("log: %+v %v", log, err)
	}
	if want := []string{"README.md", "libs/billing/index.txt"}; strings.Join(log.Operations[0].Files, ",") != strings.Join(want, ",") {
		t.Fatalf("files: %v", log.Operations[0].Files)
	}
}

func TestNoCommitWhenTheSettingIsOff(t *testing.T) {
	g := scratch(t, false)
	execute(t, "billing")
	if got := commits(t, g); len(got) != 1 {
		t.Fatalf("commits: %q", got)
	}
}

func TestNoCommitOnADryRun(t *testing.T) {
	g := scratch(t, true)
	execute(t, "billing", "--dry-run")
	if got := commits(t, g); len(got) != 1 {
		t.Fatalf("commits: %q", got)
	}
}

func TestNoCommitWhenTheTreeWasDirty(t *testing.T) {
	g := scratch(t, true)
	put(t, g.Dir, "notes.txt", "mine\n")
	_, errOut := execute(t, "billing")
	if got := commits(t, g); len(got) != 1 {
		t.Fatalf("commits: %q", got)
	}
	if !strings.Contains(errOut, "uncommitted changes") {
		t.Fatalf("no reason given:\n%s", errOut)
	}
	if status, _ := g.Run("status", "--porcelain"); !strings.Contains(status, "libs/") {
		t.Fatalf("the operation's files stay in the work tree:\n%s", status)
	}
}
