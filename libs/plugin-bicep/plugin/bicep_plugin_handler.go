package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-bicep/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-bicep/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-bicep/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-bicep/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-bicep/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "bicep"

// Bicep is the Bicep language plugin.
type Bicep struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Bicep { return &Bicep{Version: version} }

// Describe implements serve.Plugin.
func (b *Bicep) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	b.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: b.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"bicep"},
		Manifests: []string{discovery.Manifest},
		// Publishing to a Bicep registry is not supported yet.
		Kinds: []protocol.ProjectKind{protocol.ProjectApp, protocol.ProjectLib},
	}, nil
}

// Discover implements serve.Discoverer.
func (b *Bicep) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(b.root, params.Manifests)
}

// Generate implements serve.Generator.
func (b *Bicep) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(b.root, params)
}

// WriteVersion implements serve.Versioner.
func (b *Bicep) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(b.root, params)
}

// Publish implements serve.Publisher.
func (b *Bicep) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(params)
}

// CI implements serve.CIContributor.
func (b *Bicep) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(params), nil
}

// ChangeKind implements serve.KindChanger.
func (b *Bicep) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(b.root, params)
}

// Doctor implements serve.DoctorContributor.
func (b *Bicep) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(b.root, params), nil
}
