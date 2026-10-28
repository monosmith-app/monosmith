package protocol

import (
	"slices"
	"testing"
)

func TestAPluginThatDeclaresNoKindsScaffoldsAll(t *testing.T) {
	var r InitializeResult
	if got := r.ScaffoldKinds(); !slices.Equal(got, []ProjectKind{ProjectApp, ProjectLib, ProjectPackage}) {
		t.Errorf("kinds = %v", got)
	}
	if !r.Scaffolds(ProjectPackage) {
		t.Error("a plugin that declares nothing scaffolds a package")
	}
}

func TestAPluginScaffoldsOnlyTheKindsItDeclares(t *testing.T) {
	r := InitializeResult{Kinds: []ProjectKind{ProjectLib, ProjectApp}}
	if got := r.ScaffoldKinds(); !slices.Equal(got, []ProjectKind{ProjectApp, ProjectLib}) {
		t.Errorf("kinds = %v, want app and lib in that order", got)
	}
	if r.Scaffolds(ProjectPackage) {
		t.Error("a package was not declared")
	}
}
