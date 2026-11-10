package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-dart/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-dart/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-dart/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-dart/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-dart/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-dart/sbom"
	"github.com/monosmith-app/monosmith/libs/plugin-dart/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
)

// Name is the plugin's name.
const Name = "dart"

// Dart is the Dart and Flutter language plugin.
type Dart struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Dart { return &Dart{Version: version} }

// Describe implements serve.Plugin.
func (d *Dart) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	d.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: d.Version},
		Kind:      protocol.KindLanguage,
		Tools:     toolcatalog.For("dart"),
		Languages: []string{"dart", "flutter"},
		Manifests: []string{"pubspec.yaml"},
		Stacks: []protocol.Stack{{
			Name: scaffolding.FlutterStack, Description: "a Flutter app for the web, in the pub workspace", Language: Name,
		}},
	}, nil
}

// DescribesDeclared implements serve.DeclaredDescriber: Discover also describes
// the folders smith.toml declares, as folders of the package above them.
func (d *Dart) DescribesDeclared() {}

// PublishesChannels implements serve.ChannelPublisher: pub.dev orders a
// prerelease below its release, so a plain pub get skips it, and pub takes the
// semver spelling as it is. pub.dev publishes from the package's tag.
func (d *Dart) PublishesChannels() {}

// Discover implements serve.Discoverer.
func (d *Dart) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.DiscoverWithDeclared(d.root, params.Manifests, params.Declared)
}

// Generate implements serve.Generator.
func (d *Dart) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(d.root, params)
}

// Remove implements serve.Remover.
func (d *Dart) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return scaffolding.Remove(d.root, params)
}

// WriteVersion implements serve.Versioner.
func (d *Dart) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(d.root, params)
}

// Publish implements serve.Publisher.
func (d *Dart) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(d.root, params)
}

// Published implements serve.RegistryReader.
func (d *Dart) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(d.root, params)
}

// CI implements serve.CIContributor.
func (d *Dart) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(d.root, params), nil
}

// ChangeKind implements serve.KindChanger.
func (d *Dart) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(d.root, params)
}

// Doctor implements serve.DoctorContributor.
func (d *Dart) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(d.root, params), nil
}

// Install implements serve.DependencyManager.
func (d *Dart) Install(params protocol.InstallParams) (protocol.DependencyResult, error) {
	return dependencies.Install(d.root, params)
}

// Uninstall implements serve.DependencyManager.
func (d *Dart) Uninstall(params protocol.UninstallParams) (protocol.DependencyResult, error) {
	return dependencies.Uninstall(d.root, params)
}

// Declared implements serve.DeclarationEditor.
func (d *Dart) Declared(params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	return dependencies.Declared(d.root, params)
}

// Rewrite implements serve.DeclarationEditor.
func (d *Dart) Rewrite(params protocol.RewriteParams) (protocol.DependencyResult, error) {
	return dependencies.Rewrite(d.root, params)
}

// Latest implements serve.LatestReader.
func (d *Dart) Latest(params protocol.LatestParams) (protocol.LatestResult, error) {
	return dependencies.Latest(d.root, params)
}

// Components implements serve.ComponentLister.
func (d *Dart) Components(params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	return sbom.List(d.root, params)
}
