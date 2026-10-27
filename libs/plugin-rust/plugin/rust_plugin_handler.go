package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-rust/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-rust/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-rust/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-rust/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "rust"

// Rust is the Rust language plugin.
type Rust struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Rust { return &Rust{Version: version} }

// Describe implements serve.Plugin.
func (r *Rust) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	r.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: r.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"rust"},
		Manifests: []string{"Cargo.toml"},
	}, nil
}

// Discover implements serve.Discoverer.
func (r *Rust) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(r.root, params.Manifests)
}

// Generate implements serve.Generator.
func (r *Rust) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(r.root, params)
}

// WriteVersion implements serve.Versioner.
func (r *Rust) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(r.root, params)
}

// Publish implements serve.Publisher.
func (r *Rust) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(r.root, params)
}

// Published implements serve.RegistryReader.
func (r *Rust) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(params)
}

// CI implements serve.CIContributor.
func (r *Rust) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(r.root, params), nil
}

// Doctor implements serve.DoctorContributor.
func (r *Rust) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(r.root, params), nil
}

// Remove implements serve.Remover.
func (r *Rust) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return scaffolding.Remove(r.root, params)
}

// ChangeKind implements serve.KindChanger.
func (r *Rust) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(r.root, params)
}
