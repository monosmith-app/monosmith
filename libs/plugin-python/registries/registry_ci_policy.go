package registries

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Of are the registries the projects share; the host sets the same install
// registry on each.
func Of(projects []protocol.Project) *protocol.Registries {
	for _, p := range projects {
		if p.Registries != nil {
			return p.Registries
		}
	}
	return nil
}

// Index is the simple index pip installs from, or "" for PyPI.
func Index(projects []protocol.Project) string {
	if r := Of(projects); r != nil && r.Install != nil {
		return r.Install.PyPI
	}
	return ""
}

// CI is what pipelines need for the projects' Python registries:
//   - Azure Artifacts: PipAuthenticate@1 for installs and TwineAuthenticate@1
//     for uploads, both with the pipeline's own token (no secret)
//   - token: the variable smith.toml names, given to setup when installs read
//     the registry
//
// It also says whether publishing still needs PYPI_TOKEN: only for projects
// published to PyPI itself.
func CI(projects []protocol.Project) (result protocol.CIResult, publicPublish bool) {
	pipFeeds, twineFeed := map[string]bool{}, ""
	var pipOrder []string
	tokens := map[string]bool{}
	var tokenOrder []string
	token := func(r *protocol.Registry, install bool) {
		if _, ok := tokens[r.TokenEnv]; !ok {
			tokenOrder = append(tokenOrder, r.TokenEnv)
		}
		tokens[r.TokenEnv] = tokens[r.TokenEnv] || install
	}
	for _, p := range projects {
		if r := p.Registries; r != nil && r.Install != nil && r.Install.PyPI != "" {
			switch r.Install.Auth {
			case protocol.AuthAzureArtifacts:
				if feed := AzureFeed(r.Install.PyPI); feed != "" && !pipFeeds[feed] {
					pipFeeds[feed] = true
					pipOrder = append(pipOrder, feed)
				}
			case protocol.AuthToken:
				token(r.Install, true)
			}
		}
		if p.Kind != protocol.ProjectPackage {
			continue
		}
		r := p.Registries
		if r == nil || r.Publish == nil || r.Publish.PyPIUpload == "" {
			publicPublish = true
			continue
		}
		switch r.Publish.Auth {
		case protocol.AuthAzureArtifacts:
			if feed := AzureFeed(r.Publish.PyPIUpload); feed != "" {
				twineFeed = feed
			}
		case protocol.AuthToken:
			token(r.Publish, false)
		}
	}
	for _, feed := range pipOrder {
		result.Authenticate = append(result.Authenticate, protocol.Toolchain{
			Name:  "pip-registry " + feed,
			Azure: &protocol.ActionStep{Uses: "PipAuthenticate@1", Inputs: []protocol.Input{{Name: "artifactFeeds", Value: feed}, {Name: "onlyAddExtraIndex", Value: "false"}}},
		})
	}
	if twineFeed != "" {
		result.Authenticate = append(result.Authenticate, protocol.Toolchain{
			Name:  "twine-registry",
			Azure: &protocol.ActionStep{Uses: "TwineAuthenticate@1", Inputs: []protocol.Input{{Name: "artifactFeed", Value: twineFeed}}},
		})
	}
	for _, env := range tokenOrder {
		result.Credentials = append(result.Credentials, protocol.Credential{Env: env, Description: "token for a private Python registry ([registries] in smith.toml)", Required: true, Setup: tokens[env]})
	}
	return result, publicPublish
}
