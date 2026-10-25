package scaffold

import (
	"strings"
	"testing"
)

func TestNextStepsSkipTheCdWhenTheWorkspaceIsTheCurrentDirectory(t *testing.T) {
	here := nextSteps("acme", ".")
	if strings.Contains(here, "cd ") {
		t.Errorf("no cd for the current directory:\n%s", here)
	}
	if !strings.Contains(here, "smith add app web") {
		t.Errorf("the next step is still printed:\n%s", here)
	}
	if there := nextSteps("acme", "acme"); !strings.Contains(there, "  cd acme\n") {
		t.Errorf("a new directory needs the cd:\n%s", there)
	}
}

func TestAddSaysItLeavesItsFilesUncommitted(t *testing.T) {
	if long := AddCommand("dev").Long; !strings.Contains(long, "uncommitted") {
		t.Errorf("smith add --help must say it leaves its files uncommitted:\n%s", long)
	}
}
