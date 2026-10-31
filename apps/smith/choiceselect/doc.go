// Package choiceselect checks a selection of choices (smith add --with react,vitest,bootstrap)
// against the choices the loaded plugins offer, and turns it into a stack. A choice is one
// option on an axis (framework, test, style). The solver is pure and knows no language: it
// compares axes, names and the requires and conflicts lists. See ADR 0045.
package choiceselect
