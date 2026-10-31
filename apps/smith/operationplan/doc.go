// Package operationplan is the plan a mutating command prints before it
// writes: what it creates, changes and removes, grouped by what each feature
// adds, with the owner (smith or you) where smith.lock knows it. It also
// decides when a person is asked to confirm: in a terminal, outside CI, and
// without --yes. A run with no terminal, or in CI, never waits. See ADR 0047.
package operationplan
