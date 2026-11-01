package plugin

import (
	"sync"

	"github.com/monosmith-app/monosmith/libs/plugin-node/apisurface"
	"github.com/monosmith-app/monosmith/libs/plugin-node/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-node/deriving"
	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-node/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-node/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-node/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-node/sbom"
	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Node is the Node.js/TypeScript language plugin.
type Node struct {
	Version string
	root    string
	// mu keeps every handler but Published to one at a time, as a plugin that
	// declares no concurrency is served. Published only reads the registry,
	// so the host may ask for many projects at once.
	mu sync.Mutex
}

// publishedConcurrency is how many registry lookups the host may run at once.
const publishedConcurrency = 8

// New returns the plugin; version is reported in the handshake.
func New(version string) *Node { return &Node{Version: version} }

// Describe implements serve.Plugin.
func (n *Node) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	n.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "node", Version: n.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"node", "javascript", "js"},
		Manifests: []string{"package.json"},
		// Concurrency lets the host ask the registry about many projects at
		// once; the other handlers still run one at a time (see Node.mu).
		Concurrency: publishedConcurrency,
	}, nil
}

// Discover implements serve.Discoverer.
func (n *Node) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return discovery.Discover(n.root, params.Manifests)
}

// Generate implements serve.Generator.
func (n *Node) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return scaffolding.Generate(n.root, params)
}

// WriteVersion implements serve.Versioner.
func (n *Node) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return releasing.WriteVersion(n.root, params)
}

// Publish implements serve.Publisher.
func (n *Node) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return releasing.Publish(n.root, params)
}

// PublishesChannels implements serve.ChannelPublisher: Publish passes a release
// channel's name to npm as the dist-tag.
func (n *Node) PublishesChannels() {}

// Published implements serve.RegistryReader.
func (n *Node) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(n.root, params)
}

// CI implements serve.CIContributor.
func (n *Node) CI(params protocol.CIParams) (protocol.CIResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return pipeline.CI(n.root, params, scaffolding.NodeVersion), nil
}

// Doctor implements serve.DoctorContributor.
func (n *Node) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return pipeline.Doctor(n.root, params, scaffolding.NodeVersion), nil
}

// Remove implements serve.Remover.
func (n *Node) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return scaffolding.Remove(n.root, params)
}

// Install implements serve.DependencyManager.
func (n *Node) Install(params protocol.InstallParams) (protocol.DependencyResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return dependencies.Install(n.root, params)
}

// Uninstall implements serve.DependencyManager.
func (n *Node) Uninstall(params protocol.UninstallParams) (protocol.DependencyResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return dependencies.Uninstall(n.root, params)
}

// Declared implements serve.DeclarationEditor.
func (n *Node) Declared(params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return dependencies.Declared(n.root, params)
}

// Rewrite implements serve.DeclarationEditor.
func (n *Node) Rewrite(params protocol.RewriteParams) (protocol.DependencyResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return dependencies.Rewrite(n.root, params)
}

// Latest implements serve.LatestReader.
func (n *Node) Latest(params protocol.LatestParams) (protocol.LatestResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return dependencies.Latest(n.root, params)
}

// Derive implements serve.Deriver.
func (n *Node) Derive(params protocol.DeriveParams) (protocol.DeriveResult, error) {
	return deriving.Derive(n.root, params)
}

// ChangeKind implements serve.KindChanger.
func (n *Node) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return kindchange.Change(n.root, params)
}

// APISurface implements serve.APISurfaceReader.
func (n *Node) APISurface(params protocol.APISurfaceParams) (protocol.APISurfaceResult, error) {
	return apisurface.List(n.root, params)
}

// Components implements serve.ComponentLister.
func (n *Node) Components(params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return sbom.List(n.root, params)
}
