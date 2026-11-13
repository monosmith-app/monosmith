// Package lintfix is the lint check that closes a full conversion when smith
// owns the lint: in each Node project that has a lint task, ESLint fixes what
// it can, and the problems it cannot fix are counted per project. A project
// where ESLint cannot run at all is reported with the first line of its error
// and the remedy. The hidden `smith eslint-fix` runs it.
package lintfix
