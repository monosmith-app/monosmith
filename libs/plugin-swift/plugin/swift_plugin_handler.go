package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-swift/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-swift/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-swift/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-swift/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-swift/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "swift"

// Swift is the Swift language plugin.
type Swift struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Swift { return &Swift{Version: version} }

// Describe implements serve.Plugin.
func (s *Swift) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	s.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: s.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"swift"},
		Manifests: []string{discovery.Manifest},
	}, nil
}

// Discover implements serve.Discoverer.
func (s *Swift) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(s.root, params.Manifests)
}

// Generate implements serve.Generator.
func (s *Swift) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(s.root, params)
}

// WriteVersion implements serve.Versioner.
func (s *Swift) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(s.root, params)
}

// Publish implements serve.Publisher.
func (s *Swift) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(params)
}

// CI implements serve.CIContributor.
func (s *Swift) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(params), nil
}

// ChangeKind implements serve.KindChanger.
func (s *Swift) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(s.root, params)
}

// Doctor implements serve.DoctorContributor.
func (s *Swift) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(s.root, params), nil
}
