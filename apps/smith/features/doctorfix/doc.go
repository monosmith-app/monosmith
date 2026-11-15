// Package doctorfix turns the fixes doctor's findings carry into a plan and
// applies it: it folds the same fix on many projects into one, resolves a
// fix's questions (choices) into the files and steps the answers add, shows a
// fix as diffs and commands, writes its files, and runs its steps, stopping at
// the first that fails. The command line walks a plan with Walk; the
// interactive shell's wizard asks the same questions with Resolve.
package doctorfix
