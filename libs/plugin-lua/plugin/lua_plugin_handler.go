package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-lua/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-lua/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-lua/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-lua/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-lua/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "lua"

// Lua is the Lua language plugin.
type Lua struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Lua { return &Lua{Version: version} }

// Describe implements serve.Plugin.
func (l *Lua) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	l.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: l.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"lua"},
		Manifests: []string{discovery.ManifestPattern},
		// Publishing to luarocks.org is not supported yet.
		Kinds: []protocol.ProjectKind{protocol.ProjectApp, protocol.ProjectLib},
	}, nil
}

// Discover implements serve.Discoverer.
func (l *Lua) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.DiscoverWithDeclared(l.root, params.Manifests, params.Declared)
}

// DescribesDeclared implements serve.DeclaredDescriber: Discover also describes
// the projects smith.toml declares, as folders of the rock above them.
func (l *Lua) DescribesDeclared() {}

// Generate implements serve.Generator.
func (l *Lua) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(l.root, params)
}

// WriteVersion implements serve.Versioner.
func (l *Lua) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(l.root, params)
}

// Publish implements serve.Publisher.
func (l *Lua) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(params)
}

// PublishesChannels implements serve.ChannelPublisher: Publish uploads nothing,
// and LuaRocks orders a prerelease word (2.0.0-beta.1-1) below the release.
func (l *Lua) PublishesChannels() {}

// CI implements serve.CIContributor.
func (l *Lua) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(params), nil
}

// ChangeKind implements serve.KindChanger.
func (l *Lua) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(l.root, params)
}

// Doctor implements serve.DoctorContributor.
func (l *Lua) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(l.root, params), nil
}
