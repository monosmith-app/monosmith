# ADR 0100: Apps keep their slices in features, integrations and shared, and a file's role limits its imports

- Status: accepted
- Date: 2026-11-15

## Context

The slice rules said what a file is called and how slices reach each other, but not where a
technology adapter lives. With one adapter per feature, an app with twenty external systems had
twenty adapters among its business slices, and nothing kept a vendor SDK, a database driver or an
HTTP client out of a policy. The owner's reference layout (JATO.Auto.Quotes.Manager, its ADR
`README.VERTICAL-FEATURE-SLICES.md`) splits an app's `src` into three zones and limits imports by
role. That project is half migrated: the new checks found integrations mixed into its features.
The skill `vertical-feature-slices` (v3) says the same. Owner decisions of 2026-11-15:

- Zones are required for apps. A library is zoned when it has one zone folder; otherwise its package
  is its capability and its slices sit directly under `src`.
- `shared` and `integrations` never import `features`; they may use each other.
- A vendor package listed by the project is imported only in `integrations`.
- The role limits are enforced, in every language.
- `data` and `manager` are banned only as a whole name (`user-data`, `account-manager` are honest).
- A front end's chrome is a feature (`features/app-chrome`), its configuration is shared.
- smith's own apps follow the rules: each is moved into zones, move only, after a reviewed move list.

## Decision

1. **Zones.** An app's `src` (Go: the module root; Kotlin: the package root) holds `features/`,
   `integrations/` and `shared/`, and every slice is in one of them. `integrations/<system>` owns one
   external system: its client, its payloads, its vocabulary. An adapter one feature uses stays in that
   feature as a `.client` until a second feature needs it. `shared/<topic>` is a named slice with no
   business knowledge. The editor or UI framework (VS Code, IntelliJ, React) is not an integration.
2. **Direction.** `shared` and `integrations` never import `features`, by relative path or alias,
   tests included.
3. **Vendors.** The project lists its vendor packages (ESLint `verticalSlices.vendors`; smith.toml
   `[conventions.vertical-slices] vendors` for the other languages). Importing one outside
   `integrations` is reported; tests and `testing` folders are exempt.
4. **Roles.** A `.model`, `.policy`, `.validator`, `.error` or `.enum` imports no integration (not even
   its types) and no `.client`, `.repository` or `.store`. A `.handler`, `.mapper`, `.algorithm` or
   `.contract` may use an integration's types, never its code. In languages that cannot tell a type
   import from a code import, the type allowance does not apply.
5. **Enforcement.** TypeScript: the `vertical-slices/zone-import` rule and `file-role` in
   `@monosmith/eslint-config`. Other languages: `smith slices check` (the `slice-check` task), reading
   imports with `libs/plugin-architectures/imports`.

## Consequences

- Every app built on the old layout fails its lint until it is sorted into zones. smith's own
  `apps/vscode`, `apps/smith` and `apps/jetbrains` were moved in this change series.
- A contract may carry an integration's type (a state that holds a session), which the first draft
  forbade; a model may not.
- Grouping smith's 141 feature packages into capabilities (`features/adopt/{decision,flow,…}`) renames
  packages and is left for later.
