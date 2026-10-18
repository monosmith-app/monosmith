package registries

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CI is what pipelines need for the projects' NuGet feeds:
//   - Azure Artifacts: NuGetAuthenticate@1, which installs the credential
//     provider and signs in with the pipeline's own token (no secret)
//   - GitHub Packages: GITHUB_TOKEN, given to setup too, and packages: read
//     (or write, when publishing there)
//   - token: the variable smith.toml names, given to setup when restores
//     read the feed
//
// It also says whether publishing still needs NUGET_API_KEY: only for
// packages pushed to nuget.org.
func CI(projects []protocol.Project) (result protocol.CIResult, publicPublish bool) {
	azure, ghRead, ghWrite := false, false, false
	tokens := map[string]bool{}
	var order []string
	use := func(r *protocol.Registry, install bool) {
		if r == nil || r.NuGet == "" {
			return
		}
		switch r.Auth {
		case protocol.AuthAzureArtifacts:
			azure = true
		case protocol.AuthGitHubPackages:
			ghRead = ghRead || install
			ghWrite = ghWrite || !install
		case protocol.AuthToken:
			if _, ok := tokens[r.TokenEnv]; !ok {
				order = append(order, r.TokenEnv)
			}
			tokens[r.TokenEnv] = tokens[r.TokenEnv] || install
		}
	}
	for _, p := range projects {
		if p.Registries != nil {
			use(p.Registries.Install, true)
		}
		if p.Kind != protocol.ProjectPackage {
			continue
		}
		if p.Registries == nil || p.Registries.Publish == nil || p.Registries.Publish.NuGet == "" {
			publicPublish = true
			continue
		}
		use(p.Registries.Publish, false)
	}
	if azure {
		result.Authenticate = append(result.Authenticate, protocol.Toolchain{Name: "nuget-registry", Azure: &protocol.ActionStep{Uses: "NuGetAuthenticate@1"}})
	}
	if ghRead || ghWrite {
		result.Credentials = append(result.Credentials, protocol.Credential{Env: "GITHUB_TOKEN", Description: "signs in to GitHub Packages' NuGet feed", Required: true, Setup: ghRead})
		if ghWrite {
			result.Permissions = append(result.Permissions, "packages: write")
		} else {
			result.Permissions = append(result.Permissions, "packages: read")
		}
	}
	for _, env := range order {
		result.Credentials = append(result.Credentials, protocol.Credential{Env: env, Description: "token for a private NuGet feed ([registries] in smith.toml)", Required: true, Setup: tokens[env]})
	}
	return result, publicPublish
}

// Push is where dotnet nuget push sends a package and the key it gives:
// the feed smith.toml names, with a placeholder key for Azure Artifacts (the
// credential provider NuGetAuthenticate@1 installs signs the push) and the
// token for the others. ok is false for a package without a named feed.
func Push(p protocol.Project, getenv func(string) string) (source, key string, ok bool) {
	r := p.Registries
	if r == nil || r.Publish == nil || r.Publish.NuGet == "" {
		return "", "", false
	}
	if r.Publish.Auth == protocol.AuthAzureArtifacts {
		return r.Publish.NuGet, "az", true
	}
	return r.Publish.NuGet, getenv(TokenEnv(r.Publish)), true
}
