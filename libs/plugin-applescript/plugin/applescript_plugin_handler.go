package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-applescript/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-applescript/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-applescript/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-applescript/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-applescript/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "applescript"

// AppleScript is the AppleScript language plugin.
type AppleScript struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *AppleScript { return &AppleScript{Version: version} }

// Describe implements serve.Plugin.
func (a *AppleScript) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	a.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: a.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"applescript"},
		Manifests: []string{discovery.ManifestPattern},
		// Nothing publishes AppleScript.
		Kinds: []protocol.ProjectKind{protocol.ProjectApp, protocol.ProjectLib},
	}, nil
}

// Discover implements serve.Discoverer.
func (a *AppleScript) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(a.root, params.Manifests)
}

// Generate implements serve.Generator.
func (a *AppleScript) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(a.root, params)
}

// WriteVersion implements serve.Versioner.
func (a *AppleScript) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(a.root, params)
}

// Publish implements serve.Publisher.
func (a *AppleScript) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(params)
}

// PublishesChannels implements serve.ChannelPublisher: Publish uploads nothing: a script is released by its tag, and the version is a comment.
// No registry holds a stable pointer a prerelease could take.
func (a *AppleScript) PublishesChannels() {}

// CI implements serve.CIContributor.
func (a *AppleScript) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(params), nil
}

// ChangeKind implements serve.KindChanger.
func (a *AppleScript) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(a.root, params)
}

// Doctor implements serve.DoctorContributor.
func (a *AppleScript) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(a.root, params), nil
}
