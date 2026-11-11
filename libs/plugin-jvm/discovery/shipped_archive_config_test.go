package discovery

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// noWorkFolder fails when an output is a tool's work folder.
func noWorkFolder(t *testing.T, task protocol.Task) {
	t.Helper()
	for _, o := range task.Outputs {
		if strings.Contains(o, "dist/gradle") || strings.Contains(o, "dist/maven") {
			t.Errorf("a work folder is an output: %s", o)
		}
	}
}

func TestAMavenBuildShipsItsJarAndOutputsOnlyTheShippedFolder(t *testing.T) {
	build := Tasks("apps/api", "dev.acme:api", "api", protocol.ProjectApp)["build"]
	if !slices.Equal(build.Outputs, []string{"../../dist/apps/api"}) || !slices.Equal(build.Ship, []string{"../../dist/maven/api/*.jar"}) {
		t.Errorf("build %+v", build)
	}
	noWorkFolder(t, build)
}

func TestAGradleBuildShipsItsJarOrItsDebugAPK(t *testing.T) {
	jvm := GradleTasks("apps/api", GradleBuild{Path: ":apps:api", BuildDir: "dist/gradle/api"}, protocol.ProjectApp)["build"]
	if !slices.Equal(jvm.Outputs, []string{"../../dist/apps/api"}) || !slices.Equal(jvm.Ship, []string{"../../dist/gradle/api/libs/*.jar"}) {
		t.Errorf("jvm build %+v", jvm)
	}
	noWorkFolder(t, jvm)
	android := GradleTasks("apps/phone", GradleBuild{Path: ":apps:phone", BuildDir: "dist/gradle/phone", Android: true}, protocol.ProjectApp)["build"]
	if !slices.Equal(android.Ship, []string{"../../dist/gradle/phone/outputs/apk/debug/*.apk"}) {
		t.Errorf("android build %+v", android)
	}
	noWorkFolder(t, android)
}
