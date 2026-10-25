import type {
  CIParams, CIResult, ContributeParams, ContributeResult, DeclaredParams, DeclaredResult, DependencyResult,
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
  discover: (params: DiscoverParams) => Awaitable<DiscoverResult>
}

/** `generate`: scaffolds a project. */
export interface Generator {
  generate: (params: GenerateParams) => Awaitable<GenerateResult>
}

/** `version/write`: records a release version in a manifest. */
export interface Versioner {
  writeVersion: (params: VersionParams) => Awaitable<VersionResult>
}

/** `publish`: publishes a released project. */
export interface Publisher {
  publish: (params: PublishParams) => Awaitable<PublishResult>
}

/** `published`: the newest version a registry has of a project. */
export interface RegistryReader {
  published: (params: PublishedParams) => Awaitable<PublishedResult>
}

/** `ci`: what a pipeline needs for the plugin's projects. */
export interface CIContributor {
  ci: (params: CIParams) => Awaitable<CIResult>
}

/** `doctor`: checks the plugin's projects. */
export interface DoctorContributor {
  doctor: (params: DoctorParams) => Awaitable<DoctorResult>
}

/** `remove`: cleans up after a project is removed. */
export interface Remover {
  remove: (params: RemoveParams) => Awaitable<RemoveResult>
}

/** `dependencies/install` and `dependencies/uninstall`: both, or the capability is not advertised. */
export interface DependencyManager {
  install:   (params: InstallParams) => Awaitable<DependencyResult>
  uninstall: (params: UninstallParams) => Awaitable<DependencyResult>
}

/** `dependencies/declared` and `dependencies/rewrite`: both, or the capability is not advertised. */
export interface DeclarationEditor {
  declared: (params: DeclaredParams) => Awaitable<DeclaredResult>
  rewrite:  (params: RewriteParams) => Awaitable<DependencyResult>
}

/** `dependencies/latest`: the newest releases of dependencies and pinned tools. */
export interface LatestReader {
  latest: (params: LatestParams) => Awaitable<LatestResult>
}

/** `contribute`: a convention adds tasks to the projects it applies to. */
export interface Contributor {
  contribute: (params: ContributeParams) => Awaitable<ContributeResult>
}

/**
 * A plugin as {@link serve} takes it: `describe`, and the methods it implements.
 * The capabilities it advertises are derived from which methods it has.
 */
export type SmithPlugin = Plugin & Partial<
Discoverer & Generator & Versioner & Publisher & RegistryReader & CIContributor & DoctorContributor &
Remover & DependencyManager & DeclarationEditor & LatestReader & Contributor
>
