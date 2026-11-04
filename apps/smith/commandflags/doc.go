// Package commandflags is the one place that names the flags smith's commands
// share. A command registers a shared concept (the project to act on, the
// language, the output file, "do not ask") through this package instead of
// choosing its own name, so the same idea is spelled the same way everywhere.
// The neighbouring spellings stay accepted as synonyms, so no script breaks.
// The decision is ADR 0084; vocabulary_validator_test.go in package commandline
// fails when a command invents another spelling for a concept listed here.
package commandflags
