package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-objc/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-objc/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-objc/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-objc/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-objc/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "objc"

// ObjC is the Objective-C language plugin.
type ObjC struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *ObjC { return &ObjC{Version: version} }

// Describe implements serve.Plugin.
func (o *ObjC) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	o.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: o.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"objc"},
		Manifests: []string{discovery.ManifestPattern},
		// A podspec describes a library.
		Kinds: []protocol.ProjectKind{protocol.ProjectLib, protocol.ProjectPackage},
	}, nil
}

// Discover implements serve.Discoverer.
func (o *ObjC) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(o.root, params.Manifests)
}

// Generate implements serve.Generator.
func (o *ObjC) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(o.root, params)
}

// WriteVersion implements serve.Versioner.
func (o *ObjC) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(o.root, params)
}

// Publish implements serve.Publisher.
func (o *ObjC) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(params)
}

// CI implements serve.CIContributor.
func (o *ObjC) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(params), nil
}

// ChangeKind implements serve.KindChanger.
func (o *ObjC) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(o.root, params)
}

// Doctor implements serve.DoctorContributor.
func (o *ObjC) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(o.root, params), nil
}
