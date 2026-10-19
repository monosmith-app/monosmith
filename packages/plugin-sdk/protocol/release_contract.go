package protocol

// VersionParams asks a plugin to record a new version in a project manifest.
type VersionParams struct {
	Project Project `json:"project"`
	Version string  `json:"version"`
	// Dependencies maps exported identifiers of other workspace projects to
	// their new versions, so the plugin can update dependency ranges.
	Dependencies map[string]string `json:"dependencies,omitempty"`
	DryRun       bool              `json:"dryRun,omitempty"`
}

// VersionResult lists the files that were (or, in a dry run, would be) changed.
type VersionResult struct {
	Files []string `json:"files"`
}

// PublishParams asks a plugin to publish one project at one version.
type PublishParams struct {
	Project Project `json:"project"`
	Version string  `json:"version"`
	DryRun  bool    `json:"dryRun,omitempty"`
	// Assets, when set, asks the plugin to build the files attached to the
	// release (an app's per-platform binaries) instead of publishing to a
	// registry.
	Assets *AssetSpec `json:"assets,omitempty"`
	// Platform, when set, is the "os/arch" (Go's spelling) of the runner a
	// native project is published from: each native job publishes its own
	// platform's build (a VS Code extension's target) and nothing else.
	Platform string `json:"platform,omitempty"`
	// Tagged says the publish runs on the project's own tag, in the run the
	// registry trusts (pub.dev's automated publishing); a registry that takes
	// a package only from such a run is published to only then.
	Tagged bool `json:"tagged,omitempty"`
	// Commit is the full hash of the commit being released, and Branch the
	// release branch, for a plugin that tags what it publishes with them (a
	// container image's sha-<short> and branch tags). Empty when unknown.
	Commit string `json:"commit,omitempty"`
	Branch string `json:"branch,omitempty"`
}

// AssetSpec describes the release files of an app.
type AssetSpec struct {
	// Name is the file name pattern; placeholders are {product}, {version},
	// {os}, {arch} and {ext}.
	Name string `json:"name"`
	// Product fills {product}; it defaults to the project name.
	Product string `json:"product,omitempty"`
	// Platforms are "os/arch" pairs in the language's own spelling.
	Platforms []string `json:"platforms"`
	// OSAlias and ArchAlias rename an os or arch in file names only.
	OSAlias   map[string]string `json:"osAlias,omitempty"`
	ArchAlias map[string]string `json:"archAlias,omitempty"`
	// VersionVariable is the symbol the version is stamped into, for
	// languages that stamp at link time (Go: main.version).
	VersionVariable string `json:"versionVariable,omitempty"`
	// OutDir is the workspace-relative directory the files go to.
	OutDir string `json:"outDir"`
}

// PublishResult reports what happened. Skipped with a reason is a success:
// an already-published version makes a re-run safe.
type PublishResult struct {
	Published bool   `json:"published"`
	Skipped   string `json:"skipped,omitempty"`
	Registry  string `json:"registry,omitempty"`
	// Assets are the workspace-relative files built for the release, in
	// upload order; in a dry run, the files that would be built.
	Assets []string `json:"assets,omitempty"`
	// AssetPlatforms maps each archive among Assets to the "os/arch" it was
	// built for, so the host can describe them (package manager manifests).
	AssetPlatforms map[string]string `json:"assetPlatforms,omitempty"`
	// Binary is the executable's name inside each archive, without ".exe".
	Binary string `json:"binary,omitempty"`
	// Images are the container images pushed, one per registry, for a deploy
	// step or a GitOps manifest to refer to by digest.
	Images []PublishedImage `json:"images,omitempty"`
}

// PublishedImage is one container image a release pushed.
type PublishedImage struct {
	// Reference is the image without a tag: ghcr.io/acme/api.
	Reference string `json:"reference"`
	// Tags are the tags pushed, in the order they were.
	Tags []string `json:"tags"`
	// Digest is the manifest digest the registry holds ("sha256:..."); empty
	// when it could not be read.
	Digest string `json:"digest,omitempty"`
}

// PublishedParams asks which version of a project its registry has.
type PublishedParams struct {
	Project Project `json:"project"`
}

// PublishedResult is the newest version the registry has of a project.
type PublishedResult struct {
	// Version is empty when the registry does not know the project: its
	// first release is still to come.
	Version string `json:"version,omitempty"`
	// Registry names where the plugin looked, for messages.
	Registry string `json:"registry,omitempty"`
	// Unreachable is true when the registry could not be asked (offline, an
	// error), so nothing can be concluded; Detail says why.
	Unreachable bool   `json:"unreachable,omitempty"`
	Detail      string `json:"detail,omitempty"`
}
