package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-bundle/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-bundle/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-bundle/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "bundle"

// Bundle is the bundle plugin.
type Bundle struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Bundle { return &Bundle{Version: version} }

// Describe implements serve.Plugin.
func (b *Bundle) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	b.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: b.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"bundle"},
		Manifests: []string{discovery.ManifestFile},
		Kinds:     []protocol.ProjectKind{protocol.ProjectApp},
	}, nil
}

// Discover implements serve.Discoverer.
func (b *Bundle) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(b.root, params.Manifests)
}

// Generate implements serve.Generator.
func (b *Bundle) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(b.root, params)
}

// WriteVersion implements serve.Versioner.
func (b *Bundle) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(b.root, params)
}

// Publish implements serve.Publisher.
func (b *Bundle) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(params)
}

// PublishesChannels implements serve.ChannelPublisher: Publish uploads nothing: a bundle is released by its tag, and app.json takes any version string.
// No registry holds a stable pointer a prerelease could take.
func (b *Bundle) PublishesChannels() {}
