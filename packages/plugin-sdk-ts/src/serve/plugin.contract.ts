import type {
  ChangeKindParams, ChangeKindResult, CIParams, CIResult, ComponentsParams, ComponentsResult, ContributeParams, ContributeResult, DeclaredParams, DeclaredResult, DependencyResult, DeriveParams, DeriveResult,
  APISurfaceParams, APISurfaceResult, ChangeKindParams, ChangeKindResult, CIParams, CIResult, ComponentsParams, ComponentsResult, ContributeParams, ContributeResult, DeclaredParams, DeclaredResult, DependencyResult,
  DiscoverParams, DiscoverResult, DoctorParams, DoctorResult, GenerateParams, GenerateResult, InitializeParams,
  InitializeResult, InstallParams, LatestParams, LatestResult, PublishedParams, PublishedResult, PublishParams,
  PublishResult, RemoveParams, RemoveResult, RewriteParams, UninstallParams, VersionParams, VersionResult,
} from '../protocol/index.ts'

/** A result now or later: every method of a plugin may be `async`. */
export type Awaitable<T> = T | Promise<T>

/** What a plugin says about itself. The serve loop fills in `protocolVersion` and `capabilities`. */
export type PluginDescription = Omit<InitializeResult, 'protocolVersion' | 'capabilities'>

/** The one method every plugin has. */
export interface Plugin {
  /** Answers `initialize`: the plugin's identity, kind, languages, manifests and roles. */
  describe: (params: InitializeParams) => Awaitable<PluginDescription>
}

/** `discover`: describes the projects behind manifests. */
export interface Discoverer {
  discover: (params: DiscoverParams, signal: AbortSignal) => Awaitable<DiscoverResult>
}

/** `generate`: scaffolds a project. */
export interface Generator {
  generate: (params: GenerateParams, signal: AbortSignal) => Awaitable<GenerateResult>
}

/** `version/write`: records a release version in a manifest. */
export interface Versioner {
  writeVersion: (params: VersionParams, signal: AbortSignal) => Awaitable<VersionResult>
}

/** `publish`: publishes a released project. */
export interface Publisher {
  publish: (params: PublishParams, signal: AbortSignal) => Awaitable<PublishResult>
}

/** `published`: the newest version a registry has of a project. */
export interface RegistryReader {
  published: (params: PublishedParams, signal: AbortSignal) => Awaitable<PublishedResult>
}

/** `ci`: what a pipeline needs for the plugin's projects. */
export interface CIContributor {
  ci: (params: CIParams, signal: AbortSignal) => Awaitable<CIResult>
}

/** `doctor`: checks the plugin's projects. */
export interface DoctorContributor {
  doctor: (params: DoctorParams, signal: AbortSignal) => Awaitable<DoctorResult>
}

/** `remove`: cleans up after a project is removed. */
export interface Remover {
  remove: (params: RemoveParams, signal: AbortSignal) => Awaitable<RemoveResult>
}

/** `project/kind`: what changing a project between app, lib and package edits. */
export interface KindChanger {
  changeKind: (params: ChangeKindParams, signal: AbortSignal) => Awaitable<ChangeKindResult>
}

/** `sbom/components`: the components a project ships, for the SBOM a release attaches. */
export interface ComponentLister {
  components: (params: ComponentsParams, signal: AbortSignal) => Awaitable<ComponentsResult>
}

/** `derive`: the files the plugin derives from facts and the project graph (smith derive). */
export interface Deriver {
  derive: (params: DeriveParams, signal: AbortSignal) => Awaitable<DeriveResult>
}

/** `api/surface`: the public API names of a project at a git revision, for the release's version check. */
export interface APISurfaceReader {
  apiSurface: (params: APISurfaceParams, signal: AbortSignal) => Awaitable<APISurfaceResult>
}

/** `dependencies/install` and `dependencies/uninstall`: both, or the capability is not advertised. */
export interface DependencyManager {
  install:   (params: InstallParams, signal: AbortSignal) => Awaitable<DependencyResult>
  uninstall: (params: UninstallParams, signal: AbortSignal) => Awaitable<DependencyResult>
}

/** `dependencies/declared` and `dependencies/rewrite`: both, or the capability is not advertised. */
export interface DeclarationEditor {
  declared: (params: DeclaredParams, signal: AbortSignal) => Awaitable<DeclaredResult>
  rewrite:  (params: RewriteParams, signal: AbortSignal) => Awaitable<DependencyResult>
}

/** `dependencies/latest`: the newest releases of dependencies and pinned tools. */
export interface LatestReader {
  latest: (params: LatestParams, signal: AbortSignal) => Awaitable<LatestResult>
}

/** `contribute`: a convention adds tasks to the projects it applies to. */
export interface Contributor {
  contribute: (params: ContributeParams, signal: AbortSignal) => Awaitable<ContributeResult>
}

/**
 * `cancellable: true` says the plugin's methods stop their work when the
 * `AbortSignal` they are passed as their second argument aborts, which the
 * serve loop does when the host sends the `cancel` notification for the
 * request. The plugin then advertises the `cancel` capability. A method that
 * stops throws (any error is answered with the request-cancelled code) or
 * returns the result it already has. Without it the signal never aborts and
 * the host is never sent a cancel: it waits for the answer, or kills the
 * plugin.
 */
export interface Cancellable {
  cancellable: true
}

/**
 * A plugin as {@link serve} takes it: `describe`, and the methods it implements.
 * The capabilities it advertises are derived from which methods it has, and
 * from `cancellable`. How many requests it can work on at once is declared in
 * `describe`'s `concurrency`: omit it for one at a time, in order.
 */
export type SmithPlugin = Plugin & Partial<
Discoverer & Generator & Versioner & Publisher & RegistryReader & CIContributor & DoctorContributor &
Remover & KindChanger & ComponentLister & Deriver & APISurfaceReader & DependencyManager & DeclarationEditor & LatestReader & Contributor & Cancellable
>
