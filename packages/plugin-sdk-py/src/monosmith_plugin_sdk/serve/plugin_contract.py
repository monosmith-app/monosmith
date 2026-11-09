"""The shape of a plugin: `describe`, and the methods it implements.

A plugin is any object with a `describe` method and one method for each thing it does. These
protocols document the methods and let a type checker see that `describe` exists; a plugin does
not have to inherit from them.

A method takes the request's params, and may take a second parameter, the `CancelToken` of the
request. The serve loop passes the token only to a method that accepts it, so `def doctor(self,
params)` and `def doctor(self, params, cancel)` are both fine. Methods are plain functions that
may block: the loop runs them on worker threads, as many at a time as the plugin declared in
`concurrency` (one, in order, when it said nothing).
"""

from __future__ import annotations

from typing import Protocol

from ..cancellation import CancelToken
from ..protocol import (
    APISurfaceParams,
    APISurfaceResult,
    ChangeKindParams,
    ChangeKindResult,
    CIParams,
    CIResult,
    ComponentsParams,
    ComponentsResult,
    ContributeParams,
    ContributeResult,
    DeclaredParams,
    DeclaredResult,
    DependencyResult,
    DeriveParams,
    DeriveResult,
    DiscoverParams,
    DiscoverResult,
    DoctorParams,
    DoctorResult,
    GenerateParams,
    GenerateResult,
    InitializeParams,
    InstallParams,
    LatestParams,
    LatestResult,
    LinkParams,
    LinkResult,
    PluginDescription,
    PublishedParams,
    PublishedResult,
    PublishParams,
    PublishResult,
    RemoveParams,
    RemoveResult,
    RewriteParams,
    UninstallParams,
    UnlinkParams,
    VersionParams,
    VersionResult,
)


class Plugin(Protocol):
    """The one method every plugin has."""

    def describe(self, params: InitializeParams) -> PluginDescription:
        """Answer `initialize`: the plugin's identity, kind, languages, manifests and roles.

        The serve loop fills in `protocolVersion` and `capabilities`.
        """
        ...


class Discoverer(Protocol):
    """`discover`: describes the projects behind manifests."""

    def discover(self, params: DiscoverParams, cancel: CancelToken) -> DiscoverResult: ...


class Generator(Protocol):
    """`generate`: scaffolds a project."""

    def generate(self, params: GenerateParams, cancel: CancelToken) -> GenerateResult: ...


class Versioner(Protocol):
    """`version/write`: records a release version in a manifest."""

    def write_version(self, params: VersionParams, cancel: CancelToken) -> VersionResult: ...


class Publisher(Protocol):
    """`publish`: publishes a released project."""

    def publish(self, params: PublishParams, cancel: CancelToken) -> PublishResult: ...


class RegistryReader(Protocol):
    """`published`: the newest version a registry has of a project."""

    def published(self, params: PublishedParams, cancel: CancelToken) -> PublishedResult: ...


class CIContributor(Protocol):
    """`ci`: what a pipeline needs for the plugin's projects."""

    def ci(self, params: CIParams, cancel: CancelToken) -> CIResult: ...


class DoctorContributor(Protocol):
    """`doctor`: checks the plugin's projects."""

    def doctor(self, params: DoctorParams, cancel: CancelToken) -> DoctorResult: ...


class Remover(Protocol):
    """`remove`: cleans up after a project is removed."""

    def remove(self, params: RemoveParams, cancel: CancelToken) -> RemoveResult: ...


class KindChanger(Protocol):
    """`project/kind`: what changing a project between app, lib and package edits."""

    def change_kind(self, params: ChangeKindParams, cancel: CancelToken) -> ChangeKindResult: ...


class ComponentLister(Protocol):
    """`sbom/components`: the components a project ships, for the SBOM a release attaches."""

    def components(self, params: ComponentsParams, cancel: CancelToken) -> ComponentsResult: ...


class Deriver(Protocol):
    """`derive`: the files the plugin derives from facts and the project graph (smith derive)."""

    def derive(self, params: DeriveParams, cancel: CancelToken) -> DeriveResult: ...


class APISurfaceReader(Protocol):
    """`api/surface`: the public API names of a project at a git revision, for the release's version check."""

    def api_surface(self, params: APISurfaceParams, cancel: CancelToken) -> APISurfaceResult: ...


class DependencyManager(Protocol):
    """`dependencies/install` and `dependencies/uninstall`: both, or the capability is not advertised."""

    def install(self, params: InstallParams, cancel: CancelToken) -> DependencyResult: ...

    def uninstall(self, params: UninstallParams, cancel: CancelToken) -> DependencyResult: ...


class Linker(Protocol):
    """`dependencies/link` and `dependencies/unlink`: both, or the capability is not advertised."""

    def link(self, params: LinkParams, cancel: CancelToken) -> LinkResult: ...

    def unlink(self, params: UnlinkParams, cancel: CancelToken) -> DependencyResult: ...


class DeclarationEditor(Protocol):
    """`dependencies/declared` and `dependencies/rewrite`: both, or the capability is not advertised."""

    def declared(self, params: DeclaredParams, cancel: CancelToken) -> DeclaredResult: ...

    def rewrite(self, params: RewriteParams, cancel: CancelToken) -> DependencyResult: ...


class LatestReader(Protocol):
    """`dependencies/latest`: the newest releases of dependencies and pinned tools."""

    def latest(self, params: LatestParams, cancel: CancelToken) -> LatestResult: ...


class Contributor(Protocol):
    """`contribute`: a convention adds tasks to the projects it applies to."""

    def contribute(self, params: ContributeParams, cancel: CancelToken) -> ContributeResult: ...
