package releasing

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const corePom = `<project>
  <parent>
    <groupId>acme</groupId>
    <artifactId>acme-parent</artifactId>
    <version>0.0.0</version>
  </parent>
  <artifactId>acme-core</artifactId>
  <version>0.0.0</version>
  <dependencies>
    <dependency><groupId>org.junit.jupiter</groupId><artifactId>junit-jupiter</artifactId><version>5.11.4</version></dependency>
  </dependencies>
</project>
`

func TestSetVersion(t *testing.T) {
	out, ok := SetVersion(corePom, "1.2.0")
	if !ok || !strings.Contains(out, "<artifactId>acme-core</artifactId>\n  <version>1.2.0</version>") || !strings.Contains(out, "<artifactId>acme-parent</artifactId>\n    <version>0.0.0</version>") || !strings.Contains(out, "<version>5.11.4</version>") {
		t.Fatalf("%v\n%s", ok, out)
	}
	inherited := strings.Replace(corePom, "  <version>0.0.0</version>\n  <dependencies>", "  <dependencies>", 1)
	if out, ok := SetVersion(inherited, "1.2.0"); !ok || !strings.Contains(out, "<artifactId>acme-core</artifactId>\n  <version>1.2.0</version>\n  <dependencies>") {
		t.Fatalf("added:\n%s", out)
	}
}

func TestWriteVersionUpdatesDependents(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		_ = os.MkdirAll(filepath.Join(ws, filepath.Dir(rel)), 0o755)
		_ = os.WriteFile(filepath.Join(ws, rel), []byte(body), 0o644)
	}
	write("libs/core/pom.xml", corePom)
	write("apps/api/pom.xml", "<project><artifactId>acme-api</artifactId><version>0.0.0</version><dependencies><dependency><groupId>acme</groupId><artifactId>acme-core</artifactId><version>0.0.0</version></dependency></dependencies></project>")
	write("dist/maven/acme-core/pom.xml", corePom)
	core := protocol.Project{Root: "libs/core", Exports: []string{"acme:acme-core"}}
	res, err := WriteVersion(ws, protocol.VersionParams{Project: core, Version: "0.1.0"})
	if err != nil || strings.Join(res.Files, ",") != "apps/api/pom.xml,libs/core/pom.xml" {
		t.Fatal(err, res.Files)
	}
	api, _ := os.ReadFile(filepath.Join(ws, "apps/api/pom.xml"))
	if !strings.Contains(string(api), "<artifactId>acme-api</artifactId><version>0.0.0</version>") || !strings.Contains(string(api), "<artifactId>acme-core</artifactId><version>0.1.0</version>") {
		t.Fatalf("api:\n%s", api)
	}
}

func TestPublish(t *testing.T) {
	var ran []string
	real := MavenRunner
	defer func() { MavenRunner = real }()
	MavenRunner = func(_ string, argv []string, _ ...string) error { ran = argv; return nil }
	core := protocol.Project{Name: "core", Root: "libs/core", Kind: protocol.ProjectPackage, Exports: []string{"dev.acme:acme-core"}}
	if _, err := Publish(t.TempDir(), protocol.PublishParams{Project: core}); err == nil {
		t.Fatal("a repository is required")
	}
	t.Setenv("MAVEN_REPOSITORY_URL", "https://repo.example/maven")
	ws := t.TempDir()
	if res, err := Publish(ws, protocol.PublishParams{Project: core, Version: "1.0.0"}); err != nil || !res.Published {
		t.Fatal(err, res)
	}
	if got := strings.Join(ran, " "); !strings.Contains(got, "-pl dev.acme:acme-core deploy -DskipTests -s ") || !strings.HasSuffix(got, "-DaltDeploymentRepository=smith::https://repo.example/maven") {
		t.Fatal(got)
	}
}

func TestAPrereleaseIsWrittenAsIsDeployedAsIsAndNeverReportedAsTheRelease(t *testing.T) {
	out, ok := SetVersion(corePom, "2.0.0-beta.3")
	if !ok || !strings.Contains(out, "<artifactId>acme-core</artifactId>\n  <version>2.0.0-beta.3</version>") {
		t.Fatalf("Maven takes the semver prerelease as it is: %v\n%s", ok, out)
	}
	dep, _ := SetDependencyVersion(`<dependency><groupId>acme</groupId><artifactId>acme-core</artifactId><version>1.0.0</version></dependency>`, "acme", "acme-core", "2.0.0-beta.3")
	if !strings.Contains(dep, "<version>2.0.0-beta.3</version>") {
		t.Fatalf("a dependent names the prerelease: %s", dep)
	}
	var ran []string
	real := MavenRunner
	defer func() { MavenRunner = real }()
	MavenRunner = func(_ string, argv []string, _ ...string) error { ran = argv; return nil }
	t.Setenv("MAVEN_REPOSITORY_URL", "https://repo.example/maven")
	core := protocol.Project{Name: "core", Root: "libs/core", Kind: protocol.ProjectPackage, Exports: []string{"dev.acme:acme-core"}}
	if res, err := Publish(t.TempDir(), protocol.PublishParams{Project: core, Version: "2.0.0-beta.3", Tag: "beta"}); err != nil || !res.Published || !strings.Contains(strings.Join(ran, " "), "deploy") {
		t.Fatal("a repository has no channel tag: the prerelease is deployed like a release", err, res, ran)
	}
	// Maven names the last deployed version <release>, a prerelease too.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<metadata><versioning><latest>2.0.0-beta.3</latest><release>2.0.0-beta.3</release><versions><version>1.9.0</version><version>1.10.0</version><version>2.0.0-beta.3</version></versions></versioning></metadata>`))
	}))
	defer srv.Close()
	MavenCentral = srv.URL
	t.Setenv("MAVEN_REPOSITORY_URL", "")
	if res, _ := Published(protocol.PublishedParams{Project: core}); res.Version != "1.10.0" {
		t.Fatalf("the newest release is not the prerelease deployed last: %+v", res)
	}
}

func TestPublished(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/dev/acme/acme-core/maven-metadata.xml" {
			_, _ = w.Write([]byte(`<metadata><versioning><latest>2.0.0-rc1</latest><release>1.4.0</release></versioning></metadata>`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	MavenCentral = srv.URL
	core := protocol.Project{Kind: protocol.ProjectPackage, Exports: []string{"dev.acme:acme-core"}}
	if res, _ := Published(protocol.PublishedParams{Project: core}); res.Version != "1.4.0" {
		t.Fatalf("%+v", res)
	}
	core.Exports[0] = "dev.acme:other"
	if res, _ := Published(protocol.PublishedParams{Project: core}); res.Version != "" || res.Unreachable {
		t.Fatalf("%+v", res)
	}
}
