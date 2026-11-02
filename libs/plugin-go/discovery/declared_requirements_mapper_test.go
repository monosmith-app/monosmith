package discovery

import (
	"reflect"
	"testing"
)

func TestDeclaredRequirementsAddsDirectModulesOnly(t *testing.T) {
	gomod := "module example.com/r/apps/api\n\ngo 1.26\n\nrequire example.com/r/libs/core v0.0.0\n\nrequire (\n\tgithub.com/x/y v1.2.3\n\tgithub.com/z/w v1.0.0 // indirect\n)\n"
	got := declaredRequirements([]string{"fmt", "github.com/x/y"}, gomod)
	want := []string{"example.com/r/libs/core", "fmt", "github.com/x/y"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
