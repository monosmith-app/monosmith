# ADR 0013: Vertical slices in every language's own naming

- Status: accepted
- Date: 2026-10-16 to 2026-10-17

## Context

smith checks that code is laid out by what it does (see
[Conventions](../../reference/conventions.md)). The check has to work in languages whose
file-naming conventions differ. Why vertical slices were chosen as the layout is not
recorded in this log.

## Options considered

Not recorded.

## Decision

The vertical-slices check uses the most common naming of each language:

| Language | Naming |
|---|---|
| Go, Python | `<name>_<role>` (snake_case) |
| Dart, Rust, Ruby | `<name>_<role>` (snake_case) |
| C#, Kotlin, Java, PHP | a PascalCase role suffix (`PlaceOrderUseCase.cs`, tests `PlaceOrderUseCaseTests.cs`) |
| C++ | either style, but one style per folder |

Node gets the same rules through the ESLint config.

## Consequences

`smith slices check` and the per-project `slice-check` task implement the rules;
`smith doctor` checks that Node projects load the slice rules
([#436](https://github.com/monosmith-app/monosmith/pull/436)). The repository's own layout
is described in [Slice architecture](../slice-architecture.md).

## Sources

- Issues [#87](https://github.com/monosmith-app/monosmith/issues/87),
  [#236](https://github.com/monosmith-app/monosmith/issues/236)
- [Conventions](../../reference/conventions.md)
