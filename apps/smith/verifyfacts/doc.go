// Package verifyfacts reads what smith verify can say about a project beyond
// pass or fail: how many tests passed and failed, how many lint issues were
// found, how large the build is and where it was written, where the test results
// are, and the coverage with the file it came from.
//
// The counts come from what the tools printed (the summary lines of jest,
// vitest, pytest, dotnet test, cargo test, go test, eslint, golangci-lint and
// ruff). The paths and sizes come from the disk. Nothing here runs a tool or turns
// anything on: a coverage or test-results file is shown only when it is already
// in place, and a number smith cannot read is left out, never guessed.
package verifyfacts
