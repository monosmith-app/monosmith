package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-python/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-python/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-python/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-python/linking"
	"github.com/monosmith-app/monosmith/libs/plugin-python/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-python/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-python/sbom"
	"github.com/monosmith-app/monosmith/libs/plugin-python/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Python is the pip-native Python language plugin.
type Python struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Python { return &Python{Version: version} }

// Describe implements serve.Plugin.
func (p *Python) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "python", Version: p.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"python", "py"},
		Manifests: []string{"pyproject.toml"},
	}, nil
}

// Discover implements serve.Discoverer.
func (p *Python) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.DiscoverWithDeclared(p.root, params.Manifests, params.Declared)
}

// DescribesDeclared implements serve.DeclaredDescriber: Discover also describes
// the projects smith.toml declares, as folders of the distribution above them.
func (p *Python) DescribesDeclared() {}

// Generate implements serve.Generator.
func (p *Python) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(p.root, params)
}

// WriteVersion implements serve.Versioner.
func (p *Python) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(p.root, params)
}

// Publish implements serve.Publisher.
func (p *Python) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(p.root, params)
}

// PublishesChannels implements serve.ChannelPublisher: PyPI orders a
// prerelease (2.0.0b1) below its release, so a channel never becomes the version
// pip installs; Publish needs no tag.
func (p *Python) PublishesChannels() {}

// Published implements serve.RegistryReader.
func (p *Python) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(params)
}

// CI implements serve.CIContributor.
func (p *Python) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(params), nil
}

// Doctor implements serve.DoctorContributor.
func (p *Python) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(p.root, params), nil
}

// Remove implements serve.Remover.
func (p *Python) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return scaffolding.Remove(p.root, params)
}

// Install implements serve.DependencyManager.
func (p *Python) Install(params protocol.InstallParams) (protocol.DependencyResult, error) {
	return dependencies.Install(p.root, params)
}

// Link implements serve.Linker.
func (p *Python) Link(params protocol.LinkParams) (protocol.LinkResult, error) {
	return linking.Link(p.root, params)
}

// Unlink implements serve.Linker.
func (p *Python) Unlink(params protocol.UnlinkParams) (protocol.DependencyResult, error) {
	return linking.Unlink(p.root, params)
}

// Uninstall implements serve.DependencyManager.
func (p *Python) Uninstall(params protocol.UninstallParams) (protocol.DependencyResult, error) {
	return dependencies.Uninstall(p.root, params)
}

// Declared implements serve.DeclarationEditor.
func (p *Python) Declared(params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	return dependencies.Declared(p.root, params)
}

// Rewrite implements serve.DeclarationEditor.
func (p *Python) Rewrite(params protocol.RewriteParams) (protocol.DependencyResult, error) {
	return dependencies.Rewrite(p.root, params)
}

// Latest implements serve.LatestReader.
func (p *Python) Latest(params protocol.LatestParams) (protocol.LatestResult, error) {
	return dependencies.Latest(p.root, params)
}

// ChangeKind implements serve.KindChanger.
func (p *Python) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(p.root, params)
}

// Components implements serve.ComponentLister.
func (p *Python) Components(params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	return sbom.List(p.root, params)
}
