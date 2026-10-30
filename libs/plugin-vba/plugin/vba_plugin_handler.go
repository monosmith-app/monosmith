package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-vba/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-vba/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-vba/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-vba/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "vba"

// VBA is the VBA language plugin.
type VBA struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *VBA { return &VBA{Version: version} }

// Describe implements serve.Plugin.
func (v *VBA) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	v.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: v.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"vba"},
		Manifests: discovery.Manifests,
		// Nothing publishes VBA.
		Kinds: []protocol.ProjectKind{protocol.ProjectApp, protocol.ProjectLib},
	}, nil
}

// Discover implements serve.Discoverer.
func (v *VBA) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(v.root, params.Manifests)
}

// Generate implements serve.Generator.
func (v *VBA) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(v.root, params)
}

// WriteVersion implements serve.Versioner.
func (v *VBA) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(v.root, params)
}

// Publish implements serve.Publisher.
func (v *VBA) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(params)
}

// ChangeKind implements serve.KindChanger.
func (v *VBA) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(v.root, params)
}
