package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("pom.xml", `<project><groupId>dev.acme</groupId><artifactId>acme-parent</artifactId><version>0.0.0</version><packaging>pom</packaging><modules><module>apps/api</module><module>libs/core</module></modules></project>`)
	write("apps/api/pom.xml", `<project>
  <parent><groupId>dev.acme</groupId><artifactId>acme-parent</artifactId><version>0.0.0</version></parent>
  <artifactId>acme-api</artifactId><version>1.2.0</version>
  <properties><smith.kind>app</smith.kind><exec.mainClass>dev.acme.api.Main</exec.mainClass></properties>
  <dependencies><dependency><groupId>${project.groupId}</groupId><artifactId>acme-core</artifactId></dependency><dependency><groupId>org.junit.jupiter</groupId><artifactId>junit-jupiter</artifactId></dependency></dependencies>
</project>`)
	write("libs/core/pom.xml", `<project><parent><groupId>dev.acme</groupId><version>0.0.0</version></parent><artifactId>acme-core</artifactId><properties><smith.stack>java</smith.stack></properties></project>`)
	res, err := Discover(ws, []string{"pom.xml", "apps/api/pom.xml", "libs/core/pom.xml"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 2 {
		t.Fatalf("the root pom is not a project: %+v", res.Projects)
	}
	api, core := res.Projects[0], res.Projects[1]
	if api.Kind != protocol.ProjectApp || api.Version != "1.2.0" || api.Exports[0] != "dev.acme:acme-api" || api.Imports[0] != "dev.acme:acme-core" || api.Tasks["start"].Command[5] != "exec:java" {
		t.Errorf("api %+v", api)
	}
	if build := api.Tasks["build"]; build.Cwd != "." || build.Command[4] != "dev.acme:acme-api" || build.Outputs[0] != "../../dist/apps/api" || build.Ship[0] != "../../dist/maven/acme-api/*.jar" || build.DependsOn[0] != "^build" {
		t.Errorf("api build %+v", build)
	}
	if api.Tasks["lint"].DependsOn[0] != "build" {
		t.Error("lint runs after the build, which empties the directory Checkstyle writes to")
	}
	if core.Kind != protocol.ProjectLib || core.Exports[0] != "dev.acme:acme-core" || core.Version != "0.0.0" || core.Stack != "java" {
		t.Errorf("core %+v", core)
	}
}

func TestDiscoverGradle(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("build.gradle.kts", `plugins { kotlin("jvm") version "2.4.21" apply false }`)
	write("buildSrc/build.gradle.kts", `plugins { `+"`kotlin-dsl`"+` }`)
	write("apps/api/build.gradle.kts", `plugins {
    kotlin("jvm")
    id("org.jlleitschuh.gradle.ktlint")
    application
}

group = "dev.acme"
version = "1.2.0"

layout.buildDirectory = rootProject.layout.projectDirectory.dir("dist/gradle/acme-api")

dependencies {
    implementation(project(":libs:core"))
}
`)
	write("apps/api/gradle.properties", "smith.kind=app\nsmith.stack=kotlin\n")
	write("libs/core/build.gradle", `plugins {
    id 'java-library'
    id 'checkstyle'
    id 'maven-publish'
}

group = 'dev.acme'

publishing {
    publications {
        maven(MavenPublication) {
            artifactId = 'acme-core'
        }
    }
}
`)
	write("libs/core/gradle.properties", "version=0.3.0\n")
	res, err := Discover(ws, []string{"build.gradle.kts", "buildSrc/build.gradle.kts", "apps/api/build.gradle.kts", "libs/core/build.gradle"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 2 {
		t.Fatalf("the root script and buildSrc are not projects: %+v", res.Projects)
	}
	api, core := res.Projects[0], res.Projects[1]
	if api.Kind != protocol.ProjectApp || api.Stack != "kotlin" || api.Version != "1.2.0" || api.Exports[0] != ":apps:api" || len(api.Exports) != 1 || api.Imports[0] != ":libs:core" {
		t.Errorf("api %+v", api)
	}
	if build := api.Tasks["build"]; build.Cwd != "." || build.Command[3] != ":apps:api:assemble" || build.Outputs[0] != "../../dist/apps/api" || build.Ship[0] != "../../dist/gradle/acme-api/libs/*.jar" || build.DependsOn[0] != "^build" {
		t.Errorf("api build %+v", build)
	}
	if lint := api.Tasks["lint"]; lint.Command[3] != ":apps:api:ktlintCheck" || lint.DependsOn[0] != "build" || api.Tasks["start"].Command[3] != ":apps:api:run" {
		t.Errorf("api lint %+v, start %+v", lint, api.Tasks["start"])
	}
	if core.Kind != protocol.ProjectLib || core.Version != "0.3.0" || core.Exports[1] != "dev.acme:acme-core" || len(core.Tasks["build"].Outputs) != 0 {
		t.Errorf("core %+v", core)
	}
	if lint := core.Tasks["lint"].Command; lint[3] != ":libs:core:checkstyleMain" || lint[4] != ":libs:core:checkstyleTest" {
		t.Errorf("core lint %v", lint)
	}
	if _, ok := core.Tasks["start"]; ok {
		t.Error("a lib has no start")
	}
}
