// Package cachepipeline puts the secrets of the remote cache into an Azure
// Pipelines pipeline: it finds the organisation, project and pipeline from the
// git remote, reads the pipeline's variables, and adds or replaces a secret
// variable. A secret value never appears on a command line, in the
// environment of a program, in a log line or in an error: it travels in a
// request body held in a temporary file that only the user can read, and the
// file is removed when the call ends. The az program is reached through a
// port, so tests use a fake and no test calls Azure.
package cachepipeline
