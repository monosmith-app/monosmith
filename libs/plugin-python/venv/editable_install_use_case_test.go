package venv

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDevRequirementsReadsTheExtrasAndTheGroups(t *testing.T) {
	text := `
[project]
name = "itsdangerous"
[project.optional-dependencies]
dev = ["pytest"]
docs = ["sphinx"]
test = ["coverage"]

[dependency-groups]
docs = ["sphinx"]
tests = ["freezegun", "pytest", {include-group = "typing"}]
typing = ["mypy", "pytest"]
dev = [{include-group = "tests"}, "ruff"]
`
	extras, groups := DevRequirements([]byte(text))
	if !reflect.DeepEqual(extras, []string{"dev", "test"}) {
		t.Errorf("extras = %v, want dev and test (docs is not a development list)", extras)
	}
	if want := []string{"freezegun", "pytest", "mypy", "ruff"}; !reflect.DeepEqual(groups, want) {
		t.Errorf("groups = %v, want %v", groups, want)
	}
}

func TestDevRequirementsIgnoresTextThatIsNotTOML(t *testing.T) {
	if extras, groups := DevRequirements([]byte("[project")); extras != nil || groups != nil {
		t.Errorf("extras %v, groups %v", extras, groups)
	}
}

func TestEditableTargetAddsTheExtrasAndTheGroupRequirements(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "libs", "lib")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project.optional-dependencies]\ndev = [\"pytest\"]\n[dependency-groups]\ntests = [\"freezegun>=1\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, want := EditableTarget(root, "libs/lib"), []string{"-e", "./libs/lib[dev]", "freezegun>=1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("EditableTarget = %v, want %v", got, want)
	}
	if got := EditableTarget(root, "libs/missing"); !reflect.DeepEqual(got, []string{"-e", "./libs/missing"}) {
		t.Errorf("a project without pyproject.toml = %v", got)
	}
}
