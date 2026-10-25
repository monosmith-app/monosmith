# ADR 0001: Build from scratch in Go

- Status: accepted
- Date: 2026-10-15

## Context

The old mnci (TypeScript) was a thin CLI over Nx. Nx brings a Node runtime, a daemon
and its own plugin model. Please has no JavaScript, .NET or Dart support and no native
Windows.

## Options considered

- Keep building on Nx: rejected by the context above (a Node runtime and Nx's plugin model).
- Fork Nx or Please: rejected. Both are sources of ideas, not forks.
- Write a new tool in Go that is the engine itself: chosen.

## Decision

Write a new tool in Go that is the engine: discovery, task graph, hashing, cache,
scheduler, affected, release, TUI, doctor.

## Consequences

One static binary with no runtime to install. Everything Nx did for mnci has to be
rebuilt, and converting from Nx becomes a feature (`smith convert`). Native Windows is
first-class from the start.

## Sources

- Research: [current mnci inventory](../research/01-current-mnci-inventory.md),
  [Nx subsystems](../research/02-nx-subsystems.md),
  [prior art and stack](../research/03-prior-art-and-stack.md)
- Issue [#1](https://github.com/monosmith-app/monosmith/issues/1)
- [Technology](../technology.md), [Architecture](../architecture.md)
