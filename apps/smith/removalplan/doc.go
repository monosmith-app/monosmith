// Package removalplan turns the choices of a full conversion into a plan of
// removals: one group of deletions per retired tool (lint, the test runner
// not chosen, the old build, Nx and mnci), each with the reason that follows
// from a choice. A group is a doctor fix, so the same plan can be shown as
// diffs, asked one group at a time, applied, and read as JSON by a wizard.
// Each group is computed on the workspace as the groups before it leave it,
// and a package is only taken out of a package.json when no file that stays
// imports or runs it.
package removalplan
