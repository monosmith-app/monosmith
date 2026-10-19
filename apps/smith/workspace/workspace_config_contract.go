package workspace

// FileName is the workspace marker and configuration file.
const FileName = "smith.toml"

// SchemaVersion is the smith.toml schema this build reads and writes.
const SchemaVersion = 1

// Config is smith.toml. Every field is optional except workspace.name.
type Config struct {
	Schema    int              `toml:"schema"`
	Workspace WorkspaceSection `toml:"workspace"`
	CI        CISection        `toml:"ci"`
	Release   ReleaseSection   `toml:"release"`
	Cache     CacheSection     `toml:"cache"`
	Plugins   PluginsSection   `toml:"plugins"`
	// Conventions apply convention plugins, by name, to the projects of the
	// languages they support.
	Conventions map[string]ConventionConfig `toml:"conventions"`
	// Stacks declares stacks by name, next to the ones plugins publish; a
	// stack declared here wins over a published one of the same name.
	Stacks map[string]StackConfig `toml:"stacks"`
	// DefaultStacks is the stack of a project that records none, and of
	// `smith add` without --stack, by language.
	DefaultStacks map[string]string `toml:"default_stacks"`
	// Recipes names the recipe for one project kind (app, lib or package),
	// instead of [workspace] recipe.
	Recipes  map[string]string          `toml:"recipes"`
	Tasks    map[string]TaskOverride    `toml:"tasks"`
	Projects map[string]ProjectOverride `toml:"projects"`
	// Registries are private package registries, by name, that the
	// workspace installs from ([workspace] registry) and publishes to
	// ([release] registry, [projects.<name>] registry).
	Registries map[string]RegistryConfig `toml:"registries"`
}

// RegistryConfig is one private registry: how to sign in to it, and its URL
// for each ecosystem it serves. Every URL is optional.
type RegistryConfig struct {
	// Auth is "azure-artifacts" (the pipeline's own token, through Azure
	// Pipelines' authenticate tasks), "github-packages" (GITHUB_TOKEN) or
	// "token" (the variable TokenEnv names).
	Auth string `toml:"auth"`
	// TokenEnv names the environment variable holding the token, for auth
	// "token".
	TokenEnv string `toml:"token_env"`
	// NPM is the npm registry URL.
	NPM string `toml:"npm"`
	// PyPI is the simple index pip installs from (.../simple/).
	PyPI string `toml:"pypi"`
	// PyPIUpload is where twine uploads; default: PyPI with /simple/
	// replaced by /upload/, as Azure Artifacts names them.
	PyPIUpload string `toml:"pypi_upload"`
	// NuGet is the NuGet v3 service index (.../index.json).
	NuGet string `toml:"nuget"`
	// Maven is the Maven repository packages deploy to: an https URL, or
	// "github" for GitHub Packages (https://maven.pkg.github.com/<owner>/<repo>
	// of the repository the release runs in; needs auth = "github-packages").
	Maven string `toml:"maven"`
}

// StackConfig is a stack declared in smith.toml: the plugins a project is
// built with, by name.
type StackConfig struct {
	Description string   `toml:"description"`
	Language    string   `toml:"language"`
	Toolchain   []string `toml:"toolchain"`
	Structure   []string `toml:"structure"`
	Scaffold    string   `toml:"scaffold"`
	Deploy      []string `toml:"deploy"`
}

// WorkspaceSection names the monorepo.
type WorkspaceSection struct {
	Name string `toml:"name"`
	// Scope is the npm scope for packages, e.g. "@acme".
	Scope string `toml:"scope"`
	// ModulePrefix is the Go module path prefix, e.g. github.com/acme/repo.
	ModulePrefix string `toml:"module_prefix"`
	// Exclude are workspace-relative globs ("examples/**") of directories
	// whose manifests are not projects: discovery leaves them out in every
	// language, and a change there affects no project.
	Exclude []string `toml:"exclude"`
	// Registry names the [registries] entry the workspace installs from;
	// empty means the public registries.
	Registry string `toml:"registry"`
	// Recipe is the recipe new projects are created from: a path relative
	// to smith.toml, or an https URL (docs/reference/recipes.md). Empty means
	// stacks only. `smith add --stack` bypasses it.
	Recipe string `toml:"recipe"`
}

// CISection chooses which pipelines smith generates.
type CISection struct {
	// Providers is any of "github", "azure", "gitlab". Default: ["github"].
	Providers []string `toml:"providers"`
	// Devcontainer writes .devcontainer/devcontainer.json with the
	// toolchains CI installs, at the same versions. Default: false (smith new
	// turns it on).
	Devcontainer bool `toml:"devcontainer"`
	// NativeRunners are the machines the native job runs on, one per OS,
	// as the provider names them. Default: ubuntu-latest, windows-latest,
	// macos-latest.
	NativeRunners []string `toml:"native_runners"`
	// NativeLinuxPackages are apt packages the native job installs on
	// Linux before building (libraries a cgo app links).
	NativeLinuxPackages []string `toml:"native_linux_packages"`
}

// ReleaseSection configures `smith release`.
type ReleaseSection struct {
	// Branch is the only branch a real release runs from. Default: main.
	Branch string `toml:"branch"`
	// Remote receives release tags. Default: origin.
	Remote string `toml:"remote"`
	// GitHubReleases creates a GitHub Release for every tag, with the
	// release's commits as notes and an app's assets attached. Needs
	// GITHUB_TOKEN (or GH_TOKEN). Default: false.
	GitHubReleases bool `toml:"github_releases"`
	// Registry names the [registries] entry packages are published to;
	// empty means [workspace] registry, and then the public registries.
	Registry string `toml:"registry"`
}

// CacheSection configures the local task cache.
type CacheSection struct {
	// Dir is workspace-relative. Default: .smith/cache.
	Dir string `toml:"dir"`
	// MaxSizeMB triggers pruning of the least recently used entries. Default: 2048.
	MaxSizeMB int `toml:"max_size_mb"`
	// Remote is a shared cache behind the local one: https://, s3:// or
	// azblob://. Objects are signed with SMITH_CACHE_KEY. Default: none.
	Remote string `toml:"remote"`
	// RemoteReadOnly reads the remote without writing to it (a developer's
	// machine, while CI fills it). Default: false.
	RemoteReadOnly bool `toml:"remote_read_only"`
}

// PluginsSection chooses plugins.
type PluginsSection struct {
	// Enabled lists built-in plugins to load. Default: all of them.
	Enabled []string `toml:"enabled"`
	// External plugins are executables speaking the plugin protocol.
	External map[string]ExternalPlugin `toml:"external"`
}

// ExternalPlugin is a plugin outside the smith binary. It replaces a built-in
// plugin of the same name.
type ExternalPlugin struct {
	// Command runs the plugin. Empty means: find it installed on the plugin
	// search path (docs/reference/plugins.md).
	Command []string `toml:"command"`
	// Version constrains the plugin's version: "1.2.3", "^1.2", "~1.2",
	// ">=1.2". Empty means any.
	Version string `toml:"version"`
}

// ConventionConfig chooses the projects a convention applies to.
type ConventionConfig struct {
	// Languages narrows the convention to some of the languages it supports.
	// Default: all of them.
	Languages []string `toml:"languages"`
	// Exclude lists projects the convention leaves alone.
	Exclude []string `toml:"exclude"`
	// Registry is where the docker convention's release pushes images
	// (ghcr.io/acme). Unset: images are built, never pushed. Docker only.
	Registry string `toml:"registry"`
	// Registries is the list form: every image is pushed to each, as
	// "host/namespace" strings or tables with url, username_env and
	// password_env. Setting both Registry and Registries is an error.
	// Docker only.
	Registries []any `toml:"registries"`
	// DescriptionFrom is a workspace-relative file (README.md) whose text
	// becomes the repository description on Docker Hub. Docker only.
	DescriptionFrom string `toml:"description_from"`
	// ShortDescription is the one-line summary on Docker Hub (100 characters
	// at most). Unset: the project's own description. Docker only.
	ShortDescription string `toml:"short_description"`
	// TagCommit also tags every image sha-<short commit>, and TagBranch with
	// the release branch's name, for deployment platforms and GitOps agents
	// that follow them. Docker only.
	TagCommit *bool `toml:"tag_commit"`
	TagBranch *bool `toml:"tag_branch"`
	// Push turns the docker convention's push off for every app; default
	// true when a registry is set. Docker only.
	Push *bool `toml:"push"`
	// Verify false stops smith ci verify building the affected apps'
	// images on pull requests; default true. Docker only.
	Verify *bool `toml:"verify"`
}

// DockerOverride is one app's docker settings ([projects.<name>.docker]).
type DockerOverride struct {
	// Image is the image's name under the registry; default the project's
	// name, lowercase.
	Image string `toml:"image"`
	// Registry and Registries replace the convention's for this app (the
	// same forms; setting both is an error).
	Registry   string `toml:"registry"`
	Registries []any  `toml:"registries"`
	// DescriptionFrom replaces the convention's description file.
	DescriptionFrom string `toml:"description_from"`
	// ShortDescription replaces the convention's short description.
	ShortDescription string `toml:"short_description"`
	// TagCommit and TagBranch replace the convention's.
	TagCommit *bool `toml:"tag_commit"`
	TagBranch *bool `toml:"tag_branch"`
	// Push false builds the image but never pushes it on release.
	Push *bool `toml:"push"`
	// Context is the build context, workspace-relative: "." (the workspace,
	// the default) or a directory such as the app's own, for a Dockerfile
	// written for it (Nx's convention: docker build apps/api).
	Context string `toml:"context"`
	// Verify false stops smith ci verify building this app's image.
	Verify *bool `toml:"verify"`
}

// ProjectOverride adjusts one project.
type ProjectOverride struct {
	// TagPrefix replaces the release tag prefix ("<name>@" by default), so a
	// converted project keeps its existing tags.
	TagPrefix string `toml:"tag_prefix"`
	// BareTags makes the project's release tags the bare version (0.1.0),
	// for a registry that takes the tag name as the version (JitPack). Only
	// one project can; it excludes tag_prefix.
	BareTags bool `toml:"bare_tags"`
	// Stack replaces the stack the project's manifest records.
	Stack string `toml:"stack"`
	// Release makes an app releasable: it is versioned by its tag and its
	// assets are built and attached to the release.
	Release *ProjectRelease `toml:"release"`
	// SplitRepo is a GitHub repository (owner/repo) each release of the
	// project is mirrored to: its directory's history split out, pushed to
	// the mirror's main branch and tagged vX.Y.Z there, for a registry that
	// reads a package at a repository root (Packagist). Needs SPLIT_REPO_TOKEN.
	SplitRepo string `toml:"split_repo"`
	// Registry names the [registries] entry this package is published to,
	// instead of [release] registry.
	Registry string `toml:"registry"`
	// Docker is the app's settings for the docker convention.
	Docker *DockerOverride `toml:"docker"`
	// Settings are plugins' settings for this project, by plugin name
	// ([projects.<name>.settings.<plugin>]): what a recipe's settings became
	// when it created the project. A plugin reads its own table.
	Settings map[string]map[string]any `toml:"settings"`
	// Recipe records the recipe the project was created from.
	Recipe *ProjectRecipe          `toml:"recipe"`
	Tasks  map[string]TaskOverride `toml:"tasks"`
}

// ProjectRecipe is the recipe a project was created from, written by
// `smith add` ([projects.<name>.recipe]): enough to resolve the same recipe
// again with the same answers.
type ProjectRecipe struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	// Source is the recipe as [workspace] recipe named it: a path or a URL.
	Source string `toml:"source"`
	// Answers are the choices as they were answered, by choice name.
	Answers map[string]string `toml:"answers"`
}

// DefaultPlatforms are the platforms an app's assets are built for.
var DefaultPlatforms = []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64", "windows/arm64"}

// ProjectRelease configures the release of an app.
type ProjectRelease struct {
	// AssetName is the file name pattern of each asset; placeholders are
	// {product}, {version}, {os}, {arch} and {ext}.
	// Default: {product}_{version}_{os}_{arch}.{ext}.
	AssetName string `toml:"asset_name"`
	// Product fills {product}. Default: the project name.
	Product string `toml:"product"`
	// Platforms are os/arch pairs. Default: DefaultPlatforms.
	Platforms []string `toml:"platforms"`
	// OSAlias and ArchAlias rename an os or arch in asset names.
	OSAlias   map[string]string `toml:"os_alias"`
	ArchAlias map[string]string `toml:"arch_alias"`
	// VersionVariable is the symbol the version is stamped into. Default for
	// Go: main.version.
	VersionVariable string `toml:"version_variable"`
	// PackageManagers are the manifests generated from the assets and
	// attached to the GitHub Release: homebrew, scoop, winget.
	PackageManagers []string `toml:"package_managers"`
	// LinuxPackages are the Linux packages built from the linux archives
	// and attached to the GitHub Release: deb, rpm.
	LinuxPackages []string `toml:"linux_packages"`
	// NPM are the npm packages the app is published as, each installing its
	// binary; the first also names the per-platform packages holding the
	// binaries (<first>-<os>-<cpu>). Publishing needs NODE_AUTH_TOKEN.
	NPM []string `toml:"npm"`
	// PyPI is the PyPI project the app is published as, one wheel per
	// platform. Publishing needs PYPI_TOKEN.
	PyPI string `toml:"pypi"`
	// NuGet is the .NET tool package the app is published as, with one
	// <NuGet>.<rid> package per platform holding the binary. Publishing
	// needs NUGET_API_KEY.
	NuGet string `toml:"nuget"`
	// RubyGems is the gem the app is published as, one platform gem per
	// platform. Publishing needs GEM_HOST_API_KEY.
	RubyGems string `toml:"rubygems"`
	// Pub is the pub.dev package the app is published as, holding every
	// platform's binary. pub.dev accepts it only from a GitHub Actions run on
	// a tag matching the package's tag pattern, so the release pushes PubTag
	// and starts the generated pub workflow on it.
	Pub string `toml:"pub"`
	// PubTag is that tag, with {name} (the release tag prefix without its
	// "@") and {version}. Default: {name}-v{version}.
	PubTag string `toml:"pub_tag"`
	// Description, Homepage and License describe the app in those manifests
	// and packages.
	Description string `toml:"description"`
	Homepage    string `toml:"homepage"`
	License     string `toml:"license"`
}

// TaskOverride replaces individual fields of a plugin-inferred task. Unset
// fields keep the plugin's value.
type TaskOverride struct {
	Command []string `toml:"command"`
	// Cwd is workspace-relative ("." is the workspace root); empty keeps the
	// project root.
	Cwd       string            `toml:"cwd"`
	Inputs    []string          `toml:"inputs"`
	Outputs   []string          `toml:"outputs"`
	DependsOn []string          `toml:"depends_on"`
	Env       map[string]string `toml:"env"`
	Cache     *bool             `toml:"cache"`
	// Verify runs the task in smith ci verify on pull requests, for the
	// affected projects; false turns off what a plugin or smith marked.
	Verify *bool `toml:"verify"`
	// Serve, ServeURL and ServeTimeout start an app around the task (see
	// protocol.Task).
	Serve        string `toml:"serve"`
	ServeURL     string `toml:"serve_url"`
	ServeTimeout string `toml:"serve_timeout"`
}
