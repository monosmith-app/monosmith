# smith boundaries

`smith boundaries` checks every dependency between projects against the rules you wrote in
`smith.toml` under `[boundaries]`, and against one built-in rule: no project depends on an
app (an app is deployed, so nothing builds on top of it). It prints each dependency that
breaks a rule as the two projects (the one that depends, then the one it depends on), the
rule and what to do about it.

[`smith doctor`](doctor.md) reports the same findings, and [`smith ci verify`](ci.md#verify)
runs this check first and fails on one, so a pull request cannot add the dependency. The
rules are checked on the project graph, so a dependency through a shared schema (a client
generated from another project's OpenAPI file) counts like an import.

A workspace with no `[boundaries]` table changes only by one warning, and only if a project
depends on an app. Nothing fails until you ask for it with `kinds = true` or a rule of your own.

## When to use it

- To keep layers apart: `tag:ui` never depends on `tag:server`, an app never on another app.
- Before a pull request, to see the violations without running the rest of `verify`.
- In CI: `smith ci verify` runs it; no extra step is needed.

## Synopsis

```sh
smith boundaries [--json]
```

## Arguments and flags

<!-- smith:begin boundaries -->
Check which projects depend on which against your rules

Usage: `smith boundaries [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the violations as JSON |

<!-- smith:end boundaries -->

## The rules

```toml
[boundaries]
kinds = true                       # the built-in rule is an error (unset: a warning; false: off)

[[boundaries.rules]]
from = "tag:ui"                    # who the rule is about
deny = ["tag:server"]              # what they may not depend on
reason = "the browser must not pull in server code"

[[boundaries.rules]]
from = "kind:app"
allow = ["kind:lib", "kind:package"]   # the only things they may depend on

[[boundaries.exceptions]]          # one dependency that may stay, with the reason
from = "web"
to = "admin"
reason = "the shared login page, until it moves into a lib"

[projects.web]
tags = ["ui"]                      # tags for the rules, besides the ones the plugin reports
```

A project is named the way `--projects` names it: a name, a glob (`libs/*`, `lib-*`),
`kind:app`, `kind:lib`, `kind:package`, `lang:go` or `tag:ui`. `deny` and `allow` can be
used together; a dependency that matches `deny`, or that matches nothing in a non-empty
`allow`, breaks the rule. An exception needs a `reason`.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | No dependency breaks a rule that is an error (warnings may be printed). |
| 1 | At least one does, or there is no workspace. |

## Files and environment

Reads `smith.toml` and the project graph. Changes nothing.

## Examples

```sh
smith boundaries
```

```text
✖ web -> admin: web imports admin
    rule: projects matching "tag:ui" may not depend on "tag:server" (the browser must not pull in server code)
    → remove web's import of admin, or move what both need into a lib that neither depends back on; if the dependency is on purpose, add [[boundaries.exceptions]] with from = "web", to = "admin" and a reason to smith.toml

1 error(s), 0 warning(s)
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `[[boundaries.rules]] number 1 has no from` | Name the projects the rule is about. |
| `[[boundaries.exceptions]] number 1 (a -> b) needs a reason` | Say why the dependency may stay. |

## Related

[`smith doctor`](doctor.md), [`smith graph`](graph.md), [`smith ci`](ci.md),
[Configuration](../reference/configuration.md#boundaries).
