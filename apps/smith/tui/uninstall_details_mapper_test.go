package tui

import (
	"strings"
	"testing"
)

func TestTheUninstallPlanNamesThePackageAndTheRoot(t *testing.T) {
	got := strings.Join(uninstallDetails([]string{"uninstall", "-p", ".", "-p", "web", "@swc/core", "--yes"}), "\n")
	for _, want := range []string{"Removes @swc/core", "the workspace root", "  web"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if uninstallDetails([]string{"uninstall", "zod"}) != nil {
		t.Fatal("without a target there is no plan to show")
	}
}
