package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/kindchange"
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/pipeline"
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-jvm/sbom"
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

// PublishesChannels implements serve.ChannelPublisher: Maven and Gradle order
// 2.0.0-beta.1 below 2.0.0, both take that spelling as it is, and Published never
// reports a prerelease as the newest release. The plugin has no
// DescribesDeclared: a Maven or Gradle module is one build unit, and a folder of
// packages inside it cannot be built or tested on its own.
func (j *JVM) PublishesChannels() {}

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

// Declared implements serve.DeclarationEditor.
func (j *JVM) Declared(params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	return dependencies.Declared(j.root, params)
}

// Rewrite implements serve.DeclarationEditor.
func (j *JVM) Rewrite(params protocol.RewriteParams) (protocol.DependencyResult, error) {
	return dependencies.Rewrite(j.root, params)
}

// Latest implements serve.LatestReader.
func (j *JVM) Latest(params protocol.LatestParams) (protocol.LatestResult, error) {
	return dependencies.Latest(j.root, params)
}

// Doctor implements serve.DoctorContributor.
func (j *JVM) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(j.root, params), nil
}

// ChangeKind implements serve.KindChanger.
func (j *JVM) ChangeKind(params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	return kindchange.Change(j.root, params)
}

// Components implements serve.ComponentLister.
func (j *JVM) Components(params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	return sbom.List(j.root, params)
}
