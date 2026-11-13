// Package lintfix is what `smith lint --fix` reports: for each project, the
// problems its linter fixed and the problems left, as one table with a closing
// line. It judges ESLint's JSON report (the Node path of the fix) and names the
// remedy of a linter that could not run. The task runner runs the fixes the
// plugins declare (protocol.TaskFix); this package only reads and reports them.
package lintfix
