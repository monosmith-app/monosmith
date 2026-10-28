# The plugin protocol cheat sheet

Every method a plugin can implement, every message and every field, in tables. The tables of
messages are generated from the Go types in `packages/plugin-sdk/protocol` (the field names are the
JSON names, the descriptions are the doc comments of the fields), and the method tables are
checked against `method_enum.go`, `initialize_contract.go` and `serve/plugin_contract.go`, so they
follow the protocol. Version 1 of the protocol is described; it was checked against the SDK in the
smith 1.62.0 source tree. The same information, machine-readable, is the JSON Schema
[`packages/plugin-sdk/schema/protocol.schema.json`](../../../packages/plugin-sdk/schema/protocol.schema.json),
which the [TypeScript SDK](typescript-sdk.md), the [Python SDK](python-sdk.md) and the [C# SDK](csharp-sdk.md) generate their types from.

For the reasoning behind the protocol read [the plugin protocol](../../architecture/plugin-protocol.md);
for a guide, the [quickstart](quickstart.md).

## Transport

| Aspect | Value |
|---|---|
| Messages | JSON-RPC 2.0 requests, responses and notifications |
| Channel | the plugin's standard input (requests) and standard output (responses); standard error is shown to the user |
| Framing | `Content-Length: <bytes>`, `\r\n\r\n`, then that many bytes of UTF-8 JSON. The length counts bytes. As the Language Server Protocol frames |
| Largest message | 64 MiB |
| Order | the host sends `initialize` first; a request with an `id` gets exactly one response with the same `id`; a request without an `id` is a notification and gets none |
| Requests in flight | at most the plugin's `concurrency` (absent: 1, so the host sends the next request after the previous is answered); responses may come in any order, matched by `id` |
| Cancel | the notification `cancel`, `{"id": <in-flight id>}`, sent only to a plugin whose `capabilities.cancel` is true; the plugin answers that request with its result or `-32800` |
| Environment | the host sets `SMITH_PLUGIN=1` in the plugin's environment |
| Paths | workspace-relative with forward slashes; only `workspaceRoot` is absolute, with the host's separators |
| End | the host sends `shutdown`; the plugin answers `{}` and exits. Closing stdin also ends it |

A frame on the wire:

```text
Content-Length: 85

{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersions":[1],...}}
```

### Error codes

| Code | Name in the SDK | Meaning |
|---|---|---|
| `-32700` | `CodeParseError` | the message was not valid JSON |
| `-32600` | `CodeInvalidRequest` | invalid request, for example `initialize` was not first |
| `-32601` | `CodeMethodNotFound` | `method not supported: <method>` |
| `-32602` | `CodeInvalidParams` | the params do not have the expected shape |
| `-32603` | `CodeInternalError` | internal error |
| `-32000` | `CodePluginFailure` | the method ran and failed for a reason the user must act on; put the remedy in `message` and smith shows it. The Go SDK returns this for every error a method returns, except one of a request the host cancelled |
| `-32800` | `CodeRequestCancelled` | the plugin stopped the request because the host sent `cancel` for its id. A plugin that had finished answers with the result instead |

### A real conversation

This is the actual traffic with the Python example `docs/examples/plugins/python/plugin.py`
(bodies only; each is framed as above). `>>` is smith to the plugin, `<<` the answer.

```json
>> {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersions": [1], "host": {"name": "smith", "version": "0.0.0"}, "workspaceRoot": "C:\work\demo"}}
<< {"jsonrpc": "2.0", "id": 1, "result": {"protocolVersion": 1, "plugin": {"name": "widget", "version": "0.1.0"}, "kind": "language", "roles": ["language"], "languages": ["widget"], "manifests": ["widget.json"], "capabilities": {"discover": true}}}

>> {"jsonrpc": "2.0", "id": 2, "method": "discover", "params": {"manifests": ["libs/hello/widget.json"]}}
<< {"jsonrpc": "2.0", "id": 2, "result": {"projects": [{"name": "hello", "root": "libs/hello", "language": "widget", "kind": "lib", "version": "1.0.0", "exports": ["hello"], "tasks": {"hello": {"command": ["git", "--version"], "inputs": ["widget.json"], "cache": true, "description": "say hello, as a widget does"}}}]}}

>> {"jsonrpc": "2.0", "id": 3, "method": "nope", "params": {}}
<< {"jsonrpc": "2.0", "id": 3, "error": {"code": -32601, "message": "method not supported: nope"}}

>> {"jsonrpc": "2.0", "id": 4, "method": "shutdown", "params": {}}
<< {"jsonrpc": "2.0", "id": 4, "result": {}}
```

### A cancelled request

The traffic with the example `docs/examples/plugins/structure-scan-go`, which advertises `cancel` and declares a
`concurrency` of 4 (`WIDGET_SCAN_DELAY_MS=100` makes its scan slow enough to cancel). The notification has no
`id` and gets no response; the request it names is answered with the cancelled error.

```json
>> {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersions":[1],"host":{"name":"smith","version":"1.62.0"},"workspaceRoot":"/ws"}}
<< {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":1,"plugin":{"name":"widget-scan","version":"0.1.0"},"kind":"convention","capabilities":{"doctor":true,"cancel":true},"concurrency":4,"languages":["widget"],"roles":["structure"]}}

>> {"jsonrpc":"2.0","id":2,"method":"doctor","params":{"projects":[{"name":"core","root":"libs/core","language":"widget","kind":"lib"}]}}
>> {"jsonrpc":"2.0","method":"cancel","params":{"id":2}}
<< {"jsonrpc":"2.0","id":2,"error":{"code":-32800,"message":"request cancelled: the scan was stopped"}}
```

## Methods

The host never calls a method the plugin did not advertise in `capabilities`. `initialize` and
`shutdown` are always available.

| Method | Capability | Go interface (`serve`) | Params → result | Called by | Host timeout |
|---|---|---|---|---|---|
| `initialize` | always | `Plugin.Describe` | `InitializeParams` → `InitializeResult` | every command that loads plugins | 30 s |
| `shutdown` | always | built in | `{}` → `{}` | the end of the command | 5 s |
| `cancel` (a notification) | `cancel` | `Cancellable` | `CancelParams` → no response | a call whose context ends or whose time runs out | none: the plugin has 5 s to answer the request |
| `discover` | `discover` | `Discoverer` | `DiscoverParams` → `DiscoverResult` | every command that reads the workspace | 1 min |
| `contribute` | `contribute` | `Contributor` | `ContributeParams` → `ContributeResult` | the same, for stack and convention plugins | 1 min |
| `generate` | `generate` | `Generator` | `GenerateParams` → `GenerateResult` | `smith add` | 20 min |
| `version/write` | `version` | `Versioner` | `VersionParams` → `VersionResult` | `smith release`, native builds | 1 min |
| `publish` | `publish` | `Publisher` | `PublishParams` → `PublishResult` | `smith release`, `smith ci release`, native builds | 30 min |
| `published` | `published` | `RegistryReader` | `PublishedParams` → `PublishedResult` | `smith release`, `smith doctor` | 1 min |
| `ci` | `ci` | `CIContributor` | `CIParams` → `CIResult` | `smith ci ...`, pipeline generation, release preflight | 1 min |
| `doctor` | `doctor` | `DoctorContributor` | `DoctorParams` → `DoctorResult` | `smith doctor` | 1 min |
| `remove` | `remove` | `Remover` | `RemoveParams` → `RemoveResult` | `smith remove` | 20 min |
| `project/kind` | `changeKind` | `KindChanger` | `ChangeKindParams` → `ChangeKindResult` | `smith kind` | 5 min |
| `dependencies/install` | `dependencies` | `DependencyManager.Install` | `InstallParams` → `DependencyResult` | `smith install`, `smith sync` | 20 min |
| `dependencies/uninstall` | `dependencies` | `DependencyManager.Uninstall` | `UninstallParams` → `DependencyResult` | `smith uninstall` | 20 min |
| `dependencies/declared` | `declarations` | `DeclarationEditor.Declared` | `DeclaredParams` → `DeclaredResult` | `smith sync`, `smith up`, doctor | 20 min |
| `dependencies/rewrite` | `declarations` | `DeclarationEditor.Rewrite` | `RewriteParams` → `DependencyResult` | `smith sync`, `smith up` | 20 min |
| `dependencies/latest` | `latest` | `LatestReader` | `LatestParams` → `LatestResult` | `smith up` | 20 min |

Calls made without a timeout get 2 minutes. When a call runs out of time the host sends `cancel` to a plugin that
advertised it, waits 5 seconds for the answer, and kills the plugin if none comes; a plugin that cannot be
asked is killed at once. When the command is cancelled (Ctrl+C) the host does the same with its 5 seconds for
every plugin. The timeouts are the ones the host passes today (`apps/smith`); they are not part of the protocol.

The host never has more requests in flight to a plugin than its `concurrency` allows, and the plugin may answer
them in any order.

## Capabilities

The `capabilities` object of the `initialize` result: every key is optional and defaults to false.

| Key | Enables | Role it usually goes with |
|---|---|---|
| `discover` | `discover` | language |
| `generate` | `generate` | language, scaffold, toolchain, structure, deploy |
| `version` | `version/write` | language |
| `publish` | `publish` | language, deploy |
| `ci` | `ci` | any |
| `doctor` | `doctor` | any |
| `contribute` | `contribute` | toolchain, structure, deploy |
| `remove` | `remove` | language |
| `changeKind` | `project/kind` | language |
| `published` | `published` | language |
| `dependencies` | `dependencies/install`, `dependencies/uninstall` | language |
| `declarations` | `dependencies/declared`, `dependencies/rewrite` | language |
| `latest` | `dependencies/latest` | language |
| `cancel` | the `cancel` notification: the plugin stops the work of a request when the host cancels it | any plugin with a method that can run for long |

With the Go SDK the capabilities are derived from the interfaces the plugin type implements
(`serve.CapabilitiesOf`), so a plugin cannot advertise a method it lacks.

## Enumerations

| Name | Values |
|---|---|
| Role (`roles`, `PluginRef.level`) | `language`, `toolchain`, `scaffold`, `structure`, `deploy` |
| Plugin kind (`kind`) | `language`, `convention` |
| Project kind (`kind` of a project) | `app` (deployed, never published), `lib` (internal, never published), `package` (published) |
| Default roles | a `language` plugin: `language`, `toolchain`, `scaffold`; a `convention` plugin: `structure` |
| Registry sign-in (`Registry.auth`) | `azure-artifacts`, `github-packages`, `token` |
| `NoStack` | `none`: the stack value that means "the language plugin alone" |

## Messages

Every table lists the JSON field, the Go type, whether the field is required (a field with
`omitempty` is optional), and the description from the SDK. `Project`, `Task` and `Stack` appear
inside many messages.

### InitializeParams

InitializeParams is sent once, first, by the host.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `protocolVersions` | `[]int` | yes | ProtocolVersions lists every major version the host can speak. |
| `host` | `Peer` | yes |  |
| `workspaceRoot` | `string` | yes | WorkspaceRoot is an absolute path using the host OS separators. |

### InitializeResult

InitializeResult tells the host who the plugin is and what it can do.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `protocolVersion` | `int` | yes | ProtocolVersion is the version the plugin chose from the host's list. |
| `plugin` | `Peer` | yes |  |
| `kind` | `PluginKind` | yes |  |
| `capabilities` | `Capabilities` | yes |  |
| `concurrency` | `int` | no | Concurrency is how many requests the plugin can work on at the same time; the host never has more in flight. Absent or 0 means 1, which is what every plugin that says nothing does: the host sends the next request only after the previous one is answered. N above 1 allows that many, and -1 leaves the number to the host (it keeps 32 in flight). Any other negative value is invalid and read as 1. |
| `languages` | `[]string` | no | Languages are the values a user passes to --lang to reach this plugin. |
| `manifests` | `[]string` | no | Manifests are file base names (package.json, go.mod), or patterns of them (*.csproj), whose presence in a directory may make it a project of this plugin. |
| `roles` | `[]Role` | no | Roles are the levels the plugin can be chosen for in a project's make-up (a stack, a recipe). |
| `provides` | `[]string` | no | Provides are the slots a toolchain plugin fills: the task names it owns ("test" for a test runner). |
| `requires` | `[]PluginRef` | no | Requires are plugins that must also be in the project for this one to work: a Node toolchain requires the node language plugin. |
| `conflicts` | `[]PluginRef` | no | Conflicts are plugins that must not be in the same project. |
| `stacks` | `[]Stack` | no | Stacks are ready-made stacks the plugin publishes. |
| `kinds` | `[]ProjectKind` | no | Kinds are the project kinds the plugin can scaffold with generate. Absent or empty means all three (app, lib and package): what every plugin that says nothing does. A plugin that scaffolds only some, such as an app, lists them, so the host offers and the conformance suite asks for just those. |

**What `languages` means for a convention plugin.** It is the list of project languages the plugin
applies to, and the only place that says so. The host asks the plugin (`contribute`, `doctor`,
`generate`) only about projects whose `language` is in the list, and `smith doctor` names the
projects it left out. An empty or absent list means every language. `smith.toml` can narrow the
list for a workspace (`languages = [...]` in `[conventions.<name>]`) but never widen it.

### Peer

Peer names one side of the connection.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes |  |
| `version` | `string` | yes |  |

### Capabilities

Capabilities are the optional methods a plugin implements.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `discover` | `bool` | no |  |
| `generate` | `bool` | no |  |
| `version` | `bool` | no |  |
| `publish` | `bool` | no |  |
| `ci` | `bool` | no |  |
| `doctor` | `bool` | no |  |
| `contribute` | `bool` | no | Contribute is a convention plugin's: it adds tasks to projects other plugins discovered. |
| `remove` | `bool` | no | Remove cleans up a removed project's traces outside its directory. |
| `changeKind` | `bool` | no | ChangeKind answers what changing a project's kind edits (smith kind). |
| `published` | `bool` | no | Published reports what a registry has of a project. |
| `dependencies` | `bool` | no | Dependencies installs and uninstalls packages in its projects. |
| `declarations` | `bool` | no | Declarations lists and rewrites declared dependency specs (sync, up). |
| `latest` | `bool` | no | Latest reports the newest releases of dependencies and pinned tools. |
| `cancel` | `bool` | no | Cancel says the plugin stops the work of a request when the host sends a cancel notification with its id, then answers that request with its result (if the work was already done) or the request-cancelled error. A host never sends the notification to a plugin without it. |

### CancelParams

CancelParams are the params of the cancel notification, which has no response.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `id` | `int64` | yes | ID is the id of the in-flight request to stop. An id the plugin does not know, or whose request is already answered, is ignored. |

### PluginRef

PluginRef names the plugins a requirement or a conflict is about.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `level` | `Role` | no | Level is the role the plugin is chosen for in the project. |
| `name` | `string` | no | Name is the plugin's name, or for a language plugin one of its languages ("node", "javascript"). |
| `slot` | `string` | no | Slot is a task the plugin provides ("test"). |

### Stack

Stack is a named set of plugins a project is built with: one language plugin, and the toolchain, structure, scaffold and deploy plugins on top.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes |  |
| `description` | `string` | no |  |
| `language` | `string` | yes | Language is the language plugin that discovers the stack's projects. |
| `toolchain` | `[]string` | no | Toolchain plugins decide how projects build, test and lint, in order: a later one may replace a task an earlier one added. |
| `structure` | `[]string` | no |  |
| `scaffold` | `string` | no | Scaffold generates new projects; empty means the language plugin does. |
| `deploy` | `[]string` | no |  |

### DiscoverParams

DiscoverParams asks a plugin to describe the projects behind manifests the host found.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `manifests` | `[]string` | yes |  |

### DiscoverResult

DiscoverResult describes every manifest that is a project of this plugin.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `projects` | `[]Project` | yes |  |

### Project

Project is one unit of the monorepo as a plugin sees it.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes | Name is unique across the workspace; it is what users type. |
| `root` | `string` | yes | Root is the workspace-relative directory, forward slashes. |
| `language` | `string` | yes |  |
| `kind` | `ProjectKind` | yes |  |
| `exports` | `[]string` | no | Exports are the identifiers other projects import this one by: an npm package name, a Go module path, a Python import name. |
| `imports` | `[]string` | no | Imports are identifiers this project's sources import. |
| `version` | `string` | no | Version is the version recorded in the manifest, if it has one. |
| `toolchain` | `string` | no | Toolchain is the version of the compiler or runtime the project's tasks run with ("node 20.11.0", "go 1.23.4"), as the plugin probed it. |
| `publishTo` | `string` | no | PublishTo is where a package publishes, when its manifest says: "pub.dev" for the public registry that takes a package only from a run on its tag (so a release does not publish it itself), "none" for a package that is never published, or a repository URL. |
| `description` | `string` | no | Description is the one-line description the manifest records (package.json description, pyproject.toml project.description, the pom's <description>), if it has one. |
| `tagPrefix` | `string` | no | TagPrefix overrides the release tag prefix; empty means "<name>@". |
| `tasks` | `map[string]Task` | no |  |
| `tags` | `[]string` | no |  |
| `stack` | `string` | no | Stack is the stack recorded in the project's manifest, if any. |
| `settings` | `map[string]map[string]any` | no | Settings are plugins' settings for this project from smith.toml, by plugin name: [projects.<project>.settings.<name>] (what a recipe's settings became when it created the project), and for a convention its [conventions.<name>] settings and the project's [projects.<project>.<name>] on top. |
| `registries` | `*Registries` | no | Registries are the private registries the project installs from and publishes to, from smith.toml [registries]; nil means the public ones. |
| `releaseAssets` | `*AssetSpec` | no | ReleaseAssets is what an app's release builds, from smith.toml [projects.<project>.release]: the platforms and file names, so a plugin can check (doctor) and prepare (CI) for them. Nil for a project that has no release files. |

### Task

Task is a command the host can run, hash and cache.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `command` | `[]string` | yes | Command is an argv array; it is never passed through a shell. |
| `cwd` | `string` | no | Cwd is workspace-relative; empty means the project root. |
| `env` | `map[string]string` | no |  |
| `inputs` | `[]string` | no | Inputs are globs relative to the project root. |
| `outputs` | `[]string` | no | Outputs are directories relative to the project root, restored on a cache hit. |
| `dependsOn` | `[]string` | no | DependsOn lists task names: "build" runs this project's build first, "^build" runs the build of every project this one imports first. |
| `cache` | `bool` | no |  |
| `description` | `string` | no | Description is shown by `smith explain` and the TUI. |
| `verify` | `bool` | no | Verify asks CI's verify phase (smith ci verify) to run the task too, after lint, typecheck, test and build, for the affected projects that have it: an app's container image, so a broken Dockerfile fails the pull request, not the release. |
| `serve` | `string` | no | Serve names a task ("web:start", or "start" in the same project) that smith starts before this one and stops after it, whatever the result: the app an end-to-end suite tests. |
| `serveUrl` | `string` | no | ServeURL is what smith waits for (an HTTP 2xx) before it runs the task, and passes to it as E2E_BASE_URL. |
| `serveTimeout` | `string` | no | ServeTimeout is how long to wait for ServeURL, as a Go duration ("60s"); empty means 60s. |

### Registries

Registries are the registries a project uses, set by the host from smith.toml: Install is where the workspace installs packages from, Publish where the project is published.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `install` | `*Registry` | no |  |
| `publish` | `*Registry` | no |  |
| `noPublish` | `bool` | no | NoPublish means smith.toml says registry = "none": the release versions and tags the project and publishes it to no registry (a registry that builds from the tag, such as JitPack, does the rest). |

### Registry

Registry is a package registry smith.toml names in [registries.<name>]: a private feed (Azure Artifacts, GitHub Packages, any token-protected registry) a language plugin installs from or publishes to.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes |  |
| `auth` | `string` | yes | Auth is how the registry is signed in to: AuthAzureArtifacts, AuthGitHubPackages or AuthToken. |
| `tokenEnv` | `string` | no | TokenEnv is the environment variable holding the token, for AuthToken. |
| `npm` | `string` | no | NPM is the npm registry URL. |
| `pypi` | `string` | no | PyPI is the simple index pip installs from; PyPIUpload is where twine uploads. |
| `pypiUpload` | `string` | no |  |
| `nuget` | `string` | no | NuGet is the NuGet v3 service index. |
| `maven` | `string` | no | Maven is the Maven repository packages deploy to: an https URL, or MavenGitHub for the GitHub Packages repository of the repository the release runs in. |

### GenerateParams

GenerateParams asks a plugin to scaffold a project.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `kind` | `ProjectKind` | yes |  |
| `name` | `string` | yes |  |
| `directory` | `string` | yes | Directory is workspace-relative, forward slashes; it does not exist yet. |
| `workspace` | `WorkspaceInfo` | yes | Workspace carries what a template needs to know about the monorepo. |
| `stack` | `string` | no | Stack is the stack the project is created with. |
| `convention` | `bool` | no | Convention is true when the plugin is asked as a workspace convention ([conventions.<name>] applying to the new project), not as one of the stack's plugins. |
| `settings` | `map[string]any` | no | Settings are this plugin's own settings for the new project, from the recipe the project is created from (its settings.<plugin> table); the same table is later the plugin's entry in Project.Settings. |

### WorkspaceInfo

WorkspaceInfo is the subset of workspace settings plugins may use.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes |  |
| `scope` | `string` | no | Scope is the npm scope for packages, with its @ (e.g. |
| `modulePrefix` | `string` | no | ModulePrefix is the Go module path prefix, e.g. |

### GenerateResult

GenerateResult lists what the plugin wrote, workspace-relative.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `files` | `[]string` | yes |  |
| `notes` | `[]string` | no | Notes are printed to the user after generation. |
| `installPending` | `bool` | no | InstallPending is true when the plugin changed what the workspace depends on and left installing it to the host. A project is made by several plugins in turn; the host asks each plugin that reported this to install once (dependencies/install with Quiet set) after the last of them has generated, instead of once per plugin. |

### Contribution

Contribution is a convention's addition to one project.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `tasks` | `map[string]Task` | no | Tasks are new tasks. |
| `replace` | `map[string]Task` | no | Replace are tasks that replace the project's task of the same name, or are added when it has none. |
| `dependsOn` | `map[string][]string` | no | DependsOn adds dependencies to tasks the project already has, by task name ({"lint": ["slice-check"]}). |

### ContributeParams

ContributeParams hands a plugin the projects it applies to: those whose stack includes it, or, for a workspace-wide convention, those of the languages it supports that smith.toml enables it for.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `projects` | `[]Project` | yes |  |

### ContributeResult

ContributeResult is what a convention adds, by project name.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `projects` | `map[string]Contribution` | no |  |

### VersionParams

VersionParams asks a plugin to record a new version in a project manifest.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `project` | `Project` | yes |  |
| `version` | `string` | yes |  |
| `dependencies` | `map[string]string` | no | Dependencies maps exported identifiers of other workspace projects to their new versions, so the plugin can update dependency ranges. |
| `dryRun` | `bool` | no |  |

### VersionResult

VersionResult lists the files that were (or, in a dry run, would be) changed.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `files` | `[]string` | yes |  |

### PublishParams

PublishParams asks a plugin to publish one project at one version.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `project` | `Project` | yes |  |
| `version` | `string` | yes |  |
| `dryRun` | `bool` | no |  |
| `assets` | `*AssetSpec` | no | Assets, when set, asks the plugin to build the files attached to the release (an app's per-platform binaries) instead of publishing to a registry. |
| `platform` | `string` | no | Platform, when set, is the "os/arch" (Go's spelling) of the runner a native project is published from: each native job publishes its own platform's build (a VS Code extension's target) and nothing else. |
| `tagged` | `bool` | no | Tagged says the publish runs on the project's own tag, in the run the registry trusts (pub.dev's automated publishing); a registry that takes a package only from such a run is published to only then. |
| `commit` | `string` | no | Commit is the full hash of the commit being released, and Branch the release branch, for a plugin that tags what it publishes with them (a container image's sha-<short> and branch tags). |
| `branch` | `string` | no |  |

### AssetSpec

AssetSpec describes the release files of an app.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes | Name is the file name pattern; placeholders are {product}, {version}, {os}, {arch} and {ext}. |
| `product` | `string` | no | Product fills {product}; it defaults to the project name. |
| `platforms` | `[]string` | yes | Platforms are "os/arch" pairs in the language's own spelling. |
| `osAlias` | `map[string]string` | no | OSAlias and ArchAlias rename an os or arch in file names only. |
| `archAlias` | `map[string]string` | no |  |
| `versionVariable` | `string` | no | VersionVariable is the symbol the version is stamped into, for languages that stamp at link time (Go: main.version). |
| `outDir` | `string` | yes | OutDir is the workspace-relative directory the files go to. |

### PublishResult

PublishResult reports what happened.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `published` | `bool` | yes |  |
| `skipped` | `string` | no |  |
| `registry` | `string` | no |  |
| `assets` | `[]string` | no | Assets are the workspace-relative files built for the release, in upload order; in a dry run, the files that would be built. |
| `assetPlatforms` | `map[string]string` | no | AssetPlatforms maps each archive among Assets to the "os/arch" it was built for, so the host can describe them (package manager manifests). |
| `binary` | `string` | no | Binary is the executable's name inside each archive, without ".exe". |
| `images` | `[]PublishedImage` | no | Images are the container images pushed, one per registry, for a deploy step or a GitOps manifest to refer to by digest. |

### PublishedImage

PublishedImage is one container image a release pushed.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `reference` | `string` | yes | Reference is the image without a tag: ghcr.io/acme/api. |
| `tags` | `[]string` | yes | Tags are the tags pushed, in the order they were. |
| `digest` | `string` | no | Digest is the manifest digest the registry holds ("sha256:..."); empty when it could not be read. |

### PublishedParams

PublishedParams asks which version of a project its registry has.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `project` | `Project` | yes |  |

### PublishedResult

PublishedResult is the newest version the registry has of a project.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `version` | `string` | no | Version is empty when the registry does not know the project: its first release is still to come. |
| `registry` | `string` | no | Registry names where the plugin looked, for messages. |
| `unreachable` | `bool` | no | Unreachable is true when the registry could not be asked (offline, an error), so nothing can be concluded; Detail says why. |
| `detail` | `string` | no |  |

### CIParams

CIParams gives the plugin the projects it owns so it can describe what CI needs for them.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `projects` | `[]Project` | yes |  |
| `audit` | `AuditSettings` | no | Audit is the workspace's opt-in to policy checks beyond the vulnerability scan; the zero value asks for none. |

### AuditSettings

AuditSettings are the workspace's audit policy choices ([ci] audit_*).

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `policies` | `[]string` | no | Policies are the ecosystems ("node", "go", "python", "rust") whose policy checks run in `smith ci audit`, as extra Audit steps. |
| `warn` | `[]string` | no | Warn are the policies whose checks only warn instead of blocking. |
| `allowInstallScripts` | `[]string` | no | AllowInstallScripts names npm packages whose install scripts the node policy accepts. |
| `maxUnsafe` | `int` | no | MaxUnsafe is the number of unsafe blocks the rust policy accepts in all (zero: none). |

### CIResult

CIResult is the plugin's contribution to every generated pipeline.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `stores` | `[]PackageStore` | no | Stores are the package managers' download stores the pipeline caches between runs, keyed on the lockfiles that decide what is in them. |
| `toolchains` | `[]Toolchain` | no | Toolchains are rendered as provider-native setup steps. |
| `setup` | `[]Step` | no | Setup steps run in `smith ci setup`, in order, from the workspace root. |
| `audit` | `[]Step` | no | Audit steps run in `smith ci audit`. |
| `postRelease` | `[]Step` | no | PostRelease steps run at the end of `smith ci release`, whether or not anything was released, with the release's credentials: work that must follow every push to the release branch (setting a registry description from a file). |
| `credentials` | `[]Credential` | no | Credentials names the environment variables `smith ci release` needs. |
| `permissions` | `[]string` | no | Permissions are the GitHub Actions token permissions the release needs beyond the defaults ("packages: write" to push to ghcr.io). |
| `authenticate` | `[]Toolchain` | no | Authenticate are provider-native steps that sign in to a private registry before setup (Azure Pipelines' npmAuthenticate@0, say). |

### Toolchain

Toolchain is a runtime the pipeline must install, and how.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes |  |
| `version` | `string` | yes |  |
| `github` | `*ActionStep` | no | GitHub is the GitHub Actions step that installs it (an action). |
| `azure` | `*ActionStep` | no | Azure is the Azure Pipelines step that installs it (a task). |
| `install` | `string` | no | Install is a shell command that installs it, run on a provider the plugin gives no native step for. |
| `devcontainer` | `*DevcontainerFeature` | no | Devcontainer is the dev container feature that installs it at the same version, so the editor's container matches CI. |

### ActionStep

ActionStep is a provider-native step: a GitHub action or an Azure task, with its inputs in the order they are written.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `uses` | `string` | yes | Uses is the action ("actions/setup-go@v7") or task ("GoTool@0"). |
| `inputs` | `[]Input` | no |  |

### Input

Input is one input of an ActionStep.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes |  |
| `value` | `string` | yes |  |

### DevcontainerFeature

DevcontainerFeature is a dev container feature and its options (https://containers.dev/features).

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `id` | `string` | yes | ID is the feature reference, e.g. |
| `options` | `map[string]string` | no |  |

### Step

Step is one command a CI phase runs.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes |  |
| `command` | `[]string` | yes |  |
| `dir` | `string` | no | Dir is the workspace-relative directory to run in; empty means the workspace root. |
| `blocking` | `bool` | yes | Blocking false makes a failure a warning instead of an error. |
| `cacheFiles` | `[]string` | no | CacheFiles are workspace-relative files whose content decides what an audit step finds (a lockfile). |

### Credential

Credential is a secret the release phase needs.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `env` | `string` | yes |  |
| `description` | `string` | yes |  |
| `required` | `bool` | yes | Required false means release works without it (e.g. |
| `setup` | `bool` | no | Setup gives it to `smith ci setup` (and the native job's setup) as well: installs from a private registry need it, not only the release. |
| `tasks` | `bool` | no | Tasks gives it to every step that runs tasks (verify, pack, native) as well: the remote cache's key and credentials are read there, not only by the release. |

### PackageStore

PackageStore is where a package manager keeps what it downloaded (npm's ~/.npm, Go's module cache).

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes | Name identifies the store ("npm", "go") and is part of its cache key. |
| `path` | `string` | yes | Path is the store's default location, which GitHub Actions caches as it is ("~/.npm"). |
| `env` | `string` | yes | Env is the environment variable that moves the store (NPM_CONFIG_CACHE, GOMODCACHE). |
| `files` | `[]string` | yes | Files are the lockfiles that key the store: workspace-relative paths or globs. |

### DoctorParams

DoctorParams gives the plugin its projects to check.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `projects` | `[]Project` | yes |  |

### DoctorResult

DoctorResult carries the plugin's findings.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `findings` | `[]Finding` | yes |  |

### Finding

Finding is one check.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `check` | `string` | yes |  |
| `ok` | `bool` | yes |  |
| `detail` | `string` | no |  |
| `remedy` | `string` | no |  |
| `warning` | `bool` | no | Warning true reports a problem without failing doctor. |
| `fix` | `*Fix` | no | Fix is the change that clears a failed finding, when the plugin can write it; smith doctor --fix shows it as a diff before applying it. |

### Fix

Fix is a change that clears a failed finding: files' new contents, then commands to run, and questions the user answers first.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `description` | `string` | yes |  |
| `safe` | `bool` | yes | Safe marks a fix that only restores what smith or the plugin owns and loses nothing of the user's; doctor --fix --yes applies only safe fixes with no Choices. |
| `files` | `[]FileEdit` | no | Files are written first, as whole new contents. |
| `steps` | `[]FixStep` | no | Steps run after the files, in order, from the workspace root; the first one that fails stops the rest. |
| `choices` | `[]FixChoice` | no | Choices are questions the user answers before the fix applies, in order. |
| `applies_to` | `[]string` | no | AppliesTo names the projects the fix is for. |

### FixStep

FixStep is a command to run in the workspace.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `description` | `string` | yes | Description says in words what the command does. |
| `argv` | `[]string` | yes | Argv is the program and its arguments; no shell is involved. |
| `dir` | `string` | no | Dir is where it runs, workspace-relative; empty is the workspace root. |

### FixChoice

FixChoice is a question with options, one wizard step.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `question` | `string` | yes |  |
| `options` | `[]FixChoiceOption` | yes |  |

### FixChoiceOption

FixChoiceOption is one answer: a label, and the Fix it adds when picked.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `label` | `string` | yes |  |
| `fix` | `Fix` | yes |  |

### FileEdit

FileEdit is one file's new content, workspace-relative; Delete removes it.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `path` | `string` | yes |  |
| `content` | `string` | no |  |
| `delete` | `bool` | no |  |

### RemoveParams

RemoveParams tells a plugin a project is being removed.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `project` | `Project` | yes |  |
| `dryRun` | `bool` | no | DryRun asks what would change, without changing it. |

### RemoveResult

RemoveResult lists the workspace-relative files the plugin changed (or, in a dry run, would change), and notes for the user.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `files` | `[]string` | no |  |
| `notes` | `[]string` | no |  |

### ChangeKindParams

ChangeKindParams asks a plugin what changing a project's kind edits. The plugin writes nothing: it answers with the files' new contents, and the host applies them, then moves the project's directory.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `project` | `Project` | yes | Project is the project as it is now. |
| `to` | `ProjectKind` | yes | To is the kind it becomes. |
| `newRoot` | `string` | yes | NewRoot is the workspace-relative directory the host moves the project to afterwards; it equals Project.Root when the directory stays. |
| `projects` | `[]Project` | no | Projects are every project of the workspace, so the plugin can find the ones whose manifests and sources name this one. |

### ChangeKindResult

ChangeKindResult is what the change edits. Every path in Files is the path the file has before the move.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `refused` | `string` | no | Refused says why this kind makes no sense for the project's language ("a Bicep template is deployed, never published"); the host stops and shows it. Empty when the change is fine. |
| `files` | `[]FileEdit` | no | Files are the edits: manifests, workspace files and dependents that name the project's kind or its directory. |
| `warnings` | `[]string` | no | Warnings name what the change loses or leaves for the user (a binary an app has and a lib does not, a lockfile to refresh). |

### InstallParams

InstallParams asks a plugin to add packages to one of its projects with the ecosystem's own toolchain, or, with no project, to install every dependency of the projects it serves (a fresh clone's first step).

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `project` | `*Project` | no | Project is the project the packages go into; nil installs everything. |
| `projects` | `[]Project` | no | Projects are the projects the plugin serves, when Project is nil. |
| `packages` | `[]string` | no | Packages are names or pinned specs as the ecosystem writes them (left-pad, requests>=2.31, github.com/pkg/errors@v0.9.1). |
| `dev` | `bool` | no | Dev adds them as development dependencies where the ecosystem has such a thing; a plugin without one says so in a note. |
| `quiet` | `bool` | no | Quiet keeps the package manager's output unless it fails: the host sets it when it installs as one step of a larger job (smith add). |

### UninstallParams

UninstallParams asks a plugin to remove packages from one of its projects.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `project` | `Project` | yes |  |
| `packages` | `[]string` | yes |  |

### DependencyResult

DependencyResult lists the workspace-relative files the plugin changed and notes for the user.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `files` | `[]string` | no |  |
| `notes` | `[]string` | no |  |

### DeclaredParams

DeclaredParams asks for the dependencies the plugin's projects declare (and the workspace root's manifest, when the ecosystem has one).

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `projects` | `[]Project` | yes |  |

### Declaration

Declaration is one dependency one manifest declares.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `project` | `string` | no | Project is the declaring project's name, "" for the root manifest. |
| `file` | `string` | yes | File is the manifest, workspace-relative. |
| `name` | `string` | yes | Name is the dependency's name as the ecosystem compares names. |
| `spec` | `string` | yes | Spec is the declared range or version (^1.2.3, >=2.31), "" when the manifest gives none. |
| `section` | `string` | yes | Section is where it is declared: dependencies, devDependencies, dev... |
| `peer` | `bool` | no | Peer declarations state what a package works with, not what is installed: sync and up leave them alone. |
| `rewritable` | `bool` | yes | Rewritable is false for a spec smith must not rewrite: a URL, a git ref, a workspace alias, a compound range, environment markers. |

### DeclaredResult

DeclaredResult is every declaration, and the version installed of each name where the plugin can tell.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `declarations` | `[]Declaration` | yes |  |
| `installed` | `map[string]string` | no |  |

### RewriteParams

RewriteParams asks a plugin to change declared specs in place.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `changes` | `[]Rewrite` | yes |  |

### Rewrite

Rewrite changes one declaration's spec.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `declaration` | `Declaration` | yes |  |
| `spec` | `string` | yes |  |

### LatestParams

LatestParams asks for the newest stable release of each dependency, and, with Tools, how the tools the plugin pins compare with their latest.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `names` | `[]string` | no |  |
| `tools` | `bool` | no |  |

### LatestResult

LatestResult maps each name to its newest stable version.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `latest` | `map[string]string` | no |  |
| `unreachable` | `map[string]string` | no |  |
| `tools` | `[]ToolPin` | no |  |

### ToolPin

ToolPin is a tool the plugin installs at a pinned version (in CI, the dev container, the venv) against the newest release; informational.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `name` | `string` | yes |  |
| `pinned` | `string` | yes |  |
| `latest` | `string` | no | Latest is empty when its release feed could not be reached. |
| `newer` | `bool` | no | Newer is true when Latest is a release the pin does not cover. |

### Request

Request is a JSON-RPC 2.0 request, or a notification when ID is nil.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `jsonrpc` | `string` | yes |  |
| `id` | `*int64` | no |  |
| `method` | `string` | yes |  |
| `params` | `json.RawMessage` | no |  |

### Response

Response is a JSON-RPC 2.0 response.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `jsonrpc` | `string` | yes |  |
| `id` | `*int64` | yes |  |
| `result` | `json.RawMessage` | no |  |
| `error` | `*ResponseError` | no |  |

### ResponseError

ResponseError is a JSON-RPC 2.0 error object.

| JSON field | Type | Required | Meaning |
|---|---|---|---|
| `code` | `int` | yes |  |
| `message` | `string` | yes |  |

## Other rules

- Unknown fields are ignored on both sides, so a plugin written today keeps working when the protocol
  grows. An addition is a new optional field or a new capability; only a breaking change raises
  `protocolVersion`.
- `protocolVersion` in the `initialize` result must be one of the host's `protocolVersions`
  (today `[1]`), or the host refuses the plugin.
- A task's `command` is an argv array, never run through a shell. `{smith}`, `{workspaceRoot}` and
  `{projectRoot}` in a task command are replaced by the path of the running smith, the workspace root
  and the project's directory when the task runs.
- A build writes into `protocol.ProjectBuildDir(root)` (`dist/<project dir>`) and declares it in
  `outputs` as `protocol.BuildDirFromProject(root)` (`../../dist/apps/web`, relative to the project).
  `smith doctor` fails an output elsewhere.
- Skipping with a reason is a success: `publish` of an already published version answers
  `{"published": false, "skipped": "..."}`, which is what makes a failed release safe to run again.

## See also

- [Creating plugins](README.md)
- [Plugins (reference)](../../reference/plugins.md#the-protocol-version-1)
- `packages/plugin-sdk/protocol`: the types themselves
