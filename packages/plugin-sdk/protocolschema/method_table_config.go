package protocolschema

import (
	"reflect"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// MethodSpec describes one method of the protocol: what a request carries,
// what the answer carries and which capability a plugin advertises to be
// asked it.
type MethodSpec struct {
	// Method is the method name on the wire.
	Method      string
	Description string
	Params      reflect.Type
	Result      reflect.Type
	// Capability is the name of the field of protocol.Capabilities that must
	// be true for the host to call the method. Empty means the method is
	// always available.
	Capability string
	// Notification marks a method the host sends without an id: the plugin
	// answers nothing, so the row has no result.
	Notification bool
}

// noMembers is the params and result of shutdown.
var noMembers = reflect.TypeFor[struct{}]()

// Methods is the methods table of protocol version 1. Build checks it against
// the Method constants and the Capabilities fields of the protocol package, so
// a method added there fails the build until it is described here.
var Methods = []MethodSpec{
	{
		Method:      protocol.MethodInitialize,
		Description: "The first request, sent once by the host. The plugin picks a protocol version from the host's list and answers who it is and what it can do.",
		Params:      reflect.TypeFor[protocol.InitializeParams](), Result: reflect.TypeFor[protocol.InitializeResult](),
	},
	{
		Method:      protocol.MethodShutdown,
		Description: "The last request. The plugin answers an empty object and exits.",
		Params:      noMembers, Result: noMembers,
	},
	{
		Method:      protocol.MethodCancel,
		Description: "A notification, so it gets no response: the host asks the plugin to stop the in-flight request with the given id. The plugin then answers that request, with its result if the work was already done or with the request-cancelled error. An id that is unknown or already answered is ignored. The host sends it only to a plugin that advertised the capability.",
		Params:      reflect.TypeFor[protocol.CancelParams](),
		Capability:  "Cancel", Notification: true,
	},
	{
		Method:      protocol.MethodDiscover,
		Description: "Describes the projects behind manifests the host found; a language plugin's way into the workspace.",
		Params:      reflect.TypeFor[protocol.DiscoverParams](), Result: reflect.TypeFor[protocol.DiscoverResult](),
		Capability: "Discover",
	},
	{
		Method:      protocol.MethodGenerate,
		Description: "Writes the starter files of a new project.",
		Params:      reflect.TypeFor[protocol.GenerateParams](), Result: reflect.TypeFor[protocol.GenerateResult](),
		Capability: "Generate",
	},
	{
		Method:      protocol.MethodVersion,
		Description: "Records a release version in a project's manifest.",
		Params:      reflect.TypeFor[protocol.VersionParams](), Result: reflect.TypeFor[protocol.VersionResult](),
		Capability: "Version",
	},
	{
		Method:      protocol.MethodPublish,
		Description: "Publishes one project at one version, or builds the files attached to a release.",
		Params:      reflect.TypeFor[protocol.PublishParams](), Result: reflect.TypeFor[protocol.PublishResult](),
		Capability: "Publish",
	},
	{
		Method:      protocol.MethodCI,
		Description: "Describes what a generated pipeline needs for the plugin's projects.",
		Params:      reflect.TypeFor[protocol.CIParams](), Result: reflect.TypeFor[protocol.CIResult](),
		Capability: "CI",
	},
	{
		Method:      protocol.MethodDoctor,
		Description: "Checks the plugin's projects and reports findings, each with its remedy and, optionally, a fix.",
		Params:      reflect.TypeFor[protocol.DoctorParams](), Result: reflect.TypeFor[protocol.DoctorResult](),
		Capability: "Doctor",
	},
	{
		Method:      protocol.MethodContribute,
		Description: "Asks a convention (toolchain, structure, deploy) for the tasks and dependencies it adds to projects other plugins discovered.",
		Params:      reflect.TypeFor[protocol.ContributeParams](), Result: reflect.TypeFor[protocol.ContributeResult](),
		Capability: "Contribute",
	},
	{
		Method:      protocol.MethodRemove,
		Description: "Lets a plugin clean up what it owns outside a project's directory after the project is removed.",
		Params:      reflect.TypeFor[protocol.RemoveParams](), Result: reflect.TypeFor[protocol.RemoveResult](),
		Capability: "Remove",
	},
	{
		Method:      protocol.MethodChangeKind,
		Description: "Asks what changing a project between app, lib and package edits: its manifest, the workspace files and the dependents that name its kind or directory. The host applies the edits and moves the directory.",
		Params:      reflect.TypeFor[protocol.ChangeKindParams](), Result: reflect.TypeFor[protocol.ChangeKindResult](),
		Capability: "ChangeKind",
	},
	{
		Method:      protocol.MethodPublished,
		Description: "Asks which version of a project its registry has, so a release can check its tags against it.",
		Params:      reflect.TypeFor[protocol.PublishedParams](), Result: reflect.TypeFor[protocol.PublishedResult](),
		Capability: "Published",
	},
	{
		Method:      protocol.MethodInstall,
		Description: "Adds packages to a project with the ecosystem's own toolchain, or, with no project, installs every dependency of the projects the plugin serves.",
		Params:      reflect.TypeFor[protocol.InstallParams](), Result: reflect.TypeFor[protocol.DependencyResult](),
		Capability: "Dependencies",
	},
	{
		Method:      protocol.MethodUninstall,
		Description: "Removes packages from a project.",
		Params:      reflect.TypeFor[protocol.UninstallParams](), Result: reflect.TypeFor[protocol.DependencyResult](),
		Capability: "Dependencies",
	},
	{
		Method:      protocol.MethodDeclared,
		Description: "Lists the dependencies the manifests of the plugin's projects declare.",
		Params:      reflect.TypeFor[protocol.DeclaredParams](), Result: reflect.TypeFor[protocol.DeclaredResult](),
		Capability: "Declarations",
	},
	{
		Method:      protocol.MethodRewrite,
		Description: "Changes declared dependency specs in place.",
		Params:      reflect.TypeFor[protocol.RewriteParams](), Result: reflect.TypeFor[protocol.DependencyResult](),
		Capability: "Declarations",
	},
	{
		Method:      protocol.MethodLatest,
		Description: "Asks registries for the newest releases of dependencies, and how the tools the plugin pins compare with theirs.",
		Params:      reflect.TypeFor[protocol.LatestParams](), Result: reflect.TypeFor[protocol.LatestResult](),
		Capability: "Latest",
	},
	{
		Method:      protocol.MethodComponents,
		Description: "Asks which components a project ships (its locked dependencies, read from a lockfile where the ecosystem has one and from the manifest otherwise), as package URLs, for the SBOM a release attaches.",
		Params:      reflect.TypeFor[protocol.ComponentsParams](), Result: reflect.TypeFor[protocol.ComponentsResult](),
		Capability: "Components",
	},
	{
		Method:      protocol.MethodAPISurface,
		Description: "Asks for the public API names of a project, at the working tree or at a git revision (its last release tag), so a release can compare them and say what bump the change needs.",
		Params:      reflect.TypeFor[protocol.APISurfaceParams](), Result: reflect.TypeFor[protocol.APISurfaceResult](),
		Capability: "APISurface",
	},
}
