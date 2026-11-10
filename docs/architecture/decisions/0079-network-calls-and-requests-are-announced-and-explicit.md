# ADR 0079: Network calls are announced, and an explicit request wins

- Status: accepted
- Date: 2026-11-02

## Context

A real release (Metil, Rust and Python on beta.21 and beta.22) and a first-run review showed four
places where smith decided for the person or acted without saying so.

1. The update notice asked the npm registry once a day without ever saying so, and followed only
   the `latest` tag, so a person on the next channel was told nothing.
2. `--bump p=major` on a 0.x package gave 0.3.0, because the "minor before 1.0" rule of inferred
   bumps also applied to a request.
3. The npm package of a binary launcher moved its `next` dist-tag at once, while its platform
   packages were not yet readable on the registry. Someone who installed in that window got a
   launcher with no binary.
4. `smith recipe new` could define a role that the recipe's users choose only through questions or
   an answers file.

## Decision

- The update notice makes one request, the dist-tags of `monosmith` on registry.npmjs.org, for the
  channel of the running version (a prerelease follows `next`, a stable release follows `latest`).
  The choice is `[workspace] update_check = "on" | "off"` or `SMITH_UPDATE_CHECK=on|off` (the
  user's setting wins). When nothing is chosen, the first run prints one line that describes the
  request and how to choose, and asks nothing for a day. The answer is cached 24 hours, the request
  times out after one second, and CI and non-terminals never ask. `smith --version` and
  `smith doctor` repeat the line while a newer version exists.
- A bump the person asks for (`--bump p=major`) is a request, not an inference: on a 0.x package it
  gives 1.0.0. A result that differs from the request says so in a note.
- Publishing a package whose `optionalDependencies` are pinned to the same version first reads each
  one back from the registry, for 60 seconds. If one is missing, the package takes the dist-tag
  `pending`, and the report names the `npm dist-tag add` command that moves the channel tag.
- `smith recipe new --chosen role=question:answer=plugin,answer=plugin` defines a role the recipe's
  users choose. A role is fixed or chosen, never both.

## Consequences

- A person who never chose sees one announcement line, then a request after a day. No request is
  made before it is announced.
- The `pending` tag keeps the channel tag on the last version that works, at the cost of a manual
  command when the registry is slow for more than a minute.
- The guided form and the VS Code command still ask only for fixed roles. They need a step type with
  repeated sub-questions (a role, then its question, then an answer and a plugin for each option).
  That change is not made here.
