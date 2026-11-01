# ADR 0054: Defaults have vintages

- Status: accepted
- Date: 2026-11-01

## Context

Idea B5 of the design review: smith improves its opinions over time. If a new smith changes a
default (a stack, a lint rule, a task setting), every workspace made earlier changes with it,
silently. The owner cannot upgrade smith without fear.

## Decision

A workspace records the defaults vintage it follows: `[workspace] defaults = "2026.11"` (year and
month). `smith new` and `smith convert`/`adopt` write the latest vintage. A newer smith keeps the
defaults of the recorded vintage. The rule for code: a default change ships as a new entry in
`workspace.DefaultsVintages`, and the code reads the vintage before it chooses the old or the new
value. The old value never changes in place.

- `smith doctor` reports "this workspace uses the defaults of X; the latest is Y" as a warning with
  the remedy.
- `smith upgrade --defaults` is the deliberate move: it prints a plan and records the latest
  vintage. It does nothing else, so plain `smith upgrade` never moves a vintage.
- A recorded vintage newer than the smith knows is a warning ("upgrade smith"), not an error.
- A workspace with no `defaults` key was made before vintages. It counts as behind.

## Cost

Every default change from now on needs a branch on the vintage and a test for both values. That is
the price of the guarantee. If no default ever changes, the cost is zero.

## Smallest first slice (done)

One vintage only, `2026.11`. The key, the validation, the doctor warning and `upgrade --defaults`
exist. No default branches on the vintage yet, so the move changes nothing but the record. The first
real use is the first default change after 2.0.

## Consequences

- smith.toml gains one line in a new workspace.
- Existing workspaces get a doctor warning until they run `smith upgrade --defaults`.

## Sources

- `apps/smith/workspace/defaults_vintage_policy.go`, `apps/smith/upgrade/adopt_defaults_use_case.go`,
  `apps/smith/doctor/defaults_vintage_finding_mapper.go`
