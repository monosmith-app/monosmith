package pipeline

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCIDeclaresTheNpmDownloadStore(t *testing.T) {
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{{Name: "web", Root: "apps/web"}}}, "22").Stores
	want := []protocol.PackageStore{{Name: "npm", Path: "~/.npm", Env: "NPM_CONFIG_CACHE", Files: []string{"package-lock.json"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
	if none := CI(t.TempDir(), protocol.CIParams{}, "22"); len(none.Stores) != 0 {
		t.Errorf("no Node project, no store: %+v", none.Stores)
	}
}

func TestCIAddsTheNpmPolicyStepOnlyWhenOptedIn(t *testing.T) {
	projects := []protocol.Project{{Name: "web", Root: "apps/web"}}
	if got := CI(t.TempDir(), protocol.CIParams{Projects: projects}, "22").Audit; len(got) != 1 {
		t.Fatalf("off by default: %+v", got)
	}
	audit := protocol.AuditSettings{Policies: []string{"node"}, AllowInstallScripts: []string{"esbuild", "sharp"}}
	got := CI(t.TempDir(), protocol.CIParams{Projects: projects, Audit: audit}, "22").Audit
	if len(got) != 2 {
		t.Fatalf("%+v", got)
	}
	want := protocol.Step{Name: "npm policy (lockfile)", Command: []string{"{smith}", "npm-policy", "--allow-scripts=esbuild,sharp"}, Blocking: true, CacheFiles: []string{"package-lock.json"}}
	if !reflect.DeepEqual(got[1], want) {
		t.Errorf("%+v", got[1])
	}
	audit.Warn = []string{"node"}
	if CI(t.TempDir(), protocol.CIParams{Projects: projects, Audit: audit}, "22").Audit[1].Blocking {
		t.Error("warn makes it report only")
	}
}
