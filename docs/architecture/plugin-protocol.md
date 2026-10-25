# The plugin protocol: roles, handshake, capabilities and contributions

Everything that depends on a language, a tool or a registry in smith is a plugin: a separate process
that answers JSON-RPC requests on its standard input and output. This page is the architecture of
that protocol: how a plugin is started and identified, what it declares in the handshake, which
methods exist and who calls them, how several plugins are merged into one project, and what
happens when something goes wrong. It is a map of `packages/plugin-sdk/protocol` and
`apps/smith/pluginhost`, not a guide to writing a plugin.

Audience: contributors to smith and authors who want to understand why the protocol looks as it
does. To write a plugin, start with [Writing a plugin](../guides/writing-a-plugin.md); for the field
by field reference see [Plugins](../reference/plugins.md). [The plugin system](plugin-system.md)
is the longer, discursive version of the same design.

## Context

```mermaid
flowchart LR
  subgraph smith["smith process"]
    session["session.Open"] --> pluginhost["pluginhost.Start"]
    pluginhost --> hostclient["packages/plugin-sdk/host<br/>Client (one per plugin)"]
    projectgraph["projectgraph.Discover"] --> hostclient
    release["release"] --> hostclient
    cipipeline["cipipeline.Collect"] --> hostclient
    doctor["doctor"] --> hostclient
  end
  hostclient <-->|"stdin/stdout<br/>Content-Length framing"| builtin["built-in plugin<br/>smith plugin serve NAME<br/>(a child smith process)"]
  hostclient <-->|"stdin/stdout"| external["external plugin<br/>plugin.json command"]
  builtin -.->|"packages/plugin-sdk/serve"| sdk["serve.Run loop"]
  external -.->|"any language"| sdk2["its own loop"]
```

The decision to use out-of-process plugins over a pipe is [ADR 0002](decisions/0002-out-of-process-plugins.md);
the Why-not list at the end of [the plugin system](plugin-system.md#why-not-) records the alternatives
that were considered (shared libraries, WASM, gRPC). Where the decision log is silent the reasons
given there are the design notes' own.

## Transport and framing

| Aspect | Value | Source |
|---|---|---|
| Messages | JSON-RPC 2.0 requests, responses and notifications | `protocol/message_contract.go` |
| Framing | `Content-Length: N`, a blank line, N bytes of JSON, as the Language Server Protocol frames | `transport/stream_client.go` |
| Largest message | 64 MiB (`MaxMessageSize = 64 << 20`); a larger header closes the stream | `transport` |
| Channel | the child's stdin and stdout; stderr is forwarded to the user's terminal, so a plugin logs there | `host.Options.Stderr` |
| Error codes | the JSON-RPC ones, plus `-32000` (`CodePluginFailure`) for a failure the user must act on | `protocol` |
| Version | one integer, `protocol.Version = 1`; additive changes use capabilities, a breaking change raises it | `protocol/message_contract.go` |

Calls to one plugin are serialised by the client; calls to different plugins run in parallel.
The default call timeout is two minutes (`host.DefaultTimeout`); `initialize` is given 30 seconds;
the core passes its own timeout on some calls (`ci` and `version/write` one minute), and longer ones
for `generate` and `publish`, which run toolchains. On a timeout the plugin process is
killed, because its stream can no longer be trusted. At the end of a command `shutdown` is sent
and the process is killed if it does not exit within five seconds.

## The schema is the contract

The Go types in `packages/plugin-sdk/protocol` are the source of truth of the protocol, and a JSON
Schema (draft 2020-12) generated from them,
[`packages/plugin-sdk/schema/protocol.schema.json`](../../packages/plugin-sdk/schema/protocol.schema.json),
is the contract for every other language. It has one definition per message (a member is required
when its Go field has no `omitempty`; the doc comments are the descriptions), and next to the
definitions `x-methods` (each of the 16 methods with its params, its result and the capability it needs,
`null` for `initialize` and `shutdown`), `x-errorCodes`, `x-roles` (with each role's cardinality and the
kinds that default to it), `x-protocolVersion` and `x-transport` (the framing and the largest message).

| Step | Where | Fails when |
|---|---|---|
| reflect over the types and read the package's source for comments and enumerations | `packages/plugin-sdk/protocolschema`, run by `go run ./schemagen` (`smith run schema -p plugin-sdk`) | the protocol has a method, an error code or a capability that `protocolschema`'s tables do not describe |
| compare the committed file with what the generator writes | the Go test `TestCommittedSchemaIsCurrent` | someone changed `protocol` and not the schema |
| generate each SDK's types from the schema | the TypeScript SDK: `tools/generate-protocol` (`smith run generate -p plugin-sdk-ts`) | its test finds the generated files stale; another test finds a method the SDK cannot serve |
| check that the SDKs agree on the wire | `packages/plugin-sdk-ts/integration`: a TypeScript and a Go plugin with the same logic answer the same requests with the same bytes | an SDK writes members in another order, leaves out different zero values, or escapes differently |

Three things about the wire that the schema cannot say, and that a new SDK must do (the Go SDK does
them and the host relies on them):

- A required array or object that is empty in Go is written as `null` when the Go value is nil, so a
  receiver treats `null` as empty.
- Optional members follow `omitempty`: a zero value (`""`, `0`, `false`, an empty array or map) and a
  nil pointer are left out, a struct is not, and a required member is always written, `false` included.
- Maps are written with their keys sorted, and Go escapes `<`, `>` and `&` in strings; both are only
  needed to make the bytes of two SDKs equal.

## Who is started, and how

`pluginhost.Start` (`apps/smith/pluginhost/start_plugins_use_case.go`) builds a list of launches
and starts them in parallel, each with `SMITH_PLUGIN=1` in its environment and the workspace root
in `initialize`:

| Plugin | Command | Notes |
|---|---|---|
| a built-in language plugin in `Builtins` (`node`, `go`, `python`, `rust`, `dotnet`, `ruby`, `php`, `cpp`, `dart`, `jvm`, `bicep`, `recipe`) | `smith plugin serve <name>`, i.e. the running binary again | all of them unless `[plugins] enabled` lists some; skipped when `[plugins.external.<name>]` replaces it |
| a built-in addon in `BuiltinAddons` (`vertical-slices`, `eslint`, `jest`, `vitest`, `playwright`, `typescript`, `vite`, `angular`, `astro`, `vscode`, `go-smith-plugin`, `docker`) | the same | always started unless replaced |
| an external plugin, `[plugins.external.<name>]` | the `command` of the `plugin.json` found on the search path, or the command in `smith.toml` | the name it answers to and its version are checked (`checkExternal`) |

The search path (`pluginhost.SearchPath`) is the workspace's `.smith/plugins`, each directory of
`SMITH_PLUGIN_PATH`, then `$SMITH_HOME/plugins` (`SMITH_HOME` defaults to `~/.smith`); the first
directory holding `<name>/plugin.json` wins. `plugin.json` has `name`, `version`, `description` and
`command`; `./` paths are relative to the plugin directory and `{os}`, `{arch}`, `{exe}` are
filled in (`plugin_manifest_contract.go`). A plugin that fails to start stops the command, naming it.

Built-in plugins use the same protocol on purpose: they are its first users, and an external plugin
can replace one by name without the core noticing. How plugins are published, found and installed
by name is not decided; see the open list in the [decision records](decisions/README.md#open-not-decided-yet).

## The handshake

The host sends `initialize` first (`InitializeParams`): the protocol versions it speaks, its name
and version, and the absolute workspace root. The plugin answers `InitializeResult`:

| Field | Meaning |
|---|---|
| `protocolVersion` | the version the plugin picked from the host's list; one the host does not speak ends the connection |
| `plugin` | name and version; the version enters every cache key of the projects it shapes |
| `kind` | `language` or `convention`. It only sets the default roles |
| `capabilities` | the optional methods the plugin implements (below) |
| `languages` | values of `--lang` that reach this plugin |
| `manifests` | file names or patterns (`package.json`, `*.csproj`) that may make a directory a project of this plugin |
| `roles` | the levels the plugin can be chosen for; empty means `DefaultRoles(kind)` |
| `provides` | slots: task names the plugin owns (`test`, `lint`) |
| `requires`, `conflicts` | plugins that must or must not share a project |
| `stacks` | ready-made stacks the plugin publishes |

### Capabilities

The host calls a method only if the plugin advertised it; an unadvertised method means "nothing to
say here", not an error. The Go SDK derives the flags from the interfaces a plugin implements
(`packages/plugin-sdk/serve/plugin_contract.go`), and the TypeScript SDK from the members of the plugin
object. A plugin without an SDK sets the booleans by hand. The schema's methods table (below) says
which capability each method needs.

| Capability flag | Methods | Go interface | Called by |
|---|---|---|---|
| `discover` | `discover` | `Discoverer` | `projectgraph.Discover`, every session |
| `contribute` | `contribute` | `Contributor` | `projectgraph`, for stacks and conventions |
| `generate` | `generate` | `Generator` | `smith new`, `smith add` |
| `remove` | `remove` | `Remover` | `smith remove` |
| `version` | `version/write` | `Versioner` | `release` |
| `publish` | `publish` | `Publisher` | `release` (registry or release archives; deploy plugins too) |
| `published` | `published` | `RegistryReader` | release lineage checks, doctor |
| `ci` | `ci` | `CIContributor` | `cipipeline.Collect`, release preflight |
| `doctor` | `doctor` | `DoctorContributor` | `smith doctor` |
| `dependencies` | `dependencies/install`, `dependencies/uninstall` | `DependencyManager` | `smith install`, `uninstall` |
| `declarations` | `dependencies/declared`, `dependencies/rewrite` | `DeclarationEditor` | `smith sync`, `smith up --align` |
| `latest` | `dependencies/latest` | `LatestReader` | `smith up` |

`initialize` and `shutdown` are always available.

## Roles, levels and slots

A role is a level at which a plugin can take part in a project (`protocol/plugin_role_enum.go`):

| Role | Question it answers | Cardinality in a project |
|---|---|---|
| `language` | what is here, what does it import, how is it versioned and published | exactly one |
| `scaffold` | what files does a new project start with | at most one |
| `toolchain` | how are lint, test, build and typecheck run | one plugin per slot (task name) |
| `structure` | how must the source be laid out | any number, additive |
| `deploy` | what is built and shipped from it | any number, additive |

`pluginhost.ValidateMembers` (`stack_validator.go`) checks every stack and recipe: each plugin plays
only a role it declared and applies to the project's language, two toolchains may not provide the
same slot, every `requires` is met, no two plugins `conflict`. A stack is a named list of plugins a
project uses; a recipe is data that chooses stacks by answers ([ADR 0021](decisions/0021-plugin-roles-and-per-project-stacks.md),
[ADR 0023](decisions/0023-recipes-with-levels-and-slots.md)).

## The methods at a glance

All payload types live in `packages/plugin-sdk/protocol`.

| Method | Request | Response | Contribution to the whole |
|---|---|---|---|
| `discover` | `DiscoverParams{manifests}` (paths whose base name matches) | `DiscoverResult{projects}` | one `Project` per manifest the plugin owns: name, root, language, kind, exports, imports, version, tasks, tags, stack, settings |
| `contribute` | `ContributeParams{projects}` | `ContributeResult{projects: name -> Contribution}` | `Contribution{tasks, replace, dependsOn}`: add tasks, replace tasks, append dependencies |
| `generate` | `GenerateParams{kind, name, directory, workspace, stack, settings}` | `GenerateResult{files, notes}` | writes a new project's files |
| `remove` | `RemoveParams` | `RemoveResult` | cleans traces outside the project directory |
| `version/write` | `VersionParams{project, version, dependencies, dryRun}` | `VersionResult{files}` | stamps the manifests; the release restores them |
| `publish` | `PublishParams{project, version, dryRun, assets}` | `PublishResult{published, skipped, registry, assets, assetPlatforms, binary, images}` | uploads; for an app builds release archives (`AssetSpec`); for a deploy plugin reports pushed images |
| `published` | `PublishedParams{project}` | `PublishedResult{version, registry, unreachable, detail}` | what the registry already has |
| `ci` | `CIParams{projects, audit}` | `CIResult{toolchains, authenticate, setup, audit, postRelease, credentials, permissions, stores}` | see [CI generation](ci-generation.md) |
| `doctor` | `DoctorParams{projects}` | `DoctorResult{findings}` | `Finding{check, ok, detail, remedy, warning, fix}`; a `Fix` has `files`, `steps`, `choices`, `safe` |
| `dependencies/install`, `uninstall` | `InstallParams`, `UninstallParams` | `DependencyResult` | adds or removes packages in one project |
| `dependencies/declared`, `rewrite` | `DeclaredParams`, `RewriteParams` | `DeclaredResult`, `DependencyResult` | reads and edits declared specs |
| `dependencies/latest` | `LatestParams` | `LatestResult{latest, unreachable, tools}` (`tools` are `ToolPin`s) | newest releases and pinned tools |

### Tasks

A `Task` (`project_contract.go`) is a command plus what the cache needs: `command` (argv, no shell),
`cwd`, `env`, `inputs`, `outputs`, `dependsOn`, `cache`, `description`, and two flags for the
runner: `verify` (run it in `smith ci verify` besides lint, typecheck, test and build) and `serve`,
`serveUrl`, `serveTimeout` (start an app for an end-to-end task). Plugins never talk to the cache;
they have to be honest and deterministic in what they put in a task. The key is described in
[the cache page](cache.md).

### Doctor fixes

A fix is data, not code: whole new file contents (`FileEdit`, so the host can show a diff), ordered
`steps` (argv, no shell) and `choices` (questions whose options are fixes). The host applies it, which
is why a plugin in any language only describes the change
([ADR 0033](decisions/0033-doctor-fixes-with-steps-and-choices.md)).

## Assembling a project

```mermaid
sequenceDiagram
  participant S as session
  participant H as pluginhost
  participant L as language plugin
  participant G as projectgraph
  participant T as toolchain, structure, deploy plugins
  participant C as convention plugins
  participant W as smith.toml
  S->>H: Start(config)
  H->>L: initialize
  L-->>H: kind, capabilities, manifests, roles
  H->>T: initialize
  H->>C: initialize
  S->>G: Discover(files)
  G->>L: discover(matching manifests)
  L-->>G: Project{exports, imports, tasks}
  G->>G: edges: import equals or is under an export
  G->>T: contribute(projects whose stack names you)
  T-->>G: Contribution{tasks, replace, dependsOn}
  G->>C: contribute(projects you apply to)
  C-->>G: Contribution
  W-->>G: ApplyOverrides (task fields, tags, kinds)
  G->>G: cycles, order, selection
```

The merge rule for a task field is: language plugin, then the stack's plugins in order, then
conventions, then `[tasks.<name>]`, then `[projects.<p>.tasks.<name>]`; later wins, and each merge
records its source for `smith explain`. The code is `projectgraph.Discover` and
`projectgraph.ApplyOverrides` (`task_overrides_mapper.go`). Plugins never call or read each other:
what they share goes through the project (`exports`, `imports`, `tags`, `settings`) and through task
names. The result therefore depends on the set of plugins and the stack's order, not on timing.
How the edges are found is the subject of [the project graph page](project-graph.md).

## Failure modes

| What goes wrong | What the host does |
|---|---|
| plugin missing or failing to start | the command stops, naming the plugin and its command |
| a protocol version the host does not speak | the connection is refused |
| a call exceeds its timeout | the process is killed and the call fails; other plugins are unaffected |
| the plugin returns an error | the command stops with the plugin's name and message (`-32000` means act on it) |
| two plugins provide one slot, or conflict | the stack or recipe is refused, naming both |
| an external plugin's name or version differs from `smith.toml` | refused at start-up, with both versions |
| stray output on stdout | the stream is corrupt and the call fails; log to stderr |

## The conformance suite

`packages/plugin-sdk/conformance.Check` starts a plugin as the host would and verifies that it
identifies itself, that its roles match what it implements, that its stacks are complete, that
every advertised method answers a well-formed request, and, with generation enabled, that the
projects it generates are found again by its own `discover`. Authors call it from their tests;
`smith plugin check <name|path>` runs it against any plugin ([`smith plugin check`](../manual/plugin-check.md)).
The six checks and their output are shown in that page.

## Where in the code

| Concern | Path |
|---|---|
| Message types, methods, roles, contracts | `packages/plugin-sdk/protocol` |
| Framing | `packages/plugin-sdk/transport` |
| Host client (spawn, call, timeout, kill) | `packages/plugin-sdk/host` |
| Plugin-side loop and interfaces | `packages/plugin-sdk/serve` |
| Conformance | `packages/plugin-sdk/conformance` |
| The JSON Schema, and what writes it | `packages/plugin-sdk/schema`, `protocolschema`, `schemagen` |
| The TypeScript SDK | `packages/plugin-sdk-ts` |
| Starting, search path, `plugin.json`, stacks | `apps/smith/pluginhost` |
| Built-in list | `apps/smith/pluginhost/builtin_plugins_config.go`; plugins in `libs/plugin-*` |
| Assembling projects | `apps/smith/projectgraph` |

## Decisions

| Decision | Record |
|---|---|
| Plugins are processes speaking JSON-RPC over stdio | [ADR 0002](decisions/0002-out-of-process-plugins.md) |
| Roles and per-project stacks | [ADR 0021](decisions/0021-plugin-roles-and-per-project-stacks.md) |
| Recipes with levels and slots | [ADR 0023](decisions/0023-recipes-with-levels-and-slots.md) |
| Doctor fixes with steps and choices | [ADR 0033](decisions/0033-doctor-fixes-with-steps-and-choices.md) |
| Graph edges from shared schemas | [ADR 0028](decisions/0028-graph-edges-from-shared-schemas.md) |

No record explains the choice of Content-Length framing over newline-delimited JSON; the design
note gives its reasons (no escaping, large payloads) and the decision log does not cite them.
