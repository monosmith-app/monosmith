// Package runevents is the documented event stream of everything that takes
// time: a run of tasks, an install, a release. Every interface renders the same
// events: `smith run --json-events` prints them as JSON lines, and `smith serve`
// sends them as notifications. The shape is described in docs/reference/serve.md.
package runevents
