package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-go/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-go/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-go/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-go/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-go/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-go/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Go is the Go language plugin.
type Go struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Go { return &Go{Version: version} }

// Describe implements serve.Plugin.
func (g *Go) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	g.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "go", Version: g.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"go", "golang"},
		Manifests: []string{"go.mod"},
	}, nil
}

// Discover implements serve.Discoverer.
func (g *Go) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(g.root, params.Manifests)
}

// Generate implements serve.Generator.
func (g *Go) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(g.root, params)
}

// Publish implements serve.Publisher.
func (g *Go) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(g.root, params)
}

// CI implements serve.CIContributor.
func (g *Go) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(g.root, params), nil
}

// Doctor implements serve.DoctorContributor.
func (g *Go) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(g.root, params), nil
}

// Remove implements serve.Remover.
func (g *Go) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return pipeline.Remove(g.root, params)
}

// Install implements serve.DependencyManager.
func (g *Go) Install(params protocol.InstallParams) (protocol.DependencyResult, error) {
	return dependencies.Install(g.root, params)
}

// Uninstall implements serve.DependencyManager.
func (g *Go) Uninstall(params protocol.UninstallParams) (protocol.DependencyResult, error) {
	return dependencies.Uninstall(g.root, params)
}

// Declared implements serve.DeclarationEditor.
func (g *Go) Declared(params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	return dependencies.Declared(g.root, params)
}

// Rewrite implements serve.DeclarationEditor.
func (g *Go) Rewrite(params protocol.RewriteParams) (protocol.DependencyResult, error) {
	return dependencies.Rewrite(g.root, params)
}

// Latest implements serve.LatestReader: the newest module releases and the
// tools the plugin pins.
func (g *Go) Latest(params protocol.LatestParams) (protocol.LatestResult, error) {
	return dependencies.Latest(g.root, params)
}

// ChangeKind implements serve.KindChanger.
func (g *Go) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(g.root, params)
}
