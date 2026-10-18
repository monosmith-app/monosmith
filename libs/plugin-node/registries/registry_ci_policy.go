package registries

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CI is what pipelines need to install from and publish to the projects'
// npm registries:
//   - Azure Artifacts: npmAuthenticate@0 on the .npmrc, which signs in with
//     the pipeline's own token (no secret)
//   - GitHub Packages: GITHUB_TOKEN, given to setup too, and packages: read
//     (or write, when publishing there)
//   - token: the variable smith.toml names, given to setup when installs
//     read the registry
//
// It also says whether publishing still needs NODE_AUTH_TOKEN: only for
// packages published to npmjs.org.
func CI(projects []protocol.Project) (result protocol.CIResult, publicPublish bool) {
	azure, ghRead, ghWrite := false, false, false
	tokens := map[string]bool{}
	var order []string
	use := func(r *protocol.Registry, install bool) {
		if r == nil || r.NPM == "" {
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
		if p.Registries == nil || p.Registries.Publish == nil || p.Registries.Publish.NPM == "" {
			publicPublish = true
			continue
		}
		use(p.Registries.Publish, false)
	}
	if azure {
		result.Authenticate = append(result.Authenticate, protocol.Toolchain{
			Name:  "npm-registry",
			Azure: &protocol.ActionStep{Uses: "npmAuthenticate@0", Inputs: []protocol.Input{{Name: "workingFile", Value: ".npmrc"}}},
		})
	}
	if ghRead || ghWrite {
		result.Credentials = append(result.Credentials, protocol.Credential{Env: "GITHUB_TOKEN", Description: "signs in to GitHub Packages' npm registry", Required: true, Setup: ghRead})
		if ghWrite {
			result.Permissions = append(result.Permissions, "packages: write")
		} else {
			result.Permissions = append(result.Permissions, "packages: read")
		}
	}
	for _, env := range order {
		result.Credentials = append(result.Credentials, protocol.Credential{Env: env, Description: "token for a private npm registry ([registries] in smith.toml)", Required: true, Setup: tokens[env]})
	}
	return result, publicPublish
}
