# The vscode plugin

The `vscode` plugin scaffolds, builds, packs and publishes VS Code extensions written in
TypeScript. It plays three roles: the scaffold and `build`/`typecheck` toolchain of the
`vscode-extension` stack, and a deploy plugin that releases an extension that is an app.

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `scaffold`, `toolchain`, `deploy` |
| Provides (slots) | `build`, `typecheck` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate`, `publish`, `ci`, `contribute` |
| Stack it publishes | `vscode-extension`: scaffold `vscode`; toolchain `vscode`, `eslint`, `vitest` |
| Project kinds | `package` (published to the Marketplace) and `app` (an extension kept to the workspace's own `.vsix`); a `lib` is refused |

## What it does

An extension is a package or an app whose `package.json` has `engines.vscode`. The scripts the
scaffold writes become tasks through the [node plugin](node.md#tasks):

| Script | Command |
|---|---|
| `build` | `esbuild src/main.ts --bundle --platform=node --format=cjs --target=node20 --external:vscode --sourcemap --outfile=../../dist/<dir>/main.js` |
| `typecheck` | `tsc -p tsconfig.json --noEmit` |
| `lint`, `test` | from `eslint` and `vitest` |

The node plugin adds `package` (`smith npm-pack <dir>`, depending on `build`), which packs
`dist/drop/<name>-<version>.vsix` with `vsce`. For an extension that is an **app**, the `contribute`
method makes `package` depend on `build` explicitly, which also marks the plugin as that app's
deployer.

## Creating a project

```sh
smith add package a-ext --stack vscode-extension
```

```text
Wrote 10 file(s):
  .github/workflows/ci.yml
  package.json
  packages/a-ext/README.md
  packages/a-ext/package.json
  packages/a-ext/src/greeting/greet.use-case.spec.ts
  packages/a-ext/src/greeting/greet.use-case.ts
  packages/a-ext/src/greeting/greeting.contract.ts
  packages/a-ext/src/greeting/index.ts
  packages/a-ext/src/main.ts
  packages/a-ext/tsconfig.json
```

The manifest gets `publisher` (the workspace's npm scope without the `@`, else the workspace name),
`engines.vscode`, a `hello` command and `contributes`. The root gains `@types/vscode`,
`@vscode/vsce` and `esbuild`. `smith add lib --stack vscode-extension` fails with
`an extension is a package (published to the Marketplace) or an app: smith add package <name> --stack vscode-extension`.

## Releasing

| Kind | What a release does |
|---|---|
| package | the node plugin's `vsce publish` of the staged build to the Marketplace (`VSCE_PAT`), skipping a version it has |
| app | the release stamps its version into the manifest of the build, packs the `.vsix` into `dist/release/<app>/`, publishes it to the Marketplace and, when `OVSX_PAT` is set, to Open VSX. Both skip a version they already have. `[release] vsce_entra` signs in with Microsoft Entra ID instead of `VSCE_PAT` |

A native extension (a platform-specific build) is only tagged by the release job; each native job
publishes its own platform's build.

## CI

For an extension app, `ci` asks for `VSCE_PAT` (required) and `OVSX_PAT` (optional).

## Verified

smith 1.62.0, Windows 11, Node v24.12.0. Scaffolding; `smith run build` (esbuild, 222 ms) and
`smith run package -p a-ext` produced `dist/drop/a-ext-0.0.0.vsix` (6 files, 3.41 KB). **Not
verified:** publishing to the Marketplace or Open VSX (credentials), native builds, and the extension
**app** path (only a package was scaffolded). `smith doctor` also asks the Marketplace about the
published version (`vsce show`); without a login it warns that release tags were not checked.

## See also

- [The node plugin](node.md#releasing), [vitest](vitest.md), [eslint](eslint.md)
- [Releasing](../../../reference/releasing.md)
