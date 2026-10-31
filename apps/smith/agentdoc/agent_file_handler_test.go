package agentdoc

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestFinishCheckSaysMissingForAFileThatDoesNotExist(t *testing.T) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	changes := []FileChange{
		{Path: "AGENTS.md", Differs: true},
		{Path: "CLAUDE.md", Differs: true},
	}
	err := finish(cmd, t.TempDir(), changes, true)
	if err == nil || !strings.Contains(err.Error(), "AGENTS.md and CLAUDE.md are missing") {
		t.Fatalf("want a missing failure, got %v", err)
	}
	for _, p := range []string{"AGENTS.md", "CLAUDE.md"} {
		if !strings.Contains(out.String(), p+" is missing, run smith agent file") {
			t.Errorf("%s not reported missing in %q", p, out.String())
		}
	}
	if strings.Contains(out.String(), "out of date") {
		t.Errorf("a missing file must not be called out of date: %q", out.String())
	}
}

func TestFinishCheckSeparatesMissingFromOutOfDate(t *testing.T) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	changes := []FileChange{
		{Path: "AGENTS.md", Exists: true, Differs: true},
		{Path: "CLAUDE.md", Differs: true},
	}
	err := finish(cmd, t.TempDir(), changes, true)
	if err == nil || !strings.Contains(err.Error(), "CLAUDE.md is missing; AGENTS.md is not up to date") {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(out.String(), "AGENTS.md is out of date") {
		t.Errorf("got %q", out.String())
	}
}

func TestFinishCheckPassesWhenEveryFileIsCurrent(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	if err := finish(cmd, t.TempDir(), []FileChange{{Path: "AGENTS.md", Exists: true}}, true); err != nil {
		t.Fatal(err)
	}
}
