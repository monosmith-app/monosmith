package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestThePipelinesInstallTheVersionTheWorkspaceNamesOrElseTheRunningOne(t *testing.T) {
	var cfg workspace.Config
	if got := SmithVersionFor(cfg, "1.70.0"); got != "1.70.0" {
		t.Fatalf("without a setting: %q", got)
	}
	cfg.CI.SmithVersion = "latest"
	cfg.CI.Providers = []string{"github"}
	if got := SmithVersionFor(cfg, "1.70.0"); got != "latest" {
		t.Fatalf("with latest: %q", got)
	}
	m := NewModel(Contributions{}, SmithVersionFor(cfg, "1.71.0"), "main")
	found := false
	for path, content := range Rendered(cfg, m) {
		if strings.Contains(content, "SMITH_VERSION: latest") {
			found = true
		}
		if strings.Contains(content, "1.71.0") {
			t.Fatalf("%s pins the running version although the workspace asked for latest", path)
		}
	}
	if !found {
		t.Fatal("no pipeline installs latest")
	}
}
