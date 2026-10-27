package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name.
const Name = "jvm"

// JVM is the JVM language plugin: Java with Maven, Kotlin and Android with
// Gradle.
type JVM struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *JVM { return &JVM{Version: version} }

// Describe implements serve.Plugin.
func (j *JVM) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	j.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: j.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"jvm", "java"},
		Manifests: []string{"pom.xml", "build.gradle.kts", "build.gradle"},
		Stacks: []protocol.Stack{
			{Name: scaffolding.KotlinStack, Description: "a Kotlin module built with Gradle, linted with ktlint", Language: Name},
			{Name: scaffolding.AndroidStack, Description: "an Android app in Kotlin, built with Gradle (AGP), linted with ktlint and Android Lint", Language: Name},
		},
	}, nil
}

// Discover implements serve.Discoverer.
func (j *JVM) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(j.root, params.Manifests)
}

// Generate implements serve.Generator.
func (j *JVM) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(j.root, params)
}

// Remove implements serve.Remover.
func (j *JVM) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return scaffolding.Remove(j.root, params)
}

// WriteVersion implements serve.Versioner.
func (j *JVM) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(j.root, params)
}

// Publish implements serve.Publisher.
func (j *JVM) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(j.root, params)
}

// Published implements serve.RegistryReader.
func (j *JVM) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(params)
}

// CI implements serve.CIContributor.
func (j *JVM) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(j.root, params), nil
}

// Doctor implements serve.DoctorContributor.
func (j *JVM) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(j.root, params), nil
}

// ChangeKind implements serve.KindChanger.
func (j *JVM) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(j.root, params)
}
