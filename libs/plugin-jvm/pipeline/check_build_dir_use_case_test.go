package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func jvmWorkspace(t *testing.T, files map[string]string) string {
	t.Helper()
	ws := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

const modulePom = "<project>\n  <modelVersion>4.0.0</modelVersion>\n  <artifactId>kit</artifactId>\n  <packaging>jar</packaging>\n</project>\n"

var kitModule = []protocol.Project{{Name: "kit", Root: "libs/kit", Tasks: map[string]protocol.Task{"build": {Outputs: []string{"../../dist/maven/kit"}}}}}

func firstMove(t *testing.T, f protocol.Finding) protocol.Fix {
	t.Helper()
	if f.Fix == nil || len(f.Fix.Choices) != 1 || len(f.Fix.Choices[0].Options) != 2 {
		t.Fatalf("a question, move or leave: %+v", f)
	}
	leave := f.Fix.Choices[0].Options[1].Fix
	if len(leave.Files) != 0 || len(leave.Steps) != 0 {
		t.Fatalf("leave changes nothing: %+v", leave)
	}
	return f.Fix.Choices[0].Options[0].Fix
}

func TestMavenModuleThatBuildsIntoTargetGetsTheBuildDirectory(t *testing.T) {
	ws := jvmWorkspace(t, map[string]string{"libs/kit/pom.xml": modulePom})
	findings, ok := CheckBuildDir(ws, kitModule)
	if !ok || len(findings) != 1 || findings[0].OK || findings[0].Warning {
		t.Fatalf("%+v", findings)
	}
	f := findings[0]
	if !strings.Contains(f.Detail, "builds into libs/kit/target/") || !strings.Contains(f.Detail, "not into dist/maven/kit/") || !strings.Contains(f.Detail, "cache hit restores nothing") {
		t.Fatalf("where it builds now, where it should, and the empty cache: %s", f.Detail)
	}
	got := firstMove(t, f).Files[0]
	want := "<project>\n  <modelVersion>4.0.0</modelVersion>\n  <artifactId>kit</artifactId>\n  <packaging>jar</packaging>\n  <build>\n    <directory>${project.basedir}/../../dist/maven/${project.artifactId}</directory>\n  </build>\n</project>\n"
	if got.Path != "libs/kit/pom.xml" || got.Content != want {
		t.Fatalf("%+v", got)
	}
}

func TestMavenModuleThatBuildsCentrallyPassesAndAnInheritedDirectoryCounts(t *testing.T) {
	ws := jvmWorkspace(t, map[string]string{
		"pom.xml":          "<project><artifactId>root</artifactId><packaging>pom</packaging><build><directory>${maven.multiModuleProjectDirectory}/dist/maven/${project.artifactId}</directory></build></project>",
		"libs/kit/pom.xml": "<project><parent><artifactId>root</artifactId><relativePath>../../pom.xml</relativePath></parent><artifactId>kit</artifactId></project>",
	})
	if f, _ := CheckBuildDir(ws, kitModule); len(f) != 1 || !f[0].OK {
		t.Fatalf("the parent's directory applies: %+v", f)
	}
}

func TestMavenModuleWithADirectoryOfItsOwnIsUpdatedWithoutTouchingResources(t *testing.T) {
	pom := "<project>\n  <artifactId>kit</artifactId>\n  <build>\n    <directory>out</directory>\n    <resources>\n      <resource>\n        <directory>src/res</directory>\n      </resource>\n    </resources>\n  </build>\n</project>\n"
	ws := jvmWorkspace(t, map[string]string{"libs/kit/pom.xml": pom})
	findings, _ := CheckBuildDir(ws, kitModule)
	if !strings.Contains(findings[0].Detail, "builds into libs/kit/out/") {
		t.Fatalf("%s", findings[0].Detail)
	}
	got := firstMove(t, findings[0]).Files[0].Content
	want := strings.Replace(pom, "<directory>out</directory>", "<directory>${project.basedir}/../../dist/maven/${project.artifactId}</directory>", 1)
	if got != want {
		t.Fatalf("%s", got)
	}
}

func TestMavenRefusesWithAWarningAndTheReason(t *testing.T) {
	for name, tc := range map[string]struct {
		files map[string]string
		want  string
	}{
		"a Dockerfile that copies the jar": {map[string]string{"libs/kit/pom.xml": modulePom, "libs/kit/Dockerfile": "COPY target/kit.jar /app.jar\n"}, "libs/kit/Dockerfile names libs/kit/target/"},
		"a path in the pom":                {map[string]string{"libs/kit/pom.xml": strings.Replace(modulePom, "<packaging>", "<description>see target/site</description><packaging>", 1)}, "its pom names target/ as a path"},
		"a pom that is not editable":       {map[string]string{"libs/kit/pom.xml": "<project><artifactId>kit</artifactId><build/></project>"}, "cannot edit safely"},
	} {
		t.Run(name, func(t *testing.T) {
			findings, _ := CheckBuildDir(jvmWorkspace(t, tc.files), kitModule)
			f := findings[0]
			if f.OK || !f.Warning || f.Fix != nil || !strings.Contains(f.Remedy, tc.want) {
				t.Fatalf("a warning with the reason and no fix: %+v", f)
			}
		})
	}
}

func TestMavenLeavesWhatItCannotReadUnchecked(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"a property": {"libs/kit/pom.xml": "<project><artifactId>kit</artifactId><build><directory>${some.dir}</directory></build></project>"},
		"a property in the parent": {
			"pom.xml":          "<project><artifactId>root</artifactId><packaging>pom</packaging><build><directory>${some.dir}</directory></build></project>",
			"libs/kit/pom.xml": "<project><parent><artifactId>root</artifactId><relativePath>../../pom.xml</relativePath></parent><artifactId>kit</artifactId></project>",
		},
	} {
		if f, ok := CheckBuildDir(jvmWorkspace(t, files), kitModule); ok {
			t.Errorf("%s: nothing is guessed: %+v", name, f)
		}
	}
}

var apiModule = []protocol.Project{{Name: "api", Root: "apps/api"}}

func TestGradleKotlinModuleGetsTheBuildDirectoryAfterThePluginsBlock(t *testing.T) {
	script := "plugins {\n    kotlin(\"jvm\")\n}\n\ngroup = \"app\"\n"
	ws := jvmWorkspace(t, map[string]string{"apps/api/build.gradle.kts": script, "settings.gradle.kts": "include(\":apps:api\")\n", "gradle/libs.versions.toml": "[versions]\n"})
	findings, ok := CheckBuildDir(ws, apiModule)
	if !ok || findings[0].OK || findings[0].Warning || !strings.Contains(findings[0].Detail, "builds into apps/api/build/") {
		t.Fatalf("%+v", findings)
	}
	got := firstMove(t, findings[0]).Files
	if len(got) != 1 || got[0].Path != "apps/api/build.gradle.kts" {
		t.Fatalf("only the module script is written, never settings or the catalog: %+v", got)
	}
	want := "plugins {\n    kotlin(\"jvm\")\n}\n\nlayout.buildDirectory.set(rootProject.layout.projectDirectory.dir(\"dist/gradle/api\"))\n\ngroup = \"app\"\n"
	if got[0].Content != want {
		t.Fatalf("%q", got[0].Content)
	}
}

func TestGradleGroovyModuleAndAnExistingSettingAreHandled(t *testing.T) {
	groovy := "plugins {\n    id 'java'\n}\nbuildDir = 'out'\nversion = '1'\n"
	ws := jvmWorkspace(t, map[string]string{"apps/api/build.gradle": groovy})
	findings, _ := CheckBuildDir(ws, apiModule)
	if !strings.Contains(findings[0].Detail, "builds into apps/api/out/") {
		t.Fatalf("%s", findings[0].Detail)
	}
	want := "plugins {\n    id 'java'\n}\nlayout.buildDirectory.set(rootProject.layout.projectDirectory.dir(\"dist/gradle/api\"))\nversion = '1'\n"
	if got := firstMove(t, findings[0]).Files[0].Content; got != want {
		t.Fatalf("%q", got)
	}
}

func TestGradleModuleWithoutAPluginsBlockGetsTheLineAtTheEnd(t *testing.T) {
	ws := jvmWorkspace(t, map[string]string{"apps/api/build.gradle": "apply plugin: 'java'\n"})
	findings, _ := CheckBuildDir(ws, apiModule)
	want := "apply plugin: 'java'\n\nlayout.buildDirectory.set(rootProject.layout.projectDirectory.dir(\"dist/gradle/api\"))\n"
	if got := firstMove(t, findings[0]).Files[0].Content; got != want {
		t.Fatalf("%q", got)
	}
}

func TestGradlePassesWhenCentralAndSkipsWhatTheRootScriptSets(t *testing.T) {
	central := "layout.buildDirectory = rootProject.layout.projectDirectory.dir(\"dist/gradle/api\")\n"
	if f, _ := CheckBuildDir(jvmWorkspace(t, map[string]string{"apps/api/build.gradle.kts": central}), apiModule); len(f) != 1 || !f[0].OK {
		t.Fatalf("%+v", f)
	}
	ws := jvmWorkspace(t, map[string]string{"apps/api/build.gradle.kts": "plugins {}\n", "build.gradle.kts": "allprojects { layout.buildDirectory = file(\"x/${name}\") }\n"})
	if f, ok := CheckBuildDir(ws, apiModule); ok {
		t.Fatalf("the root script decides: not checked: %+v", f)
	}
}

func TestGradleRefusesWithAWarningAndTheReason(t *testing.T) {
	for name, tc := range map[string]struct {
		files map[string]string
		want  string
	}{
		"a Dockerfile":     {map[string]string{"apps/api/build.gradle.kts": "plugins {}\n", "apps/api/Dockerfile": "COPY build/libs/api.jar /app.jar\n"}, "apps/api/Dockerfile names apps/api/build/"},
		"a literal build/": {map[string]string{"apps/api/build.gradle.kts": "plugins {}\nval f = file(\"build/generated\")\n"}, "names the build/ folder as a path"},
	} {
		t.Run(name, func(t *testing.T) {
			findings, _ := CheckBuildDir(jvmWorkspace(t, tc.files), apiModule)
			f := findings[0]
			if f.OK || !f.Warning || f.Fix != nil || !strings.Contains(f.Remedy, tc.want) {
				t.Fatalf("a warning with the reason and no fix: %+v", f)
			}
		})
	}
}

func TestMavenModuleThatBuildsElsewhereInDistGetsItsWorkFolder(t *testing.T) {
	ws := jvmWorkspace(t, map[string]string{
		"libs/kit/pom.xml": "<project>\n  <artifactId>kit</artifactId>\n  <build>\n    <directory>${project.basedir}/../../dist/libs/kit</directory>\n  </build>\n</project>\n",
	})
	f, _ := CheckBuildDir(ws, kitModule)
	if len(f) != 1 || f[0].OK || f[0].Warning || !strings.Contains(f[0].Detail, "not into dist/maven/kit/") || !strings.Contains(f[0].Detail, "shipped folder stays empty") {
		t.Fatalf("%+v", f)
	}
	firstMove(t, f[0])
}

func TestJvmBuildTaskThatShipsMustCacheTheShippedFolder(t *testing.T) {
	ws := jvmWorkspace(t, map[string]string{
		"apps/api/build.gradle.kts": "layout.buildDirectory = rootProject.layout.projectDirectory.dir(\"dist/gradle/api\")\n",
	})
	api := []protocol.Project{{Name: "api", Root: "apps/api", Tasks: map[string]protocol.Task{
		"build": {Outputs: []string{"../../dist/gradle/api"}, Ship: []string{"../../dist/gradle/api/libs/*.jar"}, Cache: true},
	}}}
	f, _ := CheckBuildDir(ws, api)
	if len(f) != 2 || !f[0].OK || f[1].OK || f[1].Warning || f[1].Problem != "does not cache dist/apps/api" {
		t.Fatalf("%+v", f)
	}
	api[0].Tasks["build"] = protocol.Task{Outputs: []string{"../../dist/apps/api"}, Ship: []string{"../../dist/gradle/api/libs/*.jar"}, Cache: true}
	if f, _ := CheckBuildDir(ws, api); len(f) != 1 || !f[0].OK {
		t.Fatalf("the module's own work folder is not shared: %+v", f)
	}
}
