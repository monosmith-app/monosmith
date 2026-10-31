# smith-v2 integration QA

Date: 2026-10-31. Branch: `smith-v2` at the start of the run. Epic: [#819](https://github.com/monosmith-app/monosmith/issues/819).

The run treats smith-v2 as one product, as a first-time user would. Smith was built from the
checkout (`go build` of `apps/smith`) and every scenario used that binary. All work ran in scratch
directories outside the repository, on Windows 11 with Node 24 and Go 1.27. Nothing was published.
The release scenarios used a closed local port as the registry.

## Result

| # | Scenario | Result | Issue or commit |
|---|---|---|---|
| 1 | `smith new`; `add app web --with react,vitest,bootstrap`; `add app admin --with vue,jest,tailwind`; `add lib x`; `add package y`; build, test, lint, doctor, `ci verify`, choices, stacks, `explain --facts` | pass | none |
| 2 | `module.aliases` fact reaches tsconfig, Jest, Vite and Vitest on new projects | fail, fixed | `fix(node)`: Vite config used double quotes and a misaligned `build` key, so lint failed |
| 3 | `[derive] enabled`: restore, drift, `--check`, `--adopt`, `--eject`, `--replace`, doctor, `ci verify` | pass with 2 findings | `fix(derive)`: wrong advice on drift; [#840](https://github.com/monosmith-app/monosmith/issues/840): a smith-made workspace has nothing to derive |
| 4 | Converted-style layout, 7 packages (`@swc/jest`, esbuild, `exports` to dist): `add package z`, `--fresh` | pass | none |
| 5 | `release --dry-run`, `--rehearse`, API-surface warning, `api_check = "error"` | pass | none |
| 6 | `smith ci generate` (GitHub) with derive on and choices in use | pass | none |
| 7 | Windows | pass, notes below | none |
| 8 | `smith plugin search` and `install` with an unreachable, malformed or insecure marketplace | pass | none |
| 9 | Repository e2e (`go test -tags e2e`): see the section below | see below | none |

## Scenario log

### 1. Mixed Node workspace

```
smith new s1 --yes --language node
smith add app web --lang node --with react,vitest,bootstrap     -> Verified: build and test passed
smith add app admin --lang node --with vue,jest,tailwind       -> Verified: build and test passed
smith add lib x --lang node ; smith add package y --lang node  -> Verified
smith build | test | lint | doctor | ci verify                  -> all green (doctor: 34 passed, 0 failed)
```

`smith ci verify`: `4 projects checked, 20 of 20 tasks ok, 4 tests passed`. `smith choices` lists
framework, library, lint, style and test with the defaults. `smith stacks` shows
`react-vitest-bootstrap` and `vue-jest-tailwind` recorded in `smith.toml` with the project that uses
each. `smith explain web --facts` says `web has no shared facts` (the flag needs a project name;
`smith explain --facts` alone prints `smith explain needs a project name`, which is clear).

Wrong combinations stop with a sentence that names the fix, before any file is written:

| Command | Message |
|---|---|
| `--with vue,chakra` | `chakra requires react, which is not in the selection` |
| `--with react,jest,vitest` | `jest and vitest are both on the test axis: keep one` |
| `--with nope` | `no choice is named "nope" (smith choices lists them)` |
| `add lib c4 --with react` | `a react project is an app: smith add app c4 --stack react` (it first prints `Creating lib c4`) |
| `add app --lang python --with react` | `--lang python, but the choices react are for node projects` |
| `--with react,bootstrap,tailwind` | `bootstrap and tailwind are both on the style axis: keep one` |
| `--with react --stack node-vitest` | `--with names the choices and --stack names a stack: use one or the other` |

Observation: `smith new --language node` writes `tools/smith/install.sh`, and the first `smith add`
deletes it, because a Node workspace installs smith from npm. The file shows as `D` in `git status`
after the first add. Harmless.

### 2. The `module.aliases` fact

```toml
[facts]
"module.aliases" = { "@app" = "src" }
```

`smith explain web --facts` shows `module.aliases {"@app":"src"} (workspace, from smith.toml [facts])`.
New projects got `paths` in `tsconfig.json`, `moduleNameMapper` in the Jest config and
`resolve.alias` in `vite.config.ts`. A spec that imports `@app/greeting/...` passes under Jest and
Vitest.

Defect: the first `smith ci verify` failed in `shop:lint`:

```
vite.config.ts  9:7   error  Strings must use singlequote              @stylistic/quotes
                12:8  error  Extra space before value for key 'build'  @stylistic/key-spacing
```

The alias block was written with double quotes, and the `build:` key kept the extra space that
aligns it with a `plugins:` line, which is no longer the neighbour. Fixed in
`libs/plugin-node/modulealiases` and the Vite template, with tests. After the fix a new
React+Vitest app, a Vue+Jest app and a lib all pass `smith ci verify`.

Note: an import of the form `'@app/x.ts'` fails `tsc` in a lib (TS2877, a `.ts` extension on a
non-relative path). That is a TypeScript rule and not a smith defect.

### 3. Derived files

Setup for a smith-made workspace: `[derive] enabled = true`, then `smith derive`:
`0 derived files: 0 changed.` and no `smith.lock`. Smith writes neither `tsconfig.lib.json` nor a
Jest config with markers, so nothing is derivable (finding [#840](https://github.com/monosmith-app/monosmith/issues/840)).

With a hand-made `libs/b/tsconfig.lib.json` and a `jest.config.cjs` with the markers, and `b`
depending on `z`:

| Step | Result |
|---|---|
| `smith derive` | wrote the `references` key and `displayName: '@s2/b'`; `smith.lock` with two entries |
| delete `tsconfig.lib.json`, `smith derive` | file written again, but only `{"references": [...]}` (the other keys are not owned by the renderer; part of #840) |
| hand edit, `smith derive` | `drift` with a diff; exit 1; file untouched |
| `smith derive --check` | exit 1 (`1 derived files are out of date`) |
| `smith doctor` | one failing check naming the file, the state and the next step |
| `smith ci verify` | stops first with the drift, exit 1 |
| `--explain`, `--adopt` | owner, renderer, mode printed; `--check` green after adopt |
| `--eject` | owner `you` in `smith.lock`; `--check` and doctor green |
| `--replace` on a current file | no change |

Defect: the advice after `--check` or `ci verify` found a hand edit was `run smith derive, and
commit the files`. `smith derive` holds a drifted file back, so the advice led in a circle. Fixed:
for drift, diverged and broken files the remedy names `--explain` and `--replace`, `--adopt` or
`--eject`.

### 4. Converted-style layout

Seven packages under `packages/*`, each with an esbuild `build`, `@swc/jest`, `main` and `exports`
pointing into `dist`, and three tsconfig files. `smith add package z`:

* no question was asked;
* `z` builds with esbuild plus a `tsc` declaration pass, tests with `@swc/jest`, and points `main`
  at `dist`;
* two warnings said the stack does not follow `scripts` and `tsconfig.files` (the sibling set is
  `tsconfig.json, tsconfig.lib.json, tsconfig.spec.json`; z has `tsconfig.build.json, tsconfig.json`).
  This is as ADR 0041 describes;
* `smith add package fresh --fresh` gave the default (`tsc` build, `ts-jest`).

`smith doctor` stayed red only for the hand-made fixture (the seven siblings write `packages/*/dist`
without declared outputs, and declare dev tools per package). Nothing named `z`. One difference to
note: the siblings point `main` at `./dist/index.js` inside the package, `z` points at the
workspace `dist/packages/z`. Both count as `exports.target = dist`.

### 5. Release

A git repository with tags `api@0.1.0` and `api@0.2.0`, then `fix(api): tidy the exports` removing
the export `bye`:

```
smith release --dry-run
  api  0.2.0 -> 0.2.1  (1 releasing commit(s))
  warning: api: the commits release a patch (0.2.1), but the public API lost 1 export (bye);
  that needs a minor (0.3.0). Mark a commit as breaking (`feat!:` ...), or restore the export.
```

`--json` carries the text in `apiWarnings`. With `[release] api_check = "error"` the command exits
1 with `cannot release: the commits ask for a smaller bump than the public API change needs`.
`--rehearse` against `https://127.0.0.1:1/` printed `unreachable`, the fix to apply, and exit 1; `--json` printed the same. `--dry-run` with the same registry
prints a warning that the tags could not be checked and still plans the release. An `http://`
registry in `smith.toml` is refused with `must be an https URL`, so a local mock needs TLS.

### 6. CI generation

`smith ci generate` on the derive-enabled workspace with two choice-based stacks prints
`pipelines are up to date`. The workflow installs smith from npm, then runs `smith ci setup`,
`audit`, `verify` (which includes `smith derive --check` and the boundary check), a release preview
on pull requests, and pack and release on `main`. Nothing in it depends on the choices.

### 7. Windows

* Generated files use LF; `smith.lock` paths use forward slashes.
* A Windows path in `SMITH_MARKETPLACE` is accepted. In Git Bash, `/nonexistent/idx.json` is
  rewritten by MSYS to `C:/Program Files/Git/...`, which only looks odd in the message.
* Connection errors carry the Winsock text (`connectex: No connection could be made...`) and still
  end with the remedy line.
* npm install runs print `1 package wants a newer Node than this one`; this comes from npm and does
  not stop the run.

### 8. Marketplace

| Marketplace | Message | Exit |
|---|---|---|
| `https://127.0.0.1:1/index.json` | `cannot reach the marketplace at ...` then `check your network, or set SMITH_MARKETPLACE` | 1 |
| `http://example.com/x.json` | `only https URLs are accepted (http only for localhost)` | 1 |
| a missing file | `cannot read the marketplace index` then `check SMITH_MARKETPLACE and [plugins] marketplace` | 1 |
| a file with bad JSON | `the marketplace index is not valid JSON` | 1 |
| `plugin install foo`, `plugin search --json` (unreachable) | same message as search | 1 |
| the default index (empty today) | `No plugin matches in <url>` | 0 |

No panic and no stack trace in any case.

## Repository checks

`smith slices check --lang go` passes for `apps/smith`, `libs/plugin-node` and
`libs/plugin-node-toolchains` (it lists the 12-file review notes that were already there).

### E2E tests

`go test -tags e2e` in `apps/smith/goldenpath`, run in two batches:

| Test | Result |
|---|---|
| `TestDeriveRegeneratesTheRepeatedConfigOfAConvertedWorkspace`, `TestAddFollowsTheDialectOfSiblings`, `TestWebChoices`, `TestAddNodeLibInConvertedLayout` | pass (221 s) |
| `TestWebWorkspace` | fail on this machine only: the Angular CLI needs Node 24.15 or newer and the machine has 24.12 (`angular:build exited with code 3`). Not a smith defect |

### `smith ci verify` of this repository

126 of 135 tasks passed. Failures: `eslint-config:test` (64 tests), `plugin-sdk-dotnet:test`, a race
in `jest-preset:lint` (its tests create `.fixture-*` folders in the package while lint walks it),
and two environment gaps (no gradle, no `.venv` in the checkout). Filed as
[#841](https://github.com/monosmith-app/monosmith/issues/841). `smith:test`, `smith:lint` and the
lint and tests of `plugin-node`, `plugin-node-toolchains` and `plugin-sdk` passed.

## Open issues

* [#840](https://github.com/monosmith-app/monosmith/issues/840): a smith-made workspace has nothing for `smith derive` to derive; a deleted keys-mode file returns as a stub.
* [#841](https://github.com/monosmith-app/monosmith/issues/841): failures of `smith ci verify` on this repository on Windows.
