package scaffold

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func runNew(t *testing.T, interactive bool, input string, args ...string) (string, string, error) {
	t.Helper()
	isolatedGit(t)
	restore := newIsInteractive
	newIsInteractive = func() bool { return interactive }
	t.Cleanup(func() { newIsInteractive = restore })
	dir := filepath.Join(t.TempDir(), "acme")
	cmd := NewCommand("1.0.0")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetArgs(append([]string{"acme", "--dir", dir}, args...))
	err := cmd.Execute()
	return dir, out.String(), err
}

func TestFlagsAnswerEveryQuestionWithoutAsking(t *testing.T) {
	dir, out, err := runNew(t, false, "", "--ci", "azure", "--language", "node", "--default-stack", "node=node-vitest",
		"--architecture", "none", "--release-branch", "trunk", "--cache-max-mb", "0")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := workspace.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CI.Providers[0] != "azure" || cfg.DefaultStacks["node"] != "node-vitest" || len(cfg.Conventions) != 0 ||
		cfg.Release.Branch != "trunk" || cfg.Cache.MaxSizeMB != -1 {
		t.Errorf("%+v", cfg)
	}
	if strings.Contains(out, "Choice") || !strings.Contains(out, "smith add app web --lang node") {
		t.Errorf("nothing is asked without a terminal, and the next step names the language:\n%s", out)
	}
}

func TestAskingShowsTheChoicesBeforeWriting(t *testing.T) {
	_, out, err := runNew(t, true, "\n\n\n\n\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"New workspace: acme", "Which CI should smith write pipelines for?", "Your choices", "Code layout", "vertical-slices"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
	if strings.Index(out, "Your choices") > strings.Index(out, "wrote smith.toml") {
		t.Errorf("the summary comes before the files:\n%s", out)
	}
}

func TestYesSkipsTheQuestionsAtATerminal(t *testing.T) {
	_, out, err := runNew(t, true, "", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "Choice [") {
		t.Errorf("asked despite --yes:\n%s", out)
	}
}

func TestABadAnswerFlagCreatesNothing(t *testing.T) {
	dir, _, err := runNew(t, false, "", "--ci", "jenkins")
	if err == nil || !strings.Contains(err.Error(), "--ci jenkins") {
		t.Fatalf("%v", err)
	}
	if _, loadErr := workspace.Load(dir); loadErr == nil {
		t.Error("a workspace was created")
	}
}
