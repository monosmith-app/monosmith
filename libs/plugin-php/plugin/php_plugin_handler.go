package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-php/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-php/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-php/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-php/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-php/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-php/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "php"

// PHP is the PHP language plugin.
type PHP struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *PHP { return &PHP{Version: version} }

// Describe implements serve.Plugin.
func (p *PHP) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"php"},
		Manifests: []string{"composer.json"},
	}, nil
}

// Discover implements serve.Discoverer.
func (p *PHP) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(p.root, params.Manifests)
}

// Generate implements serve.Generator.
func (p *PHP) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(p.root, params)
}

// Remove implements serve.Remover.
func (p *PHP) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return scaffolding.Remove(p.root, params)
}

// WriteVersion implements serve.Versioner.
func (p *PHP) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(p.root, params)
}

// Publish implements serve.Publisher.
func (p *PHP) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(params)
}

// Published implements serve.RegistryReader.
func (p *PHP) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(params)
}

// CI implements serve.CIContributor.
func (p *PHP) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(p.root, params), nil
}

// ChangeKind implements serve.KindChanger.
func (p *PHP) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(p.root, params)
}

// Doctor implements serve.DoctorContributor.
func (p *PHP) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(p.root, params), nil
}

// Install implements serve.DependencyManager.
func (p *PHP) Install(params protocol.InstallParams) (protocol.DependencyResult, error) {
	return dependencies.Install(p.root, params)
}

// Uninstall implements serve.DependencyManager.
func (p *PHP) Uninstall(params protocol.UninstallParams) (protocol.DependencyResult, error) {
	return dependencies.Uninstall(p.root, params)
}
