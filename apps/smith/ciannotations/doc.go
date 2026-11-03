// Package ciannotations reads the check-run annotations GitHub attaches to the
// latest run of a repository's default branch. It finds the one case a
// workflow cannot report itself: GitHub did not start a job (billing), so the
// run has no steps and no logs and the annotation is the only clue.
package ciannotations
