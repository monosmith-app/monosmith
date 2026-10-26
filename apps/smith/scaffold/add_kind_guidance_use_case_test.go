package scaffold

import (
	"strings"
	"testing"
)

func TestAddingAPackageNameAsAKindPointsToInstall(t *testing.T) {
	const hint = "smith install -w <project> <package>"
	// The kind is checked before anything of the workspace is read.
	if _, err := AddProject(nil, AddOptions{Kind: "zod", Name: "web"}); err == nil || !strings.Contains(err.Error(), hint) {
		t.Fatalf("add with a package as the kind: %v", err)
	}
	question := addQuestions()[0]
	if err := question.Validate("zod"); err == nil || !strings.Contains(err.Error(), hint) {
		t.Fatalf("the question for the kind: %v", err)
	}
	if err := question.Validate("lib"); err != nil {
		t.Fatalf("a real kind: %v", err)
	}
}
