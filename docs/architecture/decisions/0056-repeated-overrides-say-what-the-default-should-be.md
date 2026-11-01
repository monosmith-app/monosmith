# ADR 0056: Repeated overrides say what the default should be

- Status: accepted
- Date: 2026-11-01

## Context

Idea B3 of the design review: config should evolve from what people do. If five projects set the
same test timeout, the default is wrong for the workspace. If a default is overridden by nearly every
project, it is the wrong default.

## Decision

`smith doctor` reads `[projects.<name>.tasks.<task>]` and reports, as warnings with a remedy:

- Lift: the same task, field and value in at least three projects (the threshold is a constant,
  `overridelearn.DefaultThreshold`, until a person asks to change it). Skipped when `[tasks.<task>]`
  already has that value.
- Replace: a field the workspace default `[tasks.<task>]` sets, which at least three projects, and 80
  percent of all projects, set to another value.

It reports; it does not apply. The fix is a diff the person reads: a "move to the default" edit
deletes lines from many projects, and smith never rewrites a file a person edited without a plan.
An apply step can reuse the plan engine later.

## Cost

One more section in doctor output, so doctor must stay quiet when there is no pattern (it does). An
override repeated on purpose (three libs that really need `cache = false`) gets a warning; the
remedy names `[[task_groups]]` for that case. A knob to silence it is not built.

## Smallest first slice (done)

Task overrides in `smith.toml` only (command, cwd, inputs, outputs, depends_on, env by variable,
cache, verify, serve settings). Not done: settings of plugins (`[projects.<name>.settings]`), facts,
`tags`, and an apply step.

## Sources

- `apps/smith/overridelearn/`, `apps/smith/doctor/override_suggestions_finding_mapper.go`
