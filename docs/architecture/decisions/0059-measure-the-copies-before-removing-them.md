# ADR 0059: Measure the copies before removing them

- Status: accepted (the measure); the removal is not built
- Date: 2026-11-01

## Context

Idea B2 of the design review: absence is the default. Per-package config files that are pure
derivation (tsconfig, jest config) would live once at the workspace and a project would hold a file
only when it differs.

## Challenge

A tool must find its config. `tsc`, `jest` and ESLint 9 look in the project folder or upward; the
editor's language server and many build tools look only at the project folder. So "zero files" is
true for some tools only. For the others the smallest honest form is a file of one line that
extends the shared base (`{ "extends": "../../tsconfig.base.json" }`). Which form each tool allows
is knowledge of the plugin, not of the core. smith must not delete a file whose removal a tool or an
editor then fails to see.

## Decision

1. The core measures. `smith copies` lists the files at project roots that three projects or more
   hold with the same content (line ends and trailing space ignored), the lines they cost and the
   lines they would cost once, and the projects that differ. It changes nothing. This is the
   measure that tells whether the removal is worth its risk.
2. The removal belongs to the plugin that knows the tool. A renderer (ADR 0046) declares that a
   file can be left out or reduced to an `extends` form when a base file exists. `smith derive` then
   renders it in that form and `smith adopt` and `smith convert` offer to replace identical copies,
   through the plan of ADR 0047, only for files a renderer names. Not built yet.

## Measure on a generated workspace

A workspace made by `smith new` with 22 Node packages from `smith add package pN --lang node`: two
files are copies in all 22 projects (`tsconfig.json`, 9 lines; `tsconfig.build.json`, 4 lines). They
cost 286 lines; once they would cost 13. Config at project roots is 1254 lines in all, and most of it
is `package.json` and `README.md`, which differ by name. A converted workspace with more config files
per package (jest, swc) has more to gain; run `smith copies` on it to see.

## Cost

A workspace with a shared base has a hidden coupling: a change to the base reaches every project.
That is the aim, and `smith affected` must treat the base as an input of each project that extends
it (it does for `tsconfig.base.json` through the plugin's inputs).

## Sources

- `apps/smith/copymeasure/`, `docs/architecture/derived-files-facts-and-siblings.md`
