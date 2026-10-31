package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-node/dependencies"
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
}

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
	}, nil
}

// Discover implements serve.Discoverer.
func (n *Node) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(n.root, params.Manifests)
}

// Generate implements serve.Generator.
func (n *Node) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(n.root, params)
}

// WriteVersion implements serve.Versioner.
func (n *Node) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(n.root, params)
}

// Publish implements serve.Publisher.
func (n *Node) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(n.root, params)
}

// Published implements serve.RegistryReader.
func (n *Node) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(n.root, params)
}

// CI implements serve.CIContributor.
func (n *Node) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(n.root, params, scaffolding.NodeVersion), nil
}

// Doctor implements serve.DoctorContributor.
func (n *Node) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(n.root, params, scaffolding.NodeVersion), nil
}

// Remove implements serve.Remover.
func (n *Node) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return scaffolding.Remove(n.root, params)
}

// Install implements serve.DependencyManager.
func (n *Node) Install(params protocol.InstallParams) (protocol.DependencyResult, error) {
	return dependencies.Install(n.root, params)
}

// Uninstall implements serve.DependencyManager.
func (n *Node) Uninstall(params protocol.UninstallParams) (protocol.DependencyResult, error) {
	return dependencies.Uninstall(n.root, params)
}

// Declared implements serve.DeclarationEditor.
func (n *Node) Declared(params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	return dependencies.Declared(n.root, params)
}

// Rewrite implements serve.DeclarationEditor.
func (n *Node) Rewrite(params protocol.RewriteParams) (protocol.DependencyResult, error) {
	return dependencies.Rewrite(n.root, params)
}

// Latest implements serve.LatestReader.
func (n *Node) Latest(params protocol.LatestParams) (protocol.LatestResult, error) {
	return dependencies.Latest(n.root, params)
}

// ChangeKind implements serve.KindChanger.
func (n *Node) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(n.root, params)
}

// Components implements serve.ComponentLister.
func (n *Node) Components(params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	return sbom.List(n.root, params)
}
