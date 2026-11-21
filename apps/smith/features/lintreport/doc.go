// Package lintreport runs the lint of a project that smith adopt or smith
// import brought in, and turns what failed into items of the "Left to do"
// list: how many problems the linter found, how many its --fix repairs, the
// rules found most, and for the vertical-slice check the ways on. It never
// runs a fix: the list names the command, and the person decides.
package lintreport
