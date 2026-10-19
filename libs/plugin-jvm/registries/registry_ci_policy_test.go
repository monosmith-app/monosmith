package registries

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func envNames(creds []protocol.Credential) []string {
	var out []string
	for _, c := range creds {
		out = append(out, c.Env)
	}
	return out
}

func TestGitHubPackagesNeedsTheJobsTokenAndNoMavenSecrets(t *testing.T) {
	res := CI([]protocol.Project{project(&protocol.Registry{Name: "github", Auth: protocol.AuthGitHubPackages, Maven: protocol.MavenGitHub})})
	if got := envNames(res.Credentials); len(got) != 1 || got[0] != "GITHUB_TOKEN" {
		t.Errorf("credentials: %v", got)
	}
	if len(res.Permissions) != 1 || res.Permissions[0] != "packages: write" {
		t.Errorf("permissions: %v", res.Permissions)
	}
	if len(res.Authenticate) != 0 {
		t.Errorf("no sign-in step: %+v", res.Authenticate)
	}
}

func TestWithoutARegistryTheMavenSecretsStay(t *testing.T) {
	res := CI([]protocol.Project{{Name: "core", Kind: protocol.ProjectPackage}})
	got := envNames(res.Credentials)
	if len(got) != 3 || got[0] != "MAVEN_REPOSITORY_URL" || got[1] != "MAVEN_USERNAME" || got[2] != "MAVEN_PASSWORD" || len(res.Permissions) != 0 {
		t.Fatalf("%v %v", got, res.Permissions)
	}
	if res := CI([]protocol.Project{{Name: "api", Kind: protocol.ProjectApp}}); len(res.Credentials) != 0 {
		t.Errorf("an app is not deployed: %v", res.Credentials)
	}
}

func TestATokenRegistryAsksForItsToken(t *testing.T) {
	res := CI([]protocol.Project{project(&protocol.Registry{Name: "nexus", Auth: protocol.AuthToken, TokenEnv: "NEXUS_TOKEN", Maven: "https://nexus.example/repository/maven"})})
	if got := envNames(res.Credentials); len(got) != 2 || got[0] != "MAVEN_USERNAME" || got[1] != "NEXUS_TOKEN" || len(res.Permissions) != 0 {
		t.Fatalf("%v %v", got, res.Permissions)
	}
}

func TestAzureArtifactsSignsInWithoutASecret(t *testing.T) {
	r := &protocol.Registry{Name: "libs", Auth: protocol.AuthAzureArtifacts, Maven: "https://pkgs.dev.azure.com/acme/_packaging/libs/maven/v1"}
	res := CI([]protocol.Project{project(r), project(r)})
	if len(res.Credentials) != 0 || len(res.Authenticate) != 1 {
		t.Fatalf("%+v", res)
	}
	step := res.Authenticate[0].Azure
	if step.Uses != "MavenAuthenticate@0" || step.Inputs[0].Name != "artifactsFeeds" || step.Inputs[0].Value != "libs" {
		t.Fatalf("%+v", step)
	}
}

func TestProjectsOnDifferentRegistriesGetEachOnesSignIn(t *testing.T) {
	gh := project(&protocol.Registry{Name: "github", Auth: protocol.AuthGitHubPackages, Maven: protocol.MavenGitHub})
	plain := protocol.Project{Name: "other", Kind: protocol.ProjectPackage}
	got := envNames(CI([]protocol.Project{gh, plain}).Credentials)
	if len(got) != 4 || got[0] != "GITHUB_TOKEN" || got[3] != "MAVEN_PASSWORD" {
		t.Fatalf("%v", got)
	}
}
