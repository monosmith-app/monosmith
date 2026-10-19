package registries

import (
	"errors"
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// GitHubPackagesMaven is the host of GitHub Packages' Maven repositories.
const GitHubPackagesMaven = "https://maven.pkg.github.com"

// ServerSmith is the settings.xml server smith's own settings define; its
// credentials come from MAVEN_USERNAME and MAVEN_PASSWORD.
const ServerSmith = "smith"

// Target is the Maven repository a package deploys to.
type Target struct {
	// URL is the repository's URL.
	URL string
	// Auth is how the repository is signed in to: a protocol.Auth* kind.
	Auth string
	// TokenEnv names the variable holding the password, for protocol.AuthToken.
	TokenEnv string
	// ServerID is the settings.xml server Maven reads the credentials from:
	// the feed for Azure Artifacts (MavenAuthenticate@0 writes it), else
	// ServerSmith.
	ServerID string
}

// PublishRegistry is the registry smith.toml names for the project's Maven
// deploys, or nil when it names none (MAVEN_REPOSITORY_URL then applies).
func PublishRegistry(p protocol.Project) *protocol.Registry {
	if p.Registries == nil || p.Registries.Publish == nil || p.Registries.Publish.Maven == "" {
		return nil
	}
	return p.Registries.Publish
}

// Resolve is the repository smith.toml names for the project, or ok false
// when it names none. getenv reads the environment and remote the URL of the
// workspace's git remote, for the "github" shorthand outside GitHub Actions.
func Resolve(p protocol.Project, getenv func(string) string, remote func() string) (t Target, ok bool, err error) {
	r := PublishRegistry(p)
	if r == nil {
		return Target{}, false, nil
	}
	t = Target{Auth: r.Auth, TokenEnv: r.TokenEnv, ServerID: ServerSmith}
	if r.Maven == protocol.MavenGitHub {
		repo := getenv("GITHUB_REPOSITORY")
		if repo == "" {
			repo = remote()
		}
		owner, name, err := ownerAndName(repo)
		if err != nil {
			return Target{}, false, fmt.Errorf("registry %s: maven = %q needs the GitHub repository: set GITHUB_REPOSITORY or add a GitHub remote (%w)", r.Name, protocol.MavenGitHub, err)
		}
		t.URL = GitHubPackagesMaven + "/" + owner + "/" + name
	} else {
		t.URL = strings.TrimSuffix(r.Maven, "/")
	}
	if r.Auth == protocol.AuthAzureArtifacts {
		if feed := AzureFeed(t.URL); feed != "" {
			t.ServerID = feed
		}
	}
	return t, true, nil
}

// Credentials are the user and password the repository takes: the actor and
// the job's GITHUB_TOKEN for GitHub Packages, MAVEN_USERNAME and the
// registry's token variable for a token, and the pipeline's access token
// for Azure Artifacts (Maven itself gets that from MavenAuthenticate@0; the
// password is only for Gradle, which reads SYSTEM_ACCESSTOKEN).
func (t Target) Credentials(getenv func(string) string) (user, password string, err error) {
	switch t.Auth {
	case protocol.AuthGitHubPackages:
		if getenv("GITHUB_TOKEN") == "" {
			return "", "", errors.New("GITHUB_TOKEN is not set: GitHub Packages takes the job's token with packages: write")
		}
		user = getenv("GITHUB_ACTOR")
		if user == "" {
			user = "x-access-token"
		}
		return user, getenv("GITHUB_TOKEN"), nil
	case protocol.AuthAzureArtifacts:
		return "AzureDevOps", getenv("SYSTEM_ACCESSTOKEN"), nil
	default:
		if getenv(t.TokenEnv) == "" {
			return "", "", fmt.Errorf("%s is not set: the token of the Maven registry", t.TokenEnv)
		}
		return getenv("MAVEN_USERNAME"), getenv(t.TokenEnv), nil
	}
}

// Env is what the build tools read the repository from, in the variables
// the Maven settings and the Gradle scripts smith generates already use.
func (t Target) Env(user, password string) []string {
	return []string{"MAVEN_REPOSITORY_URL=" + t.URL, "MAVEN_USERNAME=" + user, "MAVEN_PASSWORD=" + password}
}

// AzureFeed is the feed in an Azure Artifacts Maven URL (the segment after
// _packaging), or empty.
func AzureFeed(repository string) string {
	parts := strings.Split(repository, "/")
	for i, p := range parts {
		if p == "_packaging" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// ownerAndName reads owner and repository from "owner/name" or from a git
// remote URL (https, ssh or scp-like).
func ownerAndName(remote string) (owner, name string, err error) {
	s := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(remote), "/"), ".git")
	if _, rest, found := strings.Cut(s, "://"); found {
		_, s, _ = strings.Cut(rest, "/")
	} else if _, rest, found := strings.Cut(s, "@"); found {
		_, s, _ = strings.Cut(rest, ":")
	}
	parts := strings.Split(s, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("cannot read owner/repository from %q", remote)
	}
	return parts[0], parts[1], nil
}
