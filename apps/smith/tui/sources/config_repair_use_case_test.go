package sources

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigRepairOfPlansTheRenameAndWritesNothing(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", "") // smith sets it for its own tasks (CI); it would point FindRoot at the real workspace
	dir := t.TempDir()
	file := filepath.Join(dir, "smith.toml")
	text := "[workspace]\nname = \"x\"\n\n[[task_groups]]\nname = \"a\"\n"
	if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	r, ok := ConfigRepairOf(dir, "test")
	if !ok {
		t.Fatal("a renamed key is repairable")
	}
	review, err := r.Show(r.Fix)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"-[[task_groups]]", "+[[groups]]"} {
		if !strings.Contains(review, want) {
			t.Errorf("the review misses %q:\n%s", want, review)
		}
	}
	if got, _ := os.ReadFile(file); string(got) != text {
		t.Errorf("planning changed the file: %q", got)
	}
}

func TestConfigRepairOfRefusesAFileDoctorCannotFix(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", "") // smith sets it for its own tasks (CI); it would point FindRoot at the real workspace
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "smith.toml"), []byte("[workspace\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := ConfigRepairOf(dir, "test"); ok {
		t.Fatal("a syntax error is not a renamed key")
	}
}
