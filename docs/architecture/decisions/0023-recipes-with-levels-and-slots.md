# ADR 0023: Recipes with levels and slots

- Status: accepted
- Date: 2026-10-18

## Context

Four things make a polyglot monorepo hard in every tool: many languages in one workspace,
the knowledge needed to build and release, conventions that change after projects exist, and
teams that differ ([Recipes and CI](../recipes-and-ci.md), "The goal"). One opinionated setup
does not fit everyone, but a hundred options are as bad as none.

## Options considered

From the design document, which marks each item decided, proposed or open:

- A recipe has at most one parent (`extends`), like a Dockerfile's single `FROM`. Decided;
  the reason given is that order is always parent first, then child, later wins.
- CI settings are recipe settings, not a separate wizard that writes YAML. Recommended there.
- CI providers stay in the core, with a renderer role later only if needed. Recommended.
- Plugin levels are exclusive (language, each task slot, scaffold) or additive (structure,
  deploy, conventions); rules and configs are versioned packages, not plugins. Decided.

## Decision

Recipes are pure data (TOML) describing which plugins a project uses and their settings, with
choices asked at `smith add` and declarative `when` blocks. Plugins declare their level,
slots (`provides`), requirements and conflicts in `initialize`; `pluginhost.ValidateMembers`
is the one place that checks a project's plugins, whatever chose them. A project records the
recipe it came from, so `smith doctor`, `smith upgrade` and `smith explain` follow it.
`smith recipe check` validates a recipe against the installed plugins and `smith recipe new`
is the creation wizard; recipes can be released like packages.

## Consequences

An error names both sides of a violated requirement or conflict. A recipe cannot yet carry
workspace-scope settings such as the dependency policy (noted in
[#386](https://github.com/monosmith-app/monosmith/pull/386)). Plugin discovery and
installation by name remain open ([index](README.md#open-not-decided-yet)).

## Sources

- Epic [#312](https://github.com/monosmith-app/monosmith/issues/312) (closed 2026-10-18)
  with children #306 to #311
- Pull requests [#316](https://github.com/monosmith-app/monosmith/pull/316) (levels),
  [#321](https://github.com/monosmith-app/monosmith/pull/321) (resolver),
  [#327](https://github.com/monosmith-app/monosmith/pull/327) (create projects from a recipe),
  [#330](https://github.com/monosmith-app/monosmith/pull/330),
  [#339](https://github.com/monosmith-app/monosmith/pull/339),
  [#317](https://github.com/monosmith-app/monosmith/pull/317) (named private registries)
- [Recipes](../../reference/recipes.md), [Recipes and CI](../recipes-and-ci.md),
  [`smith recipe`](../../manual/recipe.md)
