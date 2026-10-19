package registries

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CI is what pipelines need for the packages the projects deploy:
//   - GitHub Packages: the job's GITHUB_TOKEN with packages: write; the user
//     is GITHUB_ACTOR, which Actions sets, so no secret to create
//   - Azure Artifacts: MavenAuthenticate@0, which signs in with the
//     pipeline's own token (no secret)
//   - token: MAVEN_USERNAME and the variable smith.toml names
//   - no registry with a Maven URL: MAVEN_REPOSITORY_URL, MAVEN_USERNAME and
//     MAVEN_PASSWORD, as before
func CI(projects []protocol.Project) protocol.CIResult {
	var result protocol.CIResult
	var feeds, tokens []string
	github, env := false, false
	for _, p := range projects {
		if !p.Publishes() {
			continue
		}
		r := PublishRegistry(p)
		switch {
		case r == nil:
			env = true
		case r.Auth == protocol.AuthGitHubPackages:
			github = true
		case r.Auth == protocol.AuthAzureArtifacts:
			if feed := AzureFeed(r.Maven); feed != "" && !slices.Contains(feeds, feed) {
				feeds = append(feeds, feed)
			}
		case !slices.Contains(tokens, r.TokenEnv):
			tokens = append(tokens, r.TokenEnv)
		}
	}
	if len(feeds) > 0 {
		result.Authenticate = append(result.Authenticate, protocol.Toolchain{Name: "maven-registry", Azure: &protocol.ActionStep{Uses: "MavenAuthenticate@0", Inputs: []protocol.Input{{Name: "artifactsFeeds", Value: strings.Join(feeds, ",")}}}})
	}
	if github {
		result.Credentials = append(result.Credentials, protocol.Credential{Env: "GITHUB_TOKEN", Description: "deploys to GitHub Packages' Maven repository: the job's token with packages: write", Required: true})
		result.Permissions = append(result.Permissions, "packages: write")
	}
	add := func(c protocol.Credential) {
		if !slices.ContainsFunc(result.Credentials, func(x protocol.Credential) bool { return x.Env == c.Env }) {
			result.Credentials = append(result.Credentials, c)
		}
	}
	if len(tokens) > 0 {
		add(protocol.Credential{Env: "MAVEN_USERNAME", Description: "Maven repository user", Required: true})
		for _, name := range tokens {
			add(protocol.Credential{Env: name, Description: "Maven repository password or token ([registries] in smith.toml)", Required: true})
		}
	}
	if env {
		add(protocol.Credential{Env: "MAVEN_REPOSITORY_URL", Description: "Maven repository the workspace's packages deploy to", Required: true})
		add(protocol.Credential{Env: "MAVEN_USERNAME", Description: "Maven repository user", Required: true})
		add(protocol.Credential{Env: "MAVEN_PASSWORD", Description: "Maven repository password or token", Required: true})
	}
	return result
}
