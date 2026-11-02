package operation

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// failingCheck is a command that writes a project and then fails its check.
func failingCheck(t *testing.T) *cobra.Command {
	inner := &cobra.Command{Use: "add", Args: cobra.ExactArgs(1)}
	inner.Flags().Bool("dry-run", false, "")
	inner.RunE = func(_ *cobra.Command, args []string) error {
		put(t, ".", "libs/"+args[0]+"/index.txt", "x\n")
		return AfterWork(errors.New("a tool it needs is missing"))
	}
	return Wrap(inner, Spec{ChecksResult: true, AddYes: true, Describe: func(_ *cobra.Command, args []string) string { return "add lib " + args[0] }})
}

func runFailingCheck(t *testing.T, input string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := failingCheck(t)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestAFailedCheckStillRecordsTheOperationAndSaysWhatHappenedToTheCommit(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
		// committed is true when the files end up in a commit.
		committed bool
	}{
		{"no flag", []string{"core"}, "Not committed.", false},
		{"commit without a terminal refuses and names the flag", []string{"core", "--commit"}, "--commit-anyway", false},
		{"commit and yes still refuse", []string{"core", "--commit", "--yes"}, "--commit-anyway", false},
		{"commit anyway commits", []string{"core", "--commit", "--commit-anyway"}, "Committed", true},
	}
	for _, c := range cases {
		g := scratch(t, false)
		out, err := runFailingCheck(t, "", c.args...)
		if err == nil || !isVerifyFailure(err) {
			t.Fatalf("%s: the failure must stay: %v", c.name, err)
		}
		if !strings.Contains(out, c.want) {
			t.Errorf("%s: output lacks %q:\n%s", c.name, c.want, out)
		}
		subjects := commits(t, g)
		if got := subjects[0] == "chore: add lib core"; got != c.committed {
			t.Errorf("%s: committed=%v, commits %q", c.name, got, subjects)
		}
		if c.committed {
			continue
		}
		u, ok := ReadPending(g.Dir)
		if !ok || u.Command != "add lib core" || len(u.Files) == 0 {
			t.Errorf("%s: the record of the operation is lost: %+v %v", c.name, u, ok)
		}
	}
}

func TestAfterAFailedCheckSmithCommitUsesTheOperationName(t *testing.T) {
	g := scratch(t, false)
	if _, err := runFailingCheck(t, "", "core", "--commit"); err == nil {
		t.Fatal("the failure was swallowed")
	}
	if _, err := runCommit(t, "", "--only-operation", "--yes"); err != nil {
		t.Fatal(err)
	}
	if got := commits(t, g)[0]; got != "chore: add lib core" {
		t.Fatalf("commit subject %q", got)
	}
}
