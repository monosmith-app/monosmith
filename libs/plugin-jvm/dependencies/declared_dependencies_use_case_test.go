package dependencies

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const pom = `<project>
  <properties>
    <junit.version>5.10.0</junit.version>
    <java.version>21</java.version>
  </properties>
  <dependencyManagement>
    <dependencies>
      <dependency>
        <groupId>com.google.guava</groupId>
        <artifactId>guava</artifactId>
        <version>32.0.0-jre</version>
      </dependency>
    </dependencies>
  </dependencyManagement>
  <dependencies>
    <dependency>
      <groupId>org.junit.jupiter</groupId>
      <artifactId>junit-jupiter</artifactId>
      <version>${junit.version}</version>
      <scope>test</scope>
    </dependency>
    <dependency>
      <groupId>com.google.guava</groupId>
      <artifactId>guava</artifactId>
    </dependency>
    <dependency>
      <groupId>org.slf4j</groupId>
      <artifactId>slf4j-api</artifactId>
      <version>[1.7,2.0)</version>
    </dependency>
    <dependency>
      <groupId>org.apache.commons</groupId>
      <artifactId>commons-lang3</artifactId>
      <version>3.12.0</version>
    </dependency>
    <dependency>
      <groupId>com.acme</groupId>
      <artifactId>inherited</artifactId>
      <version>${acme.version}</version>
    </dependency>
  </dependencies>
</project>
`

const gradle = `plugins { kotlin("jvm") }
dependencies {
    implementation("org.apache.commons:commons-lang3:3.12.0")
    testImplementation 'org.junit.jupiter:junit-jupiter:5.10.0'
    implementation(libs.guava)
    implementation("com.acme:dynamic:$acmeVersion")
}
`

var projects = []protocol.Project{{Name: "svc", Root: "svc"}, {Name: "tool", Root: "tool"}}

func writeWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	for file, body := range map[string]string{"svc/pom.xml": pom, "tool/build.gradle.kts": gradle} {
		full := filepath.Join(ws, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func TestDeclaredReadsPomsAndGradleScripts(t *testing.T) {
	res, err := Declared(writeWorkspace(t), protocol.DeclaredParams{Projects: projects})
	if err != nil {
		t.Fatal(err)
	}
	type row struct {
		Project, Name, Spec, Section string
		Rewritable                   bool
	}
	var got []row
	for _, d := range res.Declarations {
		got = append(got, row{d.Project, d.Name, d.Spec, d.Section, d.Rewritable})
	}
	want := []row{
		{"svc", "com.acme:inherited", "${acme.version}", "dependencies", false},
		{"svc", "com.google.guava:guava", "32.0.0-jre", "dependencyManagement", true},
		{"svc", "org.apache.commons:commons-lang3", "3.12.0", "dependencies", true},
		{"svc", "org.junit.jupiter:junit-jupiter", "5.10.0", "property:junit.version", true},
		{"svc", "org.slf4j:slf4j-api", "[1.7,2.0)", "dependencies", false},
		{"tool", "org.apache.commons:commons-lang3", "3.12.0", "gradle", true},
		{"tool", "org.junit.jupiter:junit-jupiter", "5.10.0", "gradle", true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("declarations:\n got %+v\nwant %+v", got, want)
	}
}

func TestRewriteEditsTheVersionTextOrTheProperty(t *testing.T) {
	ws := writeWorkspace(t)
	declared, _ := Declared(ws, protocol.DeclaredParams{Projects: projects})
	var changes []protocol.Rewrite
	for _, d := range declared.Declarations {
		if d.Rewritable {
			changes = append(changes, protocol.Rewrite{Declaration: d, Spec: "9.9.9"})
		}
	}
	res, err := Rewrite(ws, protocol.RewriteParams{Changes: changes})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 2 {
		t.Errorf("files: %v", res.Files)
	}
	read := func(rel string) string {
		raw, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		return string(raw)
	}
	p := read("svc/pom.xml")
	for _, want := range []string{
		"<junit.version>9.9.9</junit.version>", "<java.version>21</java.version>",
		"<artifactId>commons-lang3</artifactId>\n      <version>9.9.9</version>",
		"<artifactId>guava</artifactId>\n        <version>9.9.9</version>",
		"<version>[1.7,2.0)</version>", "<version>${junit.version}</version>",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("pom lacks %q:\n%s", want, p)
		}
	}
	g := read("tool/build.gradle.kts")
	for _, want := range []string{`implementation("org.apache.commons:commons-lang3:9.9.9")`, `testImplementation 'org.junit.jupiter:junit-jupiter:9.9.9'`, `"com.acme:dynamic:$acmeVersion"`} {
		if !strings.Contains(g, want) {
			t.Errorf("script lacks %q:\n%s", want, g)
		}
	}
}

func TestRewriteRefusesADependencyThatIsGone(t *testing.T) {
	_, err := Rewrite(writeWorkspace(t), protocol.RewriteParams{Changes: []protocol.Rewrite{{
		Declaration: protocol.Declaration{File: "svc/pom.xml", Name: "no:such", Section: SectionDependencies, Spec: "1.0"}, Spec: "2.0",
	}}})
	if err == nil || !strings.Contains(err.Error(), "no longer declares no:such") {
		t.Fatalf("%v", err)
	}
}

const metadata = `<metadata><versioning><versions>
<version>2.0.0-M1</version><version>1.4.9</version><version>2.1.0</version><version>1.5.2</version><version>2.2.0-RC1</version><version>1.4.10</version>
</versions></versioning></metadata>`

func TestLatestTakesTheHighestStableVersionFromMavenCentral(t *testing.T) {
	real := FetchMetadata
	t.Cleanup(func() { FetchMetadata = real })
	var asked []string
	FetchMetadata = func(url string) ([]byte, error) {
		asked = append(asked, url)
		switch {
		case strings.Contains(url, "internal"):
			return nil, nil
		case strings.Contains(url, "offline"):
			return nil, errors.New("no route")
		}
		return []byte(metadata), nil
	}
	res, err := Latest("", protocol.LatestParams{Names: []string{"com.acme:thing", "com.acme:internal", "com.acme:offline"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Latest["com.acme:thing"] != "2.1.0" || len(res.Latest) != 1 || res.Unreachable["com.acme:offline"] != "no route" || len(res.Unreachable) != 1 {
		t.Fatalf("%+v", res)
	}
	if asked[0] != MavenCentral+"/com/acme/thing/maven-metadata.xml" && asked[1] != MavenCentral+"/com/acme/thing/maven-metadata.xml" && asked[2] != MavenCentral+"/com/acme/thing/maven-metadata.xml" {
		t.Errorf("group dots become folders: %v", asked)
	}
}
