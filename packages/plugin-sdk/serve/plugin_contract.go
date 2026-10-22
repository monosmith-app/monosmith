package serve

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// Plugin is the one interface every plugin implements.
type Plugin interface {
	// Describe returns the plugin's identity, kind, languages and manifests.
	// Serve fills in the protocol version and the capabilities.
	Describe(params protocol.InitializeParams) (protocol.InitializeResult, error)
}

// Discoverer describes projects behind manifests.
type Discoverer interface {
	Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error)
}

// Generator scaffolds a project.
type Generator interface {
	Generate(params protocol.GenerateParams) (protocol.GenerateResult, error)
}

// Versioner records a release version in a manifest.
type Versioner interface {
	WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error)
}

// Publisher publishes a released project.
type Publisher interface {
	Publish(params protocol.PublishParams) (protocol.PublishResult, error)
}

// RegistryReader reports the newest version a registry has of a project.
type RegistryReader interface {
	Published(params protocol.PublishedParams) (protocol.PublishedResult, error)
}

// CIContributor describes what a pipeline needs for the plugin's projects.
type CIContributor interface {
	CI(params protocol.CIParams) (protocol.CIResult, error)
}

// DoctorContributor checks the plugin's projects.
type DoctorContributor interface {
	Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error)
}

// Remover cleans up after a project is removed.
type Remover interface {
	Remove(params protocol.RemoveParams) (protocol.RemoveResult, error)
}

// DependencyManager adds and removes packages in its projects with the
// ecosystem's own toolchain (smith install / uninstall).
type DependencyManager interface {
	Install(params protocol.InstallParams) (protocol.DependencyResult, error)
	Uninstall(params protocol.UninstallParams) (protocol.DependencyResult, error)
}

// DeclarationEditor lists and rewrites the dependency specs its projects'
// manifests declare (smith sync, smith up).
type DeclarationEditor interface {
	Declared(params protocol.DeclaredParams) (protocol.DeclaredResult, error)
	Rewrite(params protocol.RewriteParams) (protocol.DependencyResult, error)
}

// LatestReader asks registries and release feeds for the newest versions
// (smith up).
type LatestReader interface {
	Latest(params protocol.LatestParams) (protocol.LatestResult, error)
}

// ComponentLister lists the components a project ships, read from its
// lockfile or manifest, for the release's SBOM.
type ComponentLister interface {
	Components(params protocol.ComponentsParams) (protocol.ComponentsResult, error)
}

// Contributor is a convention: it adds tasks to the projects it applies to.
type Contributor interface {
	Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error)
}
