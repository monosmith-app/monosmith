package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-recipe/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-recipe/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-recipe/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "recipe"

// Recipe is the recipe language plugin.
type Recipe struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Recipe { return &Recipe{Version: version} }

// Describe implements serve.Plugin.
func (r *Recipe) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	r.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: r.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"recipe"},
		Manifests: []string{discovery.Manifest},
	}, nil
}

// Discover implements serve.Discoverer.
func (r *Recipe) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(r.root, params.Manifests)
}

// Generate implements serve.Generator.
func (r *Recipe) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(r.root, params)
}

// WriteVersion implements serve.Versioner.
func (r *Recipe) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(r.root, params)
}

// Publish implements serve.Publisher.
func (r *Recipe) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(r.root, params)
}

// PublishesChannels implements serve.ChannelPublisher: Publish uploads nothing: a recipe is released by its tag, and its file is the release's asset.
// No registry holds a stable pointer a prerelease could take.
func (r *Recipe) PublishesChannels() {}
