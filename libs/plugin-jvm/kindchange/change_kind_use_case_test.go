package kindchange

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func edited(r protocol.ChangeKindResult) map[string]string {
	out := map[string]string{}
	for _, f := range r.Files {
		out[f.Path] = f.Content
	}
	return out
}

func apply(t *testing.T, dir string, r protocol.ChangeKindResult) {
	t.Helper()
	for _, e := range r.Files {
		full := filepath.Join(dir, filepath.FromSlash(e.Path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(e.Content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

const corePom = `<project>
  <parent>
    <groupId>dev.acme</groupId>
    <artifactId>acme-parent</artifactId>
    <version>0.0.0</version>
    <relativePath>../../pom.xml</relativePath>
  </parent>
  <artifactId>acme-core</artifactId>

  <properties>
    <smith.kind>lib</smith.kind>
  </properties>
</project>
`

var mavenWorkspace = map[string]string{
	"pom.xml":           "<project>\n  <modules>\n    <module>libs/core</module>\n    <module>apps/api</module>\n  </modules>\n</project>\n",
	"libs/core/pom.xml": corePom,
	"apps/api/pom.xml":  "<project>\n  <properties>\n    <smith.kind>app</smith.kind>\n    <exec.skip>false</exec.skip>\n    <exec.mainClass>dev.acme.api.Main</exec.mainClass>\n  </properties>\n</project>\n",
	"apps/api/src/main/java/dev/acme/api/Main.java": "package dev.acme.api;\n",
}

func TestMavenLibBecomesPackage(t *testing.T) {
	dir := writeFiles(t, mavenWorkspace)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "core", Root: "libs/core", Language: "jvm"},
		To:      protocol.ProjectPackage, NewRoot: "packages/core",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	if !strings.Contains(got["libs/core/pom.xml"], "<smith.kind>package</smith.kind>") {
		t.Fatalf("the kind changes:\n%s", got["libs/core/pom.xml"])
	}
	if !strings.Contains(got["pom.xml"], "<module>packages/core</module>") || !strings.Contains(got["pom.xml"], "<module>apps/api</module>") {
		t.Fatalf("the aggregator lists the new directory:\n%s", got["pom.xml"])
	}
	if _, touched := got["libs/core/pom.xml"]; !touched {
		t.Fatal("the pom is edited")
	}
	joined := strings.Join(r.Warnings, "\n")
	if !strings.Contains(joined, "a description") || !strings.Contains(joined, "a license") {
		t.Fatalf("the repository's requirements are named: %v", r.Warnings)
	}
}

func TestMavenParentPathFollowsAMoveToADifferentDepth(t *testing.T) {
	dir := writeFiles(t, mavenWorkspace)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "core", Root: "libs/core", Language: "jvm"},
		To:      protocol.ProjectPackage, NewRoot: "packages/group/core",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := edited(r)["libs/core/pom.xml"]; !strings.Contains(got, "<relativePath>../../../pom.xml</relativePath>") {
		t.Fatalf("a deeper folder reaches the parent pom one level further:\n%s", got)
	}
}

func TestMavenLibToPackageAndBackLeavesThePomByteForByte(t *testing.T) {
	dir := writeFiles(t, map[string]string{"libs/core/pom.xml": corePom})
	project := protocol.Project{Name: "core", Root: "libs/core", Language: "jvm"}
	for _, to := range []protocol.ProjectKind{protocol.ProjectPackage, protocol.ProjectLib} {
		r, err := Change(dir, protocol.ChangeKindParams{Project: project, To: to, NewRoot: "libs/core"})
		if err != nil {
			t.Fatal(err)
		}
		apply(t, dir, r)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "libs/core/pom.xml")); string(got) != corePom {
		t.Fatalf("round trip changed the pom:\n%s", got)
	}
}

func TestMavenAppToLibDropsTheExecPropertiesAndSaysSo(t *testing.T) {
	dir := writeFiles(t, mavenWorkspace)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "api", Root: "apps/api", Language: "jvm"},
		To:      protocol.ProjectLib, NewRoot: "libs/api",
	})
	if err != nil {
		t.Fatal(err)
	}
	api := edited(r)["apps/api/pom.xml"]
	if strings.Contains(api, "exec.") || !strings.Contains(api, "<smith.kind>lib</smith.kind>") {
		t.Fatalf("a lib is not run:\n%s", api)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "program") {
		t.Fatalf("the program is named: %v", r.Warnings)
	}
}

func TestMavenLibBecomesAppWithItsMainClass(t *testing.T) {
	files := map[string]string{
		"libs/tool/pom.xml": "<project>\n  <properties>\n    <smith.kind>lib</smith.kind>\n  </properties>\n</project>\n",
		"libs/tool/src/main/java/dev/acme/tool/Main.java": "package dev.acme.tool;\n",
	}
	dir := writeFiles(t, files)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "tool", Root: "libs/tool", Language: "jvm"},
		To:      protocol.ProjectApp, NewRoot: "apps/tool",
	})
	if err != nil {
		t.Fatal(err)
	}
	pom := edited(r)["libs/tool/pom.xml"]
	for _, want := range []string{"<smith.kind>app</smith.kind>", "<exec.skip>false</exec.skip>", "<exec.mainClass>dev.acme.tool.Main</exec.mainClass>"} {
		if !strings.Contains(pom, want) {
			t.Fatalf("lacks %s:\n%s", want, pom)
		}
	}
}

var gradleWorkspace = map[string]string{
	"settings.gradle.kts":         "rootProject.name = \"acme\"\ninclude(\":libs:core\")\ninclude(\":apps:api\")\n",
	"libs/core/build.gradle.kts":  "plugins {\n    kotlin(\"jvm\")\n}\n",
	"libs/core/gradle.properties": "smith.kind=lib\nsmith.stack=kotlin\n",
	"apps/api/build.gradle.kts":   "dependencies {\n    implementation(project(\":libs:core\"))\n}\n",
}

func TestGradleLibBecomesPackage(t *testing.T) {
	dir := writeFiles(t, gradleWorkspace)
	projects := []protocol.Project{
		{Name: "core", Root: "libs/core", Language: "jvm"},
		{Name: "api", Root: "apps/api", Language: "jvm"},
	}
	r, err := Change(dir, protocol.ChangeKindParams{Project: projects[0], To: protocol.ProjectPackage, NewRoot: "packages/core", Projects: projects})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	if got["libs/core/gradle.properties"] != "smith.kind=package\nsmith.stack=kotlin\n" {
		t.Fatalf("the kind changes and the rest stays:\n%q", got["libs/core/gradle.properties"])
	}
	if !strings.Contains(got["settings.gradle.kts"], `include(":packages:core")`) || !strings.Contains(got["settings.gradle.kts"], `include(":apps:api")`) {
		t.Fatalf("the settings include the new project path:\n%s", got["settings.gradle.kts"])
	}
	if !strings.Contains(got["apps/api/build.gradle.kts"], `project(":packages:core")`) {
		t.Fatalf("dependents follow:\n%s", got["apps/api/build.gradle.kts"])
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "maven-publish") {
		t.Fatalf("the publishing plugin is named: %v", r.Warnings)
	}
}

func TestGradleModuleWithoutPropertiesGetsThem(t *testing.T) {
	dir := writeFiles(t, map[string]string{"libs/core/build.gradle.kts": "plugins {}\n"})
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "core", Root: "libs/core", Language: "jvm"},
		To:      protocol.ProjectLib, NewRoot: "libs/core",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := edited(r)["libs/core/gradle.properties"]; got != "smith.kind=lib\n" {
		t.Fatalf("a new properties file: %q", got)
	}
}
