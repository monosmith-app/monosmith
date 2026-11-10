# ADR 0041: `smith add` follows the dialect of sibling projects

- Status: accepted
- Date: 2026-10-31

## Context

`smith add` builds a new project from a fixed default: a stack's templates. A workspace that was
converted from Nx, Turborepo or another tool has its own habits. Its libs build with esbuild, test
with `@swc/jest`, point `main` at `dist/` and keep three tsconfig files. The new project differs
from all of them, and the person fixes it by hand
(epic [#819](https://github.com/monosmith-app/monosmith/issues/819), phase 1,
issue [#820](https://github.com/monosmith-app/monosmith/issues/820)).

The core of smith must stay free of any language. A plugin never calls another plugin.

## Definition: dialect

The dialect of a project is its shape: a small set of named facts, each one a short string.
For Node projects these facts are:

| Key | Value |
|---|---|
| `build.tool` | the program of the `build` script (`tsc`, `esbuild`, `tsup`), or `none` |
| `test.runner` | `jest`, `ts-jest`, `@swc/jest`, `vitest`, `node:test`, or `none` |
| `exports.target` | where `exports`, `main`, `module` and `types` point: `dist`, `src` or `other`, or `none` |
| `tsconfig.files` | the sorted `tsconfig*.json` files of the project folder |
| `lint.config` | `project` when the folder has its own ESLint config, else `root` |
| `scripts` | the sorted script names |
| `layout` | the sorted top-level folders of the project, without `node_modules` and `dist` |
| `build.output` | where the build script writes: `project` (its own folder, `dist`), `workspace` (`../../dist/<dir>`), or `none` |
| `exports.map` | `map` when package.json has an `exports` object, else `none` |
| `jest.config` | `file` for a `jest.config.*` file, `package` for the `jest` field, or `none` |

The core knows no key. It treats a shape as a map of strings.

## Options considered

**Read the sibling manifests in the core.** Rejected: the core would learn `package.json`, Jest and
tsconfig. It cannot do the same for Go, Python and the rest.

**Use the tasks that discovery already returns.** Rejected: a task holds an `npm run build` argv,
not the script behind it. The test runner, the build tool and the exports are not in it.

**A new `project/shape` method.** Rejected: a second pass over every manifest, a new capability for
every plugin, and a shape that can differ from the project the host discovered.

**A fixed template per dialect in each plugin (`node-swc`, `node-esbuild`).** Rejected: it is the
plugin-per-combination growth that phase 2 of the epic removes.

## Decision

Three optional fields join the protocol. The protocol stays at version 1.

- `Project.shape` (discovery): the shape of the project, a map of strings. The plugin that
  discovers the project fills it from the manifest and files it already reads. Absent means the
  plugin does not describe shapes.
- `InitializeResult.follows`: the shape keys the plugin applies when it generates. The host asks
  a question about a key only if a plugin of the stack follows it.
- `GenerateParams.shape`: the shape the new project should have, sent to every plugin of the
  stack. A plugin applies the keys it knows and ignores the rest.

The host does the generic part:

1. It takes the projects of the same kind and language as the new one that report a shape.
2. For each key it counts the values. A value is **dominant** when at least 3 projects have it
   and at least 2/3 of the projects that report the key have it.
3. It prints what the siblings agree on with the reason ("22 of 22 packages use `@swc/jest` as
   `test.runner`") and sends the dominant values in `GenerateParams.shape`.
4. After generation, the host reads the shape of the new project, which discovery gives back. A
   dominant key whose value differs is a warning: the stack cannot follow it. The warning names
   the key, the siblings' value and the new value.

**Thresholds.** Fewer than 3 siblings with a shape: nothing is inferred, and nothing is printed.
A key where no value is dominant is **split**.

**Ties.** A split key that a plugin follows is asked, never guessed. The question lists the values
with their counts and the built-in default. A run with no terminal, or with `CI` set, never
blocks: it uses the built-in default for the key and prints a warning that names the split.
A split key that no plugin follows is only reported.

**`--fresh`.** `smith add --fresh` skips the inference. The project gets the built-in default and
nothing about siblings is printed.

**Node.** `libs/plugin-node` describes the shape in discovery. The toolchain plugins follow it:
`typescript` follows `build.tool` (`tsc`, `esbuild`) and `exports.target` (`dist`, `src`); `jest`
follows `test.runner` (`jest`, `ts-jest`, `@swc/jest`); `vitest` follows `test.runner` (`vitest`).
`typescript` also follows `build.output` (the build goes to the project's own `dist`, with
`--declarationDir`), `exports.map` and `tsconfig.files`: siblings with a `tsconfig.lib.json` get
it, a `tsconfig.json` that only references it and, when they have one, a `tsconfig.spec.json`;
`jest` follows `jest.config` (a `jest.config.cts`, and a `.spec.swcrc` for `@swc/jest`).
A key that no plugin follows is reported and warned about. A sibling set with only a
`tsconfig.json` is not followed: the new project keeps its default set and gets the warning.

## Consequences

- The core gains one generic step and no language knowledge. A plugin for another language adds
  shape facts and `follows` keys with no change to the host.
- Plugins that say nothing change nothing: the three fields are optional.
- A stack cannot switch to another plugin because of a shape. Siblings that use Vitest in a
  workspace whose default stack uses Jest give a warning that points to `--stack`.
- The 2/3 threshold ignores a minority. A workspace with a few odd projects still has a dominant
  shape. A workspace split in two halves asks.
- Phases 2 and 3 of the epic (axes, shared facts) can replace the `follows` list with data. The
  shape stays the measure of what siblings do.

## Sources

- Issues [#819](https://github.com/monosmith-app/monosmith/issues/819) and
  [#820](https://github.com/monosmith-app/monosmith/issues/820)
- [ADR 0039](0039-a-plugin-declares-the-project-kinds-it-scaffolds.md): the same kind of optional
  protocol field
- [The plugin protocol](../plugin-protocol.md)
