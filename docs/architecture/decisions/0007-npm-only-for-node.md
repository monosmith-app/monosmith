# ADR 0007: npm only for Node

- Status: accepted (clarified 2026-10-18 and 2026-10-19)
- Date: 2026-10-15

## Context

The Node plugin needs one package manager to drive workspaces, installs and lockfiles.
Rationale for choosing npm over pnpm or yarn is not recorded.

## Options considered

- Support pnpm and yarn as well as npm: not chosen. The 2026-10-18 clarification says it
  is not a gap to close, because people do not call the package manager themselves.
- Refuse workspaces that use pnpm or yarn: not chosen; they are converted (below).

## Decision

The Node plugin runs npm workspaces. pnpm and yarn are not supported, and `smith convert`
reports them as such.

Clarified 2026-10-18: `smith install` and `smith uninstall` find the project's ecosystem
and run the right tool through its plugin, so not supporting pnpm or yarn does not make
smith harder to use. A workspace coming from another tool (Turborepo, a pnpm or yarn Nx
workspace) is converted to npm workspaces instead of being refused.

Decided 2026-10-19 ([#353](https://github.com/monosmith-app/monosmith/issues/353)):
`smith install` (`smith i`) is smith's own install command, independent of each language's
tooling; for Node it runs npm. A converted Turborepo workspace uses `smith i` instead of
pnpm or yarn. `smith convert` turns pnpm and yarn workspaces into npm workspaces and
reports what changes or goes, and smith does not grow pnpm or yarn support.

## Consequences

Resolved versions are not carried over from `pnpm-lock.yaml` or `yarn.lock`: `npm install`
resolves every range again, so a dependency can move inside its range.
See [Converting](../../getting-started/converting.md#converting-from-turborepo).

## Sources

- Issue [#353](https://github.com/monosmith-app/monosmith/issues/353); pull request
  [#356](https://github.com/monosmith-app/monosmith/pull/356) (docs: why pnpm and yarn
  need no support); pull request [#379](https://github.com/monosmith-app/monosmith/pull/379)
  (convert Turborepo workspaces)
