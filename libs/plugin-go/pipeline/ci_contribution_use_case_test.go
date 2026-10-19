package pipeline

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCIDeclaresTheGoModuleCacheKeyedOnEachModulesGoSum(t *testing.T) {
	projects := []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "core", Root: "libs/core"}}
	got := CI(t.TempDir(), protocol.CIParams{Projects: projects}).Stores
	want := []protocol.PackageStore{{Name: "go", Path: "~/go/pkg/mod", Env: "GOMODCACHE", Files: []string{"apps/api/go.sum", "libs/core/go.sum"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
}
