# ADR 0003: A content-addressed cache behind a Storage interface

- Status: accepted
- Date: 2026-10-15

## Context

The cache must be correct across machines and operating systems, and shareable later.

## Options considered

Bazel's remote-cache layout (action cache pointing at a content-addressed store) is the
model followed; other designs considered are not recorded. See
[research 03](../research/03-prior-art-and-stack.md).

## Decision

A task's key is a SHA-256 over a versioned, canonical encoding of everything that can
change its result, with no absolute paths or timestamps. Results are stored the way
Bazel's remote cache stores them: `ac/` entries pointing at `cas/` blobs, behind a
`Storage` interface. A local directory comes first; remote backends come later.

## Consequences

The remote cache (2026-10-16, [#195](https://github.com/monosmith-app/monosmith/pull/195))
slots in behind the local store: HTTP, S3 or Azure Blob, with every object signed with
HMAC-SHA256 over its kind, key and content. An object that fails verification, or an
unreachable store, is a miss with a warning; without the signing key the remote is off, so
nothing is shared unsigned. `smith why-miss` can explain a miss because each entry keeps
what its key was computed from. Since [#365](https://github.com/monosmith-app/monosmith/pull/365)
(2026-10-18) the installed toolchain version is part of the key.

## Sources

- [Research 03](../research/03-prior-art-and-stack.md)
- [Tasks, cache and affected](../../reference/tasks-and-cache.md),
  [Execution and cache](../execution-and-cache.md)
- Remote cache: pull request [#195](https://github.com/monosmith-app/monosmith/pull/195)
