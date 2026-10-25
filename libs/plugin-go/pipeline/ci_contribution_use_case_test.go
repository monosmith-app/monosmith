package pipeline

import (
	"os"
	"path/filepath"
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

func TestWithoutGoWorkTheToolchainIsTheNewestGoModVersionNotTheDefault(t *testing.T) {
	root := t.TempDir()
	for dir, body := range map[string]string{"apps/api": "module a\n\ngo 1.9\n", "libs/core": "module c\n\ngo 1.27.1\n"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(dir), "go.mod"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	projects := []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "core", Root: "libs/core"}}
	got := CI(root, protocol.CIParams{Projects: projects}).Toolchains[0].Version
	if got != "1.27" {
		t.Errorf("version %q, want 1.27", got)
	}
	if v := CI(t.TempDir(), protocol.CIParams{Projects: projects}).Toolchains[0].Version; v != DefaultGoVersion {
		t.Errorf("no go.work and no go.mod: %q, want the default", v)
	}
}
