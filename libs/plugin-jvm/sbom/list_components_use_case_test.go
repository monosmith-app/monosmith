package sbom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func put(t *testing.T, root, rel, body string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(abs), 0o755)
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListReadsGradleLockfileWithoutTestOnlyDependencies(t *testing.T) {
	root := t.TempDir()
	put(t, root, "gradle.lockfile", "# comment\ncom.google.guava:guava:33.0.0-jre=compileClasspath,runtimeClasspath\norg.junit:junit:5.0=testCompileClasspath,testRuntimeClasspath\nempty=annotationProcessor\n")
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "app"}})
	if err != nil || len(res.Components) != 1 || res.Components[0].PURL != "pkg:maven/com.google.guava/guava@33.0.0-jre" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestListReadsPOMDependencies(t *testing.T) {
	root := t.TempDir()
	put(t, root, "pom.xml", `<project><properties><jackson.version>2.17.0</jackson.version></properties>
<dependencyManagement><dependencies><dependency><groupId>x</groupId><artifactId>managed</artifactId><version>1</version></dependency></dependencies></dependencyManagement>
<dependencies>
<dependency><groupId>com.fasterxml.jackson.core</groupId><artifactId>jackson-core</artifactId><version>${jackson.version}</version></dependency>
<dependency><groupId>org.junit.jupiter</groupId><artifactId>junit-jupiter</artifactId><version>5.10.0</version><scope>test</scope></dependency>
</dependencies></project>`)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || res.Source != "pom.xml" || len(res.Components) != 1 || res.Components[0].PURL != "pkg:maven/com.fasterxml.jackson.core/jackson-core@2.17.0" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestListFallsBackToTheGradleScript(t *testing.T) {
	root := t.TempDir()
	put(t, root, "build.gradle.kts", "dependencies {\n    implementation(\"io.ktor:ktor-server-core:2.3.0\")\n    testImplementation(\"junit:junit:4.13\")\n}\n")
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || len(res.Components) != 1 || res.Components[0].Name != "io.ktor/ktor-server-core" {
		t.Fatalf("%+v %v", res, err)
	}
}
