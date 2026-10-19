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

func githubCore() protocol.Project {
	return protocol.Project{
		Name: "core", Root: "libs/core", Kind: protocol.ProjectPackage, Exports: []string{"dev.acme:acme-core"},
		Registries: &protocol.Registries{Publish: &protocol.Registry{Name: "github", Auth: protocol.AuthGitHubPackages, Maven: protocol.MavenGitHub}},
	}
}

func TestPublishToGitHubPackagesUsesTheJobsToken(t *testing.T) {
	var ran, ranEnv []string
	real := MavenRunner
	defer func() { MavenRunner = real }()
	MavenRunner = func(_ string, argv []string, env ...string) error { ran, ranEnv = argv, env; return nil }
	t.Setenv("MAVEN_REPOSITORY_URL", "https://ignored.example")
	t.Setenv("GITHUB_REPOSITORY", "acme/tools")
	t.Setenv("GITHUB_TOKEN", "job-token")
	t.Setenv("GITHUB_ACTOR", "octocat")
	res, err := Publish(t.TempDir(), protocol.PublishParams{Project: githubCore(), Version: "1.0.0"})
	if err != nil || !res.Published || res.Registry != "https://maven.pkg.github.com/acme/tools" {
		t.Fatalf("%+v %v", res, err)
	}
	if got := strings.Join(ran, " "); !strings.Contains(got, " -s ") || !strings.HasSuffix(got, "-DaltDeploymentRepository=smith::https://maven.pkg.github.com/acme/tools") {
		t.Errorf("argv: %s", got)
	}
	if got := strings.Join(ranEnv, " "); got != "MAVEN_REPOSITORY_URL=https://maven.pkg.github.com/acme/tools MAVEN_USERNAME=octocat MAVEN_PASSWORD=job-token" {
		t.Errorf("env: %s", got)
	}
}

func TestPublishToGitHubPackagesNeedsTheToken(t *testing.T) {
	real := MavenRunner
	defer func() { MavenRunner = real }()
	MavenRunner = func(string, []string, ...string) error { t.Fatal("must not deploy without a token"); return nil }
	t.Setenv("GITHUB_REPOSITORY", "acme/tools")
	t.Setenv("GITHUB_TOKEN", "")
	if _, err := Publish(t.TempDir(), protocol.PublishParams{Project: githubCore()}); err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Fatalf("%v", err)
	}
	// A dry run needs neither the token nor a deploy.
	res, err := Publish(t.TempDir(), protocol.PublishParams{Project: githubCore(), DryRun: true})
	if err != nil || res.Skipped != "dry run" || res.Registry != "https://maven.pkg.github.com/acme/tools" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestPublishToAnAzureFeedLeavesTheSignInToMavenAuthenticate(t *testing.T) {
	var ran []string
	real := MavenRunner
	defer func() { MavenRunner = real }()
	MavenRunner = func(_ string, argv []string, _ ...string) error { ran = argv; return nil }
	p := githubCore()
	p.Registries.Publish = &protocol.Registry{Name: "libs", Auth: protocol.AuthAzureArtifacts, Maven: "https://pkgs.dev.azure.com/acme/_packaging/libs/maven/v1"}
	if _, err := Publish(t.TempDir(), protocol.PublishParams{Project: p}); err != nil {
		t.Fatal(err)
	}
	got := strings.Join(ran, " ")
	if strings.Contains(got, " -s ") || !strings.HasSuffix(got, "-DaltDeploymentRepository=libs::https://pkgs.dev.azure.com/acme/_packaging/libs/maven/v1") {
		t.Fatal(got)
	}
}

func TestPublishGradleToATokenRegistry(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "libs", "core"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "libs", "core", "build.gradle.kts"), nil, 0o644)
	var ranEnv []string
	real := GradleRunner
	defer func() { GradleRunner = real }()
	GradleRunner = func(_ string, _ []string, env ...string) error { ranEnv = env; return nil }
	t.Setenv("MAVEN_USERNAME", "deploy")
	t.Setenv("NEXUS_TOKEN", "s3cret")
	p := githubCore()
	p.Registries.Publish = &protocol.Registry{Name: "nexus", Auth: protocol.AuthToken, TokenEnv: "NEXUS_TOKEN", Maven: "https://nexus.example/repository/maven/"}
	if res, err := Publish(ws, protocol.PublishParams{Project: p}); err != nil || !res.Published {
		t.Fatal(res, err)
	}
	if got := strings.Join(ranEnv, " "); got != "MAVEN_REPOSITORY_URL=https://nexus.example/repository/maven MAVEN_USERNAME=deploy MAVEN_PASSWORD=s3cret" {
		t.Fatal(got)
	}
}

func TestPublishedReadsTheRegistryWithTheJobsToken(t *testing.T) {
	var user, password string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		user, password, ok = r.BasicAuth()
		if !ok || r.URL.Path != "/acme/tools/dev/acme/acme-core/maven-metadata.xml" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`<metadata><versioning><release>2.3.0</release></versioning></metadata>`))
	}))
	defer srv.Close()
	t.Setenv("GITHUB_TOKEN", "job-token")
	t.Setenv("GITHUB_ACTOR", "octocat")
	p := githubCore()
	p.Registries.Publish = &protocol.Registry{Name: "github", Auth: protocol.AuthGitHubPackages, Maven: srv.URL + "/acme/tools"}
	res, err := Published(protocol.PublishedParams{Project: p})
	if err != nil || res.Version != "2.3.0" || res.Registry != srv.URL+"/acme/tools" || user != "octocat" || password != "job-token" {
		t.Fatalf("%+v %v %q %q", res, err, user, password)
	}
}
