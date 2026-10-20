package cipipeline

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAuditSettingsCarriesTheCISectionsAuditKeys(t *testing.T) {
	got := auditSettings(workspace.CISection{AuditPolicies: []string{"node", "go"}, AuditWarn: []string{"go"}, AuditAllowInstallScripts: []string{"esbuild"}})
	want := protocol.AuditSettings{Policies: []string{"node", "go"}, Warn: []string{"go"}, AllowInstallScripts: []string{"esbuild"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
	if got := auditSettings(workspace.CISection{}); got.Enabled("node") || got.Enabled("go") {
		t.Errorf("off by default: %+v", got)
	}
}
