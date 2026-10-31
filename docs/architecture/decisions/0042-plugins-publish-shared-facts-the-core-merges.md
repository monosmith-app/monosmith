# ADR 0042: Plugins publish shared facts that the core merges

- Status: accepted
- Date: 2026-10-31

## Context

A plugin Contribution holds tasks, replacements and dependencies. Facts that several tools
need are written by hand into each tool's own config: an import alias lives in `tsconfig.json`,
in the Jest `moduleNameMapper` and in the Vite `resolve.alias`; the source folder lives in
all three. Changing one means editing every file, and nothing says the values must agree
(epic [#819](https://github.com/monosmith-app/monosmith/issues/819), phase 3, issue
[#822](https://github.com/monosmith-app/monosmith/issues/822)).

The rules of the epic apply: the core stays language-free, plugins never call each other,
the protocol stays at version 1, and there are no shims.

## Options considered

**Plugins read each other's settings.** Rejected: plugins would depend on each other by name,
and a Jest plugin would need to know the TypeScript plugin.

**Put the shared values only in `smith.toml`.** Rejected: the value has a sensible default
that a plugin knows (a TypeScript project has sources in `src`), and a person should write
only what differs.

**A second request after contribute, to hand each plugin the merged facts.** Rejected: it
doubles the calls to every plugin to serve a few of them. The merged facts travel in requests
a plugin already receives.

## Decision

A fact is a typed value under a namespaced key, `<namespace>.<name>`: `module.aliases` (an
object of strings), `source.roots`, `test.patterns` and `build.outputs` (arrays of strings),
`runtime.node` (a string). Those keys are well-known: the protocol validates their type. A plugin
may publish any other namespaced key, for example `acme.region`, with any JSON value.

Three parties publish facts:

- A plugin, through the new optional `Contribution.facts`. These are plugin defaults.
- The person, in `smith.toml`: `[facts]` for the workspace and `[projects.<name>.facts]` for one
  project. Both tables have a schema.
- A recipe is a fourth scope between the plugin defaults and the workspace. The merge knows it;
  no recipe carries facts yet.

The core merges them per project and per key. The narrowest scope that sets a key wins:
project, then workspace, then recipe, then plugin default. A wider scope is not looked at.
Two plugin defaults for one key that differ are an error that names the project, the key, both
plugins and both values; equal values agree. A `smith.toml` fact decides such a conflict,
because it is a narrower scope. An invalid fact (a bad key, or a well-known key of the wrong
type) is an error that names its table or plugin.

Plugins never see each other's facts as such. A plugin publishes values; it does not know
who reads them. A plugin reads the merged value, not the sources.

A plugin reads facts from a field in a request it already gets, so the change is additive and
the protocol stays at version 1:

- `Project.facts` in every request that carries a project (`contribute`, `ci`, `doctor`, ...).
- `GenerateParams.facts` in `generate`, from `smith.toml` only: the project does not exist yet,
  and no plugin has published for it.

An absent key means no one set it, and the plugin uses its own default. So behaviour without
overrides is the same as before. `Project.facts` holds `smith.toml` facts plus the plugin
defaults published so far: the `contribute` requests of the stack's plugins carry the `smith.toml`
facts, the `contribute` requests of the conventions also carry the stack plugins' defaults, and
later requests carry all of them. One round of `contribute` is enough because a plugin that
must adapt its tasks to a fact falls back to its default when the key is absent.

`smith explain <project> --facts` shows each fact's value, scope and source; `--fact <key>`
shows one.

## Consequences

- One edit in `smith.toml` reaches every tool that reads the fact. The Node toolchains read
  `module.aliases` when they scaffold a project: TypeScript writes `paths`, Jest writes
  `moduleNameMapper`, Vite writes `resolve.alias`, Vitest writes an alias config.
- The facts reach a project when its files are generated. Files already on disk follow only
  when a later phase (derived files with an ownership record) regenerates them.
- The SDKs of Go, TypeScript, Python and C# carry the field through the generated protocol
  types. `smith plugin check` checks that a plugin accepts the field and publishes valid facts.
- Plugin defaults do not reach the `contribute` request of the plugin that publishes them or
  of its stack siblings. A fact that two sibling plugins must agree on belongs in `smith.toml`.

## Sources

- Epic [#819](https://github.com/monosmith-app/monosmith/issues/819), issue
  [#822](https://github.com/monosmith-app/monosmith/issues/822)
- [The plugin protocol](../plugin-protocol.md), [shared facts](../../plugins/creating/shared-facts.md)
- `packages/plugin-sdk/protocol/fact_keys_contract.go`, `apps/smith/factmerge`
