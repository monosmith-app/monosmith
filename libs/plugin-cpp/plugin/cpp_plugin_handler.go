package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-cpp/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-cpp/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-cpp/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-cpp/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-cpp/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "cpp"

// Cpp is the C++ language plugin.
type Cpp struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Cpp { return &Cpp{Version: version} }

// Describe implements serve.Plugin.
func (c *Cpp) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	c.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: c.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"cpp", "c++"},
		Manifests: []string{"CMakeLists.txt"},
	}, nil
}

// Discover implements serve.Discoverer.
func (c *Cpp) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(c.root, params.Manifests)
}

// Generate implements serve.Generator.
func (c *Cpp) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(c.root, params)
}

// WriteVersion implements serve.Versioner.
func (c *Cpp) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(c.root, params)
}

// Publish implements serve.Publisher.
func (c *Cpp) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(c.root, params)
}

// CI implements serve.CIContributor.
func (c *Cpp) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(c.root, params), nil
}

// ChangeKind implements serve.KindChanger.
func (c *Cpp) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(c.root, params)
}

// Doctor implements serve.DoctorContributor.
func (c *Cpp) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(c.root, params), nil
}
