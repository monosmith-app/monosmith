package workspace

import (
	"strings"
	"testing"
)

func TestTheGitHubReleaseRegistryNeedsGitHubReleases(t *testing.T) {
	cfg := validConfig()
	cfg.Release.Registry = GitHubReleaseRegistry
	err := validateRegistries(&cfg)
	if err == nil || !strings.Contains(err.Error(), "github_releases = true") {
		t.Fatalf("the packages cannot be attached to releases that are off: %v", err)
	}
	cfg.Release.GitHubReleases = true
	if err := validateRegistries(&cfg); err != nil {
		t.Fatalf("with GitHub Releases on, the value is valid for [release]: %v", err)
	}
	cfg.Release.Registry = ""
	cfg.Projects = map[string]ProjectOverride{"core": {Registry: GitHubReleaseRegistry}}
	if err := validateRegistries(&cfg); err != nil {
		t.Fatalf("and for a project: %v", err)
	}
	cfg.Release.GitHubReleases = false
	if err := validateRegistries(&cfg); err == nil || !strings.Contains(err.Error(), "[projects.core]") {
		t.Fatalf("a project that asks for it names itself: %v", err)
	}
}

func TestTheGitHubReleaseRegistryIsReservedAndNotAnInstallSource(t *testing.T) {
	cfg := validConfig()
	cfg.Registries = map[string]RegistryConfig{GitHubReleaseRegistry: {Auth: "token", TokenEnv: "T", NPM: "https://f/npm/"}}
	if err := validateRegistries(&cfg); err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("the name is reserved: %v", err)
	}
	cfg = validConfig()
	cfg.Workspace.Registry = GitHubReleaseRegistry
	if err := validateRegistries(&cfg); err == nil || !strings.Contains(err.Error(), "installs always come from somewhere") {
		t.Fatalf("installs need a registry: %v", err)
	}
}
