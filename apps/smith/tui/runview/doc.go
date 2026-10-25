// Package runview turns the output of a `smith run` child process into the
// state of a run: which tasks are queued, running, cached, passed, failed or
// skipped, how long each took, how far the run has got, and each task's log.
// It reads the lines smith prints (a run also has a JSON event stream, `--json-events`, which this does not use yet) and has no
// terminal in it, so it is tested with plain strings.
package runview
