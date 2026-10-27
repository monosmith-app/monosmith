package tooldefaults

import (
	"reflect"
	"testing"
)

func TestDefaultsNamedInArgsBecomeInputs(t *testing.T) {
	base := []string{"**/*"}
	got := WithDefaultInputs(base, []string{"--config", "{workspaceRoot}/.smith/defaults/ruff.toml", "--other", "x"})
	want := []string{"**/*", "{workspaceRoot}/.smith/defaults/ruff.toml"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
	if len(base) != 1 {
		t.Fatal("the base slice is not modified")
	}
	if got := WithDefaultInputs(base, nil); !reflect.DeepEqual(got, base) {
		t.Fatalf("no defaults, no change: %v", got)
	}
}
