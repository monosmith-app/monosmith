# ADR 0028: Graph edges from shared schemas, and a blast-radius query

- Status: accepted
- Date: 2026-10-20

## Context

smith's project graph links projects by the imports their language plugin reports
([project graph](../architecture.md): projects from `discover`, edges from what each project
imports matched to what others export). A JVM service that exports an OpenAPI spec, with a Go
client and a TypeScript frontend generated from it, therefore did not mark the clients as
affected when the schema changed.

## Options considered

The issue proposed both declared schemas (a setting with globs) and discovery that proposes
consumers by reading common references. Only the first shipped.

## Decision

A project owns the OpenAPI, AsyncAPI, protobuf and GraphQL files in its directory (plus
`schemas` globs) and consumers declare `consumes_schemas`. Consumers depend on the owner,
so `smith affected` and `smith run --affected` rebuild generated clients.
`smith graph --blast-radius <project|file>` lists everything a change reaches, across
languages.

## Consequences

Not built: schema-aware discovery that proposes consumers (for example reading generator
configs or proto imports), and parsing GraphQL and AsyncAPI references. Schema-driven code
generation as a cached `generate` task came the same day
([#392](https://github.com/monosmith-app/monosmith/pull/392)).

## Sources

- Issue [#343](https://github.com/monosmith-app/monosmith/issues/343); pull request
  [#391](https://github.com/monosmith-app/monosmith/pull/391)
- [Tasks, cache and affected](../../reference/tasks-and-cache.md), [`smith graph`](../../manual/graph.md)
