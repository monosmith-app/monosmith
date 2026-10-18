package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// module writes a module's manifest under ws and returns the project.
func module(t *testing.T, ws, root, manifest string, kind protocol.ProjectKind) protocol.Project {
	t.Helper()
	dir := filepath.Join(ws, root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifest), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return protocol.Project{Root: root, Kind: kind}
}

func TestCI(t *testing.T) {
	ws := t.TempDir()
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{module(t, ws, "core", "pom.xml", protocol.ProjectPackage)}})
	if len(res.Toolchains) != 1 || res.Toolchains[0].Version != JavaVersion || res.Setup[0].Command[3] != "dependency:go-offline" || len(res.Credentials) != 3 {
		t.Fatalf("%+v", res)
	}
	_ = os.WriteFile(filepath.Join(ws, "pom.xml"), []byte("<properties><maven.compiler.release>17</maven.compiler.release></properties>"), 0o644)
	if v := Version(ws); v != "17" {
		t.Fatalf("the root pom pins %s", v)
	}
	_ = os.WriteFile(filepath.Join(ws, ".java-version"), []byte("25.0.1\n"), 0o644)
	if v := Version(ws); v != "25" {
		t.Fatalf(".java-version pins %s", v)
	}
}

func TestCIWithGradle(t *testing.T) {
	ws := t.TempDir()
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{module(t, ws, "api", "build.gradle.kts", protocol.ProjectApp)}})
	if g := res.Toolchains[len(res.Toolchains)-1]; len(res.Toolchains) != 2 || g.Name != "gradle" || g.GitHub != nil || !strings.Contains(g.Install, "gradle-"+GradleVersion+"-bin.zip") || !strings.Contains(g.Install, "GITHUB_PATH") {
		t.Fatalf("Gradle is installed: %+v", res.Toolchains)
	}
	if res.Toolchains[0].GitHub.Inputs[2].Value != "gradle" || len(res.Setup) != 0 || len(res.Credentials) != 0 {
		t.Fatalf("no Maven steps for a Gradle-only workspace: %+v", res)
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(_ string, argv []string) (string, error) {
		if argv[0] == "mvn" {
			return "", errors.New("missing")
		}
		return "", nil
	}
	ws := t.TempDir()
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{module(t, ws, "core", "pom.xml", protocol.ProjectLib)}})
	if len(res.Findings) != 2 || !res.Findings[0].OK || res.Findings[1].OK {
		t.Fatalf("%+v", res)
	}
	res = Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{module(t, ws, "api", "build.gradle.kts", protocol.ProjectApp)}})
	if len(res.Findings) != 2 || res.Findings[1].Check != "Gradle is installed" || !res.Findings[1].OK {
		t.Fatalf("a Gradle module checks Gradle, not Maven: %+v", res)
	}
}

func TestAndroid(t *testing.T) {
	ws := t.TempDir()
	app := module(t, ws, "apps/mobile", "build.gradle.kts", protocol.ProjectApp)
	_ = os.WriteFile(filepath.Join(ws, "apps", "mobile", "build.gradle.kts"), []byte(`plugins { id("com.android.application") }`), 0o644)
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{app}})
	if sdk := res.Toolchains[len(res.Toolchains)-1]; sdk.Name != "android-sdk" || !strings.Contains(sdk.Install, "commandlinetools-linux-"+AndroidToolsRevision) {
		t.Fatalf("%+v", res.Toolchains)
	}
	t.Setenv("ANDROID_HOME", "")
	t.Setenv("ANDROID_SDK_ROOT", "")
	if androidSDKFound(ws) {
		t.Fatal("no SDK")
	}
	_ = os.WriteFile(filepath.Join(ws, "local.properties"), []byte("sdk.dir="+filepath.ToSlash(ws)+"\n"), 0o644)
	if !androidSDKFound(ws) {
		t.Fatal("local.properties names the SDK")
	}
}

func TestAPackageReleasedToNoRegistryNeedsNoMavenSecrets(t *testing.T) {
	ws := t.TempDir()
	java := module(t, ws, "java", "pom.xml", protocol.ProjectPackage)
	java.Registries = &protocol.Registries{NoPublish: true}
	if res := CI(ws, protocol.CIParams{Projects: []protocol.Project{java}}); len(res.Credentials) != 0 {
		t.Errorf("credentials %+v for a package that publishes nowhere", res.Credentials)
	}
}
