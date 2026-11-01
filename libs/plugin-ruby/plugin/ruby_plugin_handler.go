package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-ruby/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-ruby/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-ruby/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-ruby/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-ruby/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-ruby/sbom"
	"github.com/monosmith-app/monosmith/libs/plugin-ruby/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "ruby"

// Ruby is the Ruby language plugin.
type Ruby struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Ruby { return &Ruby{Version: version} }

// Describe implements serve.Plugin.
func (r *Ruby) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	r.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: r.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"ruby"},
		Manifests: []string{"*.gemspec"},
	}, nil
}

// DescribesDeclared implements serve.DeclaredDescriber: Discover also describes
// the folders smith.toml declares, as folders of the gem above them.
func (r *Ruby) DescribesDeclared() {}

// PublishesChannels implements serve.ChannelPublisher: RubyGems orders a
// prerelease (2.0.0.pre.beta.1) below its release, a plain install skips it, and
// WriteVersion and Publish spell it that way.
func (r *Ruby) PublishesChannels() {}

// Discover implements serve.Discoverer.
func (r *Ruby) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.DiscoverWithDeclared(r.root, params.Manifests, params.Declared)
}

// Generate implements serve.Generator.
func (r *Ruby) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(r.root, params)
}

// Remove implements serve.Remover.
func (r *Ruby) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return scaffolding.Remove(r.root, params)
}

// WriteVersion implements serve.Versioner.
func (r *Ruby) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(r.root, params)
}

// Publish implements serve.Publisher.
func (r *Ruby) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(r.root, params)
}

// Published implements serve.RegistryReader.
func (r *Ruby) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(params)
}

// CI implements serve.CIContributor.
func (r *Ruby) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(r.root, params), nil
}

// ChangeKind implements serve.KindChanger.
func (r *Ruby) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(r.root, params)
}

// Doctor implements serve.DoctorContributor.
func (r *Ruby) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(r.root, params), nil
}

// Install implements serve.DependencyManager.
func (r *Ruby) Install(params protocol.InstallParams) (protocol.DependencyResult, error) {
	return dependencies.Install(r.root, params)
}

// Uninstall implements serve.DependencyManager.
func (r *Ruby) Uninstall(params protocol.UninstallParams) (protocol.DependencyResult, error) {
	return dependencies.Uninstall(r.root, params)
}

// Declared implements serve.DeclarationEditor.
func (r *Ruby) Declared(params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	return dependencies.Declared(r.root, params)
}

// Rewrite implements serve.DeclarationEditor.
func (r *Ruby) Rewrite(params protocol.RewriteParams) (protocol.DependencyResult, error) {
	return dependencies.Rewrite(r.root, params)
}

// Latest implements serve.LatestReader.
func (r *Ruby) Latest(params protocol.LatestParams) (protocol.LatestResult, error) {
	return dependencies.Latest(r.root, params)
}

// Components implements serve.ComponentLister.
func (r *Ruby) Components(params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	return sbom.List(r.root, params)
}
