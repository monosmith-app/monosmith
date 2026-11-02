package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWithoutATerminalOrYesRefusesAndNamesEveryOpenChoice(t *testing.T) {
	dir, _, err := runNew(t, false, "")
	if err == nil {
		t.Fatal("smith new applied without a terminal and without --yes")
	}
	for _, want := range []string{"--ci", "--language", "--architecture", "--agent", "--devcontainer", "--first-commit", "--yes", "--dry-run"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal lacks %s:\n%v", want, err)
		}
	}
	if _, statErr := os.Stat(dir); statErr == nil {
		t.Error("a folder was created before the refusal")
	}
}

func TestNewDryRunListsTheFilesAndTheFirstCommitAndWritesNothing(t *testing.T) {
	dir, out, err := runNew(t, false, "", "--dry-run")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, want := range []string{"smith.toml", "First commit: chore: create the acme workspace with smith", "Dry run: nothing was changed."} {
		if !strings.Contains(out, want) {
			t.Errorf("the plan lacks %q:\n%s", want, out)
		}
	}
	if _, statErr := os.Stat(dir); statErr == nil {
		t.Error("--dry-run created the folder")
	}
}

func TestNewSaysWhatTheFirstCommitHoldsAndLetsTheFlagChooseNone(t *testing.T) {
	_, out, err := runNew(t, false, "", "--yes")
	if err != nil || !strings.Contains(out, "First commit: chore: create the acme workspace with smith") {
		t.Fatalf("%v\n%s", err, out)
	}
	dir, out, err := runNew(t, false, "", "--yes", "--first-commit", "none")
	if err != nil || !strings.Contains(out, "First commit: none") {
		t.Fatalf("%v\n%s", err, out)
	}
	if _, statErr := os.Stat(filepath.Join(dir, ".git")); statErr != nil {
		t.Errorf("git init still runs: %v", statErr)
	}
}

func TestInATerminalTheFirstCommitOfAFreshFolderIsAQuestion(t *testing.T) {
	_, out, err := runNew(t, true, "\n\n\n\n\n\n\n\n")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(out, "Should smith commit the files it writes?") {
		t.Errorf("the first commit was not asked:\n%s", out)
	}
}
