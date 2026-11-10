package discovery

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAnAppCrateShipsItsBinaryAndOutputsOnlyTheShippedFolder(t *testing.T) {
	build := Tasks("apps/cli", "acme-cli", protocol.ProjectApp, nil)["build"]
	if !slices.Equal(build.Outputs, []string{"../../dist/apps/cli"}) {
		t.Errorf("outputs %v, want only the shipped folder", build.Outputs)
	}
	want := []string{"../../dist/cargo/debug/acme-cli", "../../dist/cargo/debug/acme-cli.exe"}
	if !slices.Equal(build.Ship, want) {
		t.Errorf("ship %v, want %v", build.Ship, want)
	}
	for _, o := range build.Outputs {
		if strings.Contains(o, "dist/cargo") {
			t.Errorf("the shared target folder is an output: %s", o)
		}
	}
	if build.Cache {
		t.Error("the cargo build stays uncached (ADR 0092, owner decision 7)")
	}
}

func TestALibraryCrateShipsNothing(t *testing.T) {
	build := Tasks("libs/core", "core", protocol.ProjectLib, nil)["build"]
	if len(build.Ship) != 0 || len(build.Outputs) != 0 {
		t.Errorf("a library ships nothing: %+v", build)
	}
}

func TestADeclaredProgramShipsItsOwnBinary(t *testing.T) {
	build := binTasks("tools/src/bin/gen", "acme-tools", "gen", protocol.ProjectApp)["build"]
	if !slices.Equal(build.Outputs, []string{"../../../../dist/tools/src/bin/gen"}) || build.Ship[0] != "../../../../dist/cargo/debug/gen" {
		t.Errorf("build %+v", build)
	}
}
