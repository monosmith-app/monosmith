package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-dart/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-dart/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-dart/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-dart/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-dart/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
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
		Languages: []string{"dart", "flutter"},
		Manifests: []string{"pubspec.yaml"},
		Stacks: []protocol.Stack{{
			Name: scaffolding.FlutterStack, Description: "a Flutter app for the web, in the pub workspace", Language: Name,
		}},
	}, nil
}

// Discover implements serve.Discoverer.
func (d *Dart) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(d.root, params.Manifests)
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
