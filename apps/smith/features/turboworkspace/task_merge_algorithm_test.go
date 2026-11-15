package turboworkspace

import (
	"reflect"
	"testing"
)

func TestMergeTask(t *testing.T) {
	no := false
	parent := TaskConfig{DependsOn: []string{"^build"}, Outputs: []string{"dist/**"}, Inputs: []string{"src/**"}}
	child := TaskConfig{
		Outputs: []string{ExtendsMarker, ".next/**"},
		Inputs:  []string{"app/**"},
		Cache:   &no,
	}
	got := MergeTask(parent, child)
	if !reflect.DeepEqual(got.DependsOn, []string{"^build"}) {
		t.Fatalf("an unset field keeps the parent's: %v", got.DependsOn)
	}
	if !reflect.DeepEqual(got.Outputs, []string{"dist/**", ".next/**"}) {
		t.Fatalf("$TURBO_EXTENDS$ splices the parent's entries: %v", got.Outputs)
	}
	if !reflect.DeepEqual(got.Inputs, []string{"app/**"}) {
		t.Fatalf("a set list replaces the parent's: %v", got.Inputs)
	}
	if got.Cache == nil || *got.Cache {
		t.Fatal("cache: false carries over")
	}
}
