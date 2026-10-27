package taskcache

import (
	"reflect"
	"testing"
)

func TestDefaultsNamedAsInputsAreKeyedEvenThoughGitIgnoresThem(t *testing.T) {
	globs := []string{"**/*", "{workspaceRoot}/ruff.toml", "{workspaceRoot}/.smith/defaults/ruff.toml", "{workspaceRoot}/.smith/defaults/*.toml", ".smith/defaults/local.toml"}
	got := DefaultInputs([]string{"a.py"}, globs)
	if want := []string{".smith/defaults/ruff.toml"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
	if got := DefaultInputs([]string{".smith/defaults/ruff.toml"}, globs); len(got) != 0 {
		t.Fatalf("already selected: %v", got)
	}
}
