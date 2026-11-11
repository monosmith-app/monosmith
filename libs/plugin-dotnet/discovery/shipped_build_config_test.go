package discovery

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/functions"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// noSharedWorkFolder fails when an output is in the shared artifacts folder.
func noSharedWorkFolder(t *testing.T, task protocol.Task) {
	t.Helper()
	for _, o := range task.Outputs {
		if strings.Contains(o, "dist/dotnet") {
			t.Errorf("the shared artifacts folder is an output: %s", o)
		}
	}
}

func TestAnAppBuildShipsItsArtifactsFolderAndOutputsOnlyTheShippedFolder(t *testing.T) {
	build := Tasks("apps/api", "Acme.Api.csproj", "tests/Acme.Api.Tests.csproj", protocol.ProjectApp)["build"]
	if !slices.Equal(build.Outputs, []string{"../../dist/apps/api"}) || !slices.Equal(build.Ship, []string{"../../dist/dotnet/bin/Acme.Api/debug/*"}) {
		t.Errorf("build %+v", build)
	}
	noSharedWorkFolder(t, build)
	if build.Cache {
		t.Error("the .NET build stays uncached (ADR 0092, owner decision 7)")
	}
}

func TestTheReleaseProfileBuildsReleaseAndShipsTheReleaseFolder(t *testing.T) {
	tasks := Tasks("apps/api", "Acme.Api.csproj", "tests/Acme.Api.Tests.csproj", protocol.ProjectApp)
	build := tasks["build"]
	if build.Release == nil || !slices.Equal(build.Release.Command, []string{"dotnet", "build", "tests/Acme.Api.Tests.csproj", "--nologo", "-c", "Release"}) {
		t.Fatalf("release profile %+v", build.Release)
	}
	if !slices.Equal(build.Release.Ship, []string{"../../dist/dotnet/bin/Acme.Api/release/*"}) {
		t.Errorf("release ship %v", build.Release.Ship)
	}
	// test and start read the build without building: they must read the
	// configuration the build wrote.
	for _, name := range []string{"test", "start"} {
		r := tasks[name].Release
		if r == nil || !slices.Equal(r.Command[len(r.Command)-2:], []string{"-c", "Release"}) {
			t.Errorf("%s has no release profile: %+v", name, r)
		}
	}
}

func TestALibraryBuildShipsNothing(t *testing.T) {
	build := Tasks("libs/kit", "Kit.csproj", "", protocol.ProjectLib)["build"]
	if len(build.Ship) != 0 || len(build.Outputs) != 0 {
		t.Errorf("build %+v", build)
	}
}

func TestAFunctionAppShipsWhatItPublishes(t *testing.T) {
	tasks := Tasks("apps/fn", "Acme.Fn.csproj", "", protocol.ProjectApp)
	functions.WithFunctionTasks(tasks, functionhost.Azure, "apps/fn", "Acme.Fn.csproj", t.TempDir(), inputs)
	pkg, build := tasks["package"], tasks["build"]
	if !slices.Equal(pkg.Outputs, []string{"../../dist/apps/fn"}) || !slices.Equal(pkg.Ship, []string{"../../dist/dotnet/publish/Acme.Fn/release/*"}) {
		t.Errorf("package %+v", pkg)
	}
	if len(build.Ship)+len(build.Outputs) != 0 || (build.Release != nil && len(build.Release.Ship) != 0) {
		t.Errorf("the build of a function app ships nothing, the package does: %+v", build)
	}
	noSharedWorkFolder(t, pkg)
	if tasks["start"].Cwd != "dist/apps/fn" {
		t.Errorf("start runs the shipped folder: %+v", tasks["start"])
	}
}
