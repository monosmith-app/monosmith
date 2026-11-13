# @monosmith/tsconfig

The TypeScript compiler options a MonoSmith monorepo starts from, kept in one
installed package so the file in your repo stays three lines long.

The workspace's root `tsconfig.base.json` extends it, and every project extends the base, so
one edit of the base reaches every project:

```json
{ "extends": "@monosmith/tsconfig/node.json" }
```

```json
{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": { "rootDir": "src" },
  "include": ["src"]
}
```

Code that a bundler builds (a browser app, a VS Code extension) extends the root
`tsconfig.browser.json`, which extends the base and adds the DOM libraries, bundler resolution
and no emit. There is no `react.json` here: a preset that a project extends beside the base
would override the base. The `jsx` option is set by the React project itself.

| File | For |
|---|---|
| `base.json` | every TypeScript project: strict, ES2023, Node's module resolution |
| `node.json` | Node apps, libraries and packages (adds the `node` types) |

Only `compilerOptions` live here: `include`, `exclude` and `paths` resolve
relative to the file that declares them, so they stay in your project's own
tsconfig. The output folder is not here either: `smith build` passes
`--outDir` so the build goes to the workspace's one `dist/`.

Change anything by setting it in your tsconfig; your value wins.
