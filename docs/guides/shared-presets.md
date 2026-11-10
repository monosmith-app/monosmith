# Share your team's config as a preset package

A preset is a versioned package that holds your team's taste: an ESLint config, a Jest preset, a
`tsconfig` base. Workspaces depend on it. smith needs no feature for the package itself: you publish
it like any other package, and monosmith's own `@monosmith/eslint-config` and `@monosmith/jest-preset`
work this way.

## The pattern

1. Make a package that exports the config (`@acme/eslint-config` with `index.js`, `@acme/tsconfig`
   with `base.json`). Release it with smith like any package.
2. In each workspace, depend on it and extend it: `extends: "@acme/tsconfig/base.json"`, `preset:
   "@acme/jest-preset"`.
3. Tell smith it is a preset, in `smith.toml`:

   ```toml
   [dependencies]
   presets = ["@acme/eslint-config", "@acme/tsconfig", "@acme/jest-preset"]
   ```

4. When the platform team releases a new version, run `smith up` in each workspace. The presets are
   shown first, as one group (`preset`), before patch, minor and major updates. Every project that
   uses a preset is bumped in the same change, and the workspace is installed again.

   ```text
   preset (your team's shared config, bumped in every project that uses it)
     @acme/eslint-config          2.1.0 → 2.2.0  (root), api, web
   patch
     zod                          3.23.0 → 3.23.1  api
   ```

Use `smith try up @acme/eslint-config` to see the result of a bump in a copy first
([`smith try`](../manual/try.md)).

## What smith does not do (yet)

- It does not open the upgrade pull request in every repository when the preset changes. Run
  `smith up --yes` from a scheduled job and open the pull request with your CI.
- It does not scaffold a preset package (`smith preset new`). A preset is an ordinary package, so
  `smith add package eslint-config --lang node` makes the folder.

See [ADR 0060](../architecture/decisions/0060-a-shared-preset-is-a-package-smith-up-groups.md).
