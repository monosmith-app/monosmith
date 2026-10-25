# The astro plugin

The `astro` plugin scaffolds a documentation site with Astro Starlight: a splash page, guides and
reference with generated sidebars. It is the scaffold and the `build` toolchain of the `docs-site`
stack, which lints with [eslint](eslint.md).

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `scaffold`, `toolchain` |
| Provides (slots) | `build` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate` |
| Stack it publishes | `docs-site`: scaffold `astro`; toolchain `astro`, `eslint` |
| Project kinds | `app` |

## What it does

```json
"scripts": {
  "build": "astro check && astro build",
  "start": "astro dev",
  "lint": "eslint ."
}
```

`build` runs `astro check` and then `astro build` into `dist/<app dir>`; `start` runs `astro dev`.
The site has no `test` and no `typecheck` script, so `smith doctor` reports
`a-docs has no test task, so CI never checks it`. The root `package.json` gains `@astrojs/check`.

## Creating a project

```sh
smith add app a-docs --stack docs-site
```

```text
Wrote 10 file(s):
  apps/a-docs/.gitignore
  apps/a-docs/astro.config.mjs
  apps/a-docs/eslint.config.mjs
  apps/a-docs/package.json
  apps/a-docs/src/content.config.ts
  apps/a-docs/src/content/docs/guides/getting-started.md
  apps/a-docs/src/content/docs/index.mdx
  apps/a-docs/src/content/docs/reference/overview.md
  apps/a-docs/tsconfig.json
  package.json
```

## Settings, releasing, CI

No settings; no `publish` or `ci` method. A site is released by its tag.

## Verified

smith 1.62.0, Windows 11, Node v24.12.0. `smith run build -p a-docs` passed in 9.9 s: four pages
built into `dist/apps/a-docs` and a Pagefind search index. Astro printed two warnings: the `i18n`
content collection is empty and the sitemap integration needs the `site` option of
`astro.config.mjs`; neither fails the build.

## See also

- [vite](vite.md), [angular](angular.md), [docker](docker.md) (an nginx image for a static site)
