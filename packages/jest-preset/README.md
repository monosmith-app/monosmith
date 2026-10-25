# @monosmith/jest-preset

> The Jest preset every MonoSmith (`smith`) monorepo uses. **One line in a
> project's Jest config, and no helper files at the workspace root.**

```js
// jest.config.js (or .ts, .cjs)
module.exports = {
  preset: '@monosmith/jest-preset',
}
```

## Why this exists

`smith convert` takes Nx out of a workspace, and Nx provided three things Jest
needs: the preset values (`@nx/jest/preset`), a resolver that falls back to
TypeScript (`@nx/jest/plugins/resolver`) and, for React projects, a transformer
for the files a bundler would load as assets (`@nx/react/plugins/jest`). Convert
used to write those out as files at the workspace root (`jest.resolver.cjs`,
`jest.asset-transform.cjs`, a rewritten `jest.preset.js`). They looked like
clutter, so a tidy user deleted them and every package's tests broke (issue
#447). They are this package now: one devDependency, upgraded with `npm update`,
tested here.

## What the preset sets

| Key | Value |
|---|---|
| `testMatch` | `**/?(*.)+(spec\|test).?([mc])[jt]s?(x)` |
| `moduleFileExtensions` | `ts`, `js`, `mts`, `mjs`, `cts`, `cjs`, `html` |
| `moduleNameMapper` | `./x.js` resolves to `./x` (an ESM-style import of a TypeScript file) |
| `transform` | `ts-jest` over TypeScript, JavaScript and HTML, reading `<rootDir>/tsconfig.spec.json` |
| `resolver` | the TypeScript-aware resolver below, wired in by the preset |
| `testEnvironment` | `jsdom` (set `testEnvironment: 'node'` in a Node project) |
| `testEnvironmentOptions` | `customExportConditions: ['node', 'require', 'default']` |
| `modulePathIgnorePatterns` | `<rootDir>/dist/`, `<rootDir>/out-tsc/` |
| `coverageReporters` | `html` |

These are the values `@nx/jest/preset` provided, so a converted workspace keeps
testing the way it did. Whatever a project writes beside `preset` wins, and
Jest merges `moduleNameMapper` and `transform` instead of replacing them.

### The resolver

Jest's own resolution runs first. When it cannot find a module, TypeScript is
asked, reading the first of `tsconfig.spec.json`, `tsconfig.test.json`,
`tsconfig.jest.json` and `tsconfig.json` above the project's root. That finds:

- what a tsconfig `paths` entry names (`@lib/greet`), and
- a module that only has types: `jest.mock('vscode')` with just `@types/vscode`
  installed, which the mock replaces anyway.

When neither finds it, you get Jest's own `Cannot find module` error. TypeScript
is optional: without it the resolver is Jest's.

To use the resolver without the preset:
`resolver: require.resolve('@monosmith/jest-preset/resolver')`.

## React projects

```js
module.exports = {
  preset: '@monosmith/jest-preset/react',
}
```

The base preset plus what `@nx/react` added to a project's Jest config:

- `tsx` and `jsx` join the module file extensions;
- scripts go through `babel-jest` with `@babel/preset-env`,
  `@babel/preset-typescript` and `@babel/preset-react` (automatic runtime);
- every other file that is not JavaScript, TypeScript, CSS or JSON goes through
  the asset transform: an imported SVG is a module whose default export is its
  file name and whose `ReactComponent` is a component named as SVGR names it
  (`my-logo.svg` is `SvgMyLogo`); any other file is its file name.

The asset transform alone, for a config with its own transform table, is
`'@monosmith/jest-preset/asset-transform'`:

```js
transform: {
  '^(?!.*\\.(js|jsx|ts|tsx|css|json)$)': '@monosmith/jest-preset/asset-transform',
}
```

## Peer dependencies

Nothing is installed for you beyond Jest itself. Add what your project uses:

| Package | When |
|---|---|
| `jest` | always |
| `ts-jest`, `typescript` | the base preset (its `transform` and the resolver's fallback) |
| `jest-environment-jsdom` | the default `testEnvironment` (`jsdom`) |
| `babel-jest`, `@babel/preset-env`, `@babel/preset-typescript`, `@babel/preset-react` | the `/react` preset (use Babel 7 presets with `babel-jest` 30) |
| `@swc/jest` | optional: when your transform table uses it, the asset transform hands scripts to it instead of `babel-jest` |

## Layout

The package is plain CommonJS that Jest loads directly: nothing is compiled.
The files at the package root are the doors Jest and consumers use
(`jest-preset.cjs` is the name Jest looks for, `react/jest-preset.cjs` for the
sub-preset); the code is three small slices under `src/`: `preset` (the values),
`resolver` and `asset-transform`. The specs sit beside the code and the
integration spec runs the real Jest binary over throwaway projects that use the
preset.

## License

MIT
