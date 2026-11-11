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

func TestTheReleaseProfileBuildsWithReleaseAndShipsFromTheReleaseFolder(t *testing.T) {
	tasks := Tasks("apps/cli", "acme-cli", protocol.ProjectApp, nil)
	build := tasks["build"]
	if build.Release == nil || !slices.Equal(build.Release.Command, []string{"cargo", "build", "-p", "acme-cli", "--release"}) {
		t.Fatalf("release profile %+v", build.Release)
	}
	want := []string{"../../dist/cargo/release/acme-cli", "../../dist/cargo/release/acme-cli.exe"}
	if !slices.Equal(build.Release.Ship, want) {
		t.Errorf("release ship %v, want %v", build.Release.Ship, want)
	}
	for _, name := range []string{"test", "start"} {
		if r := tasks[name].Release; r == nil || r.Command[len(r.Command)-1] != "--release" {
			t.Errorf("%s has no release profile: %+v", name, r)
		}
	}
	for _, name := range []string{"lint", "typecheck", "format"} {
		if tasks[name].Release != nil {
			t.Errorf("%s needs no release profile", name)
		}
	}
}

func TestACrateShipsEachOfItsPrograms(t *testing.T) {
	build := Tasks("apps/cli", "loomas-cli", protocol.ProjectApp, nil, "loomas", "loomas-admin")["build"]
	want := []string{
		"../../dist/cargo/debug/loomas", "../../dist/cargo/debug/loomas.exe",
		"../../dist/cargo/debug/loomas-admin", "../../dist/cargo/debug/loomas-admin.exe",
	}
	if !slices.Equal(build.Ship, want) {
		t.Errorf("ship %v, want %v", build.Ship, want)
	}
	if build.Release.Ship[0] != "../../dist/cargo/release/loomas" {
		t.Errorf("release ship %v", build.Release.Ship)
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
	if build.Release == nil || build.Release.Command[len(build.Release.Command)-1] != "--release" || build.Release.Ship[0] != "../../../../dist/cargo/release/gen" {
		t.Errorf("release %+v", build.Release)
	}
}
