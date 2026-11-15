package dependencies

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestUndeclared(t *testing.T) {
	ds := []protocol.Declaration{{Project: "web", Name: "react"}, {Project: "web", Name: "zod"}, {Project: "api", Name: "nothere"}}
	missing, declared := undeclared(ds, "web", []string{"React", "nothere"})
	if !reflect.DeepEqual(missing, []string{"nothere"}) {
		t.Fatalf("missing = %v", missing)
	}
	if !reflect.DeepEqual(declared, []string{"react", "zod"}) {
		t.Fatalf("declared = %v", declared)
	}
}
