package protocol

// Registry is a package registry smith.toml names in [registries.<name>]: a
// private feed (Azure Artifacts, GitHub Packages, any token-protected
// registry) a language plugin installs from or publishes to. Each URL is
// optional; a plugin reads the one for its ecosystem.
type Registry struct {
	Name string `json:"name"`
	// Auth is how the registry is signed in to: AuthAzureArtifacts,
	// AuthGitHubPackages or AuthToken.
	Auth string `json:"auth"`
	// TokenEnv is the environment variable holding the token, for AuthToken.
	TokenEnv string `json:"tokenEnv,omitempty"`
	// NPM is the npm registry URL.
	NPM string `json:"npm,omitempty"`
	// PyPI is the simple index pip installs from; PyPIUpload is where twine
	// uploads.
	PyPI       string `json:"pypi,omitempty"`
	PyPIUpload string `json:"pypiUpload,omitempty"`
	// NuGet is the NuGet v3 service index.
	NuGet string `json:"nuget,omitempty"`
}

// Registry sign-in kinds.
const (
	// AuthAzureArtifacts signs in with the pipeline's own token, through
	// Azure Pipelines' authenticate tasks: no secret to create.
	AuthAzureArtifacts = "azure-artifacts"
	// AuthGitHubPackages signs in with the workflow's GITHUB_TOKEN.
	AuthGitHubPackages = "github-packages"
	// AuthToken reads a token from the environment variable TokenEnv.
	AuthToken = "token"
)

// Registries are the registries a project uses, set by the host from
// smith.toml: Install is where the workspace installs packages from, Publish
// where the project is published. Nil, or a nil member, means the public
// registry, as before.
type Registries struct {
	Install *Registry `json:"install,omitempty"`
	Publish *Registry `json:"publish,omitempty"`
}
