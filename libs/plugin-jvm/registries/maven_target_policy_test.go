package registries

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func env(values map[string]string) func(string) string {
	return func(k string) string { return values[k] }
}

func project(r *protocol.Registry) protocol.Project {
	return protocol.Project{Name: "core", Kind: protocol.ProjectPackage, Registries: &protocol.Registries{Publish: r}}
}

func TestTheGitHubShorthandIsTheRepositorysPackagesURL(t *testing.T) {
	gh := &protocol.Registry{Name: "github", Auth: protocol.AuthGitHubPackages, Maven: protocol.MavenGitHub}
	cases := map[string]struct {
		repository, remote, want string
	}{
		"actions":    {"acme/tools", "", "https://maven.pkg.github.com/acme/tools"},
		"https":      {"", "https://github.com/acme/tools.git", "https://maven.pkg.github.com/acme/tools"},
		"ssh":        {"", "git@github.com:acme/tools.git", "https://maven.pkg.github.com/acme/tools"},
		"ssh scheme": {"", "ssh://git@github.com/acme/tools", "https://maven.pkg.github.com/acme/tools"},
	}
	for name, c := range cases {
		got, ok, err := Resolve(project(gh), env(map[string]string{"GITHUB_REPOSITORY": c.repository}), func() string { return c.remote })
		if err != nil || !ok || got.URL != c.want || got.ServerID != ServerSmith {
			t.Errorf("%s: %+v %v %v", name, got, ok, err)
		}
	}
	if _, _, err := Resolve(project(gh), env(nil), func() string { return "" }); err == nil || !strings.Contains(err.Error(), "GITHUB_REPOSITORY") {
		t.Errorf("no repository to name: %v", err)
	}
}

func TestAnExplicitURLIsUsedAsGiven(t *testing.T) {
	r := &protocol.Registry{Name: "gh", Auth: protocol.AuthGitHubPackages, Maven: "https://maven.pkg.github.com/acme/other/"}
	got, ok, err := Resolve(project(r), env(nil), func() string { return "" })
	if err != nil || !ok || got.URL != "https://maven.pkg.github.com/acme/other" {
		t.Fatalf("%+v %v %v", got, ok, err)
	}
}

func TestAPackageWithoutAMavenRegistryKeepsTheEnvironment(t *testing.T) {
	for _, p := range []protocol.Project{
		{Kind: protocol.ProjectPackage},
		{Kind: protocol.ProjectPackage, Registries: &protocol.Registries{}},
		project(&protocol.Registry{Name: "npm-only", Auth: protocol.AuthToken, TokenEnv: "T", NPM: "https://npm.example"}),
	} {
		if _, ok, err := Resolve(p, env(nil), func() string { return "" }); ok || err != nil {
			t.Errorf("%+v: %v %v", p.Registries, ok, err)
		}
	}
}

func TestAzureArtifactsNamesTheFeedAsTheServer(t *testing.T) {
	r := &protocol.Registry{Name: "libs", Auth: protocol.AuthAzureArtifacts, Maven: "https://pkgs.dev.azure.com/acme/_packaging/libs/maven/v1"}
	got, ok, err := Resolve(project(r), env(nil), func() string { return "" })
	if err != nil || !ok || got.ServerID != "libs" {
		t.Fatalf("%+v %v %v", got, ok, err)
	}
}

func TestCredentialsFollowTheAuth(t *testing.T) {
	gh := Target{Auth: protocol.AuthGitHubPackages}
	user, password, err := gh.Credentials(env(map[string]string{"GITHUB_TOKEN": "tok", "GITHUB_ACTOR": "octocat"}))
	if err != nil || user != "octocat" || password != "tok" {
		t.Errorf("the actor and the job's token: %q %q %v", user, password, err)
	}
	if user, _, _ := gh.Credentials(env(map[string]string{"GITHUB_TOKEN": "tok"})); user == "" {
		t.Error("a user is always given")
	}
	if _, _, err := gh.Credentials(env(nil)); err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Errorf("without the token: %v", err)
	}

	token := Target{Auth: protocol.AuthToken, TokenEnv: "FEED_TOKEN"}
	user, password, err = token.Credentials(env(map[string]string{"MAVEN_USERNAME": "deploy", "FEED_TOKEN": "s3"}))
	if err != nil || user != "deploy" || password != "s3" {
		t.Errorf("MAVEN_USERNAME and the registry's token: %q %q %v", user, password, err)
	}
	if _, _, err := token.Credentials(env(nil)); err == nil || !strings.Contains(err.Error(), "FEED_TOKEN") {
		t.Errorf("without the token: %v", err)
	}

	azure := Target{Auth: protocol.AuthAzureArtifacts}
	if _, _, err := azure.Credentials(env(nil)); err != nil {
		t.Errorf("Maven signs in through MavenAuthenticate@0: %v", err)
	}
}

func TestTheBuildToolsReadTheRepositoryFromTheMavenVariables(t *testing.T) {
	got := strings.Join(Target{URL: "https://repo.example/m"}.Env("u", "p"), " ")
	if got != "MAVEN_REPOSITORY_URL=https://repo.example/m MAVEN_USERNAME=u MAVEN_PASSWORD=p" {
		t.Fatal(got)
	}
}
