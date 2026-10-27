// The interfaces a plugin implements: one for each capability of the protocol. A plugin advertises a
// capability by implementing its interface, and the serve loop derives `capabilities` from that.
//
// A method added to the protocol that is not listed here fails the generator, so the SDK cannot fall
// behind the schema unnoticed.

export const PLUGIN_INTERFACES = [
  { name: 'IDiscoverer', capability: 'discover', summary: 'A plugin that finds projects from manifests.', methods: { discover: 'DiscoverAsync' } },
  { name: 'IGenerator', capability: 'generate', summary: 'A plugin that scaffolds projects.', methods: { generate: 'GenerateAsync' } },
  { name: 'IVersioner', capability: 'version', summary: 'A plugin that records a release version in a manifest.', methods: { 'version/write': 'WriteVersionAsync' } },
  { name: 'IPublisher', capability: 'publish', summary: 'A plugin that publishes a released project.', methods: { publish: 'PublishAsync' } },
  { name: 'IRegistryReader', capability: 'published', summary: 'A plugin that says which version a registry has of a project.', methods: { published: 'PublishedAsync' } },
  { name: 'ICiContributor', capability: 'ci', summary: 'A plugin that says what a pipeline needs for its projects.', methods: { ci: 'CiAsync' } },
  { name: 'IDoctorContributor', capability: 'doctor', summary: 'A plugin that checks its projects.', methods: { doctor: 'DoctorAsync' } },
  { name: 'IRemover', capability: 'remove', summary: 'A plugin that cleans up after a project is removed.', methods: { remove: 'RemoveAsync' } },
  { name: 'IKindChanger', capability: 'changeKind', summary: 'A plugin that says what changing a project between app, lib and package edits.', methods: { 'project/kind': 'ChangeKindAsync' } },
  {
    name: 'IDependencyManager',
    capability: 'dependencies',
    summary: 'A plugin that installs and uninstalls packages in its projects. Both methods, or the capability is not advertised.',
    methods: { 'dependencies/install': 'InstallAsync', 'dependencies/uninstall': 'UninstallAsync' },
  },
  {
    name: 'IDeclarationEditor',
    capability: 'declarations',
    summary: 'A plugin that lists and rewrites the dependency specs its projects declare. Both methods, or the capability is not advertised.',
    methods: { 'dependencies/declared': 'DeclaredAsync', 'dependencies/rewrite': 'RewriteAsync' },
  },
  { name: 'ILatestReader', capability: 'latest', summary: 'A plugin that reports the newest releases of dependencies and pinned tools.', methods: { 'dependencies/latest': 'LatestAsync' } },
  { name: 'IContributor', capability: 'contribute', summary: 'A convention that adds tasks to the projects it applies to.', methods: { contribute: 'ContributeAsync' } },
]

/** The methods the serve loop answers itself. */
export const SERVE_LOOP_METHODS = ['initialize', 'shutdown', 'cancel']
