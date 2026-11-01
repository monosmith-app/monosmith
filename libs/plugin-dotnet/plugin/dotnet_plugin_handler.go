package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/sbom"
	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "dotnet"

// Dotnet is the .NET language plugin.
type Dotnet struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Dotnet { return &Dotnet{Version: version} }

// Describe implements serve.Plugin.
func (d *Dotnet) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	d.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: d.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"dotnet", "csharp", "fsharp"},
		Manifests: []string{"*.csproj", "*.fsproj"},
	}, nil
}

// Discover implements serve.Discoverer.
func (d *Dotnet) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(d.root, params.Manifests)
}

// Generate implements serve.Generator.
func (d *Dotnet) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(d.root, params)
}

// WriteVersion implements serve.Versioner.
func (d *Dotnet) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(d.root, params)
}

// Publish implements serve.Publisher.
func (d *Dotnet) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(d.root, params)
}

// PublishesChannels implements serve.ChannelPublisher: NuGet orders a
// prerelease (2.0.0-beta.1) below its release, so a channel never becomes the
// version a plain install picks; Publish needs no tag.
func (d *Dotnet) PublishesChannels() {}

// Published implements serve.RegistryReader.
func (d *Dotnet) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(params)
}

// CI implements serve.CIContributor.
func (d *Dotnet) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(d.root, params), nil
}

// Install implements serve.DependencyManager.
func (d *Dotnet) Install(params protocol.InstallParams) (protocol.DependencyResult, error) {
	return dependencies.Install(d.root, params)
}

// Uninstall implements serve.DependencyManager.
func (d *Dotnet) Uninstall(params protocol.UninstallParams) (protocol.DependencyResult, error) {
	return dependencies.Uninstall(d.root, params)
}

// Declared implements serve.DeclarationEditor.
func (d *Dotnet) Declared(params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	return dependencies.Declared(d.root, params)
}

// Rewrite implements serve.DeclarationEditor.
func (d *Dotnet) Rewrite(params protocol.RewriteParams) (protocol.DependencyResult, error) {
	return dependencies.Rewrite(d.root, params)
}

// Latest implements serve.LatestReader.
func (d *Dotnet) Latest(params protocol.LatestParams) (protocol.LatestResult, error) {
	return dependencies.Latest(d.root, params)
}

// Doctor implements serve.DoctorContributor.
func (d *Dotnet) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(d.root, params), nil
}

// ChangeKind implements serve.KindChanger.
func (d *Dotnet) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(d.root, params)
}

// Components implements serve.ComponentLister.
func (d *Dotnet) Components(params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	return sbom.List(d.root, params)
}
