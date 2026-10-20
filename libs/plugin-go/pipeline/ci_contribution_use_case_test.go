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

func TestCIAddsGoModVerifyPerModuleOnlyWhenOptedIn(t *testing.T) {
	projects := []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "core", Root: "libs/core"}}
	if got := CI(t.TempDir(), protocol.CIParams{Projects: projects}).Audit; len(got) != 2 {
		t.Fatalf("off by default: %+v", got)
	}
	got := CI(t.TempDir(), protocol.CIParams{Projects: projects, Audit: protocol.AuditSettings{Policies: []string{"go"}}}).Audit
	if len(got) != 4 {
		t.Fatalf("%+v", got)
	}
	want := protocol.Step{Name: "go mod verify api", Command: []string{"go", "mod", "verify"}, Dir: "apps/api", Blocking: true, CacheFiles: []string{"apps/api/go.mod", "apps/api/go.sum"}}
	if !reflect.DeepEqual(got[2], want) {
		t.Errorf("%+v", got[2])
	}
	warn := protocol.AuditSettings{Policies: []string{"go"}, Warn: []string{"go"}}
	if CI(t.TempDir(), protocol.CIParams{Projects: projects, Audit: warn}).Audit[2].Blocking {
		t.Error("warn makes it report only")
	}
}
