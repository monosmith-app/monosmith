// Package fullconversion drives the second half of `smith convert --level
// full`: once the conversion itself is planned, it works out what the old
// tooling leaves (on a scratch copy of the converted workspace, so nothing is
// touched before the user agrees), puts the choices, shows the removals, and,
// after the conversion is applied, asks each group of deletions in turn,
// applies the ones agreed, installs once and checks the workspace. The
// questions, the review and the steps are doctor fixes (see package
// doctorfix), so the interactive shell's wizard drives the same plan.
package fullconversion
