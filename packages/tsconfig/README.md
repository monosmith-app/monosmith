# @monosmith/tsconfig

The TypeScript compiler options a MonoSmith monorepo starts from, kept in one
installed package so the file in your repo stays three lines long.

```json
{
  "extends": "@monosmith/tsconfig/node.json",
  "compilerOptions": { "rootDir": "src" },
  "include": ["src"]
}
```

| File | For |
|---|---|
| `base.json` | every TypeScript project: strict, ES2023, Node's module resolution |
| `node.json` | Node apps, libraries and packages (adds the `node` types) |
| `react.json` | browser code with React (DOM libs, `react-jsx`, bundler resolution, no emit) |

Only `compilerOptions` live here: `include`, `exclude` and `paths` resolve
relative to the file that declares them, so they stay in your project's own
tsconfig. The output folder is not here either: `smith build` passes
`--outDir` so the build goes to the workspace's one `dist/`.

Change anything by setting it in your tsconfig; your value wins.
