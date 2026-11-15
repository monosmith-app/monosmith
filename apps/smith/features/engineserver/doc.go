// Package engineserver is `smith serve`: the engine as a JSON-RPC server on
// stdin and stdout, one JSON message per line, the framing `smith mcp` uses. A
// method per function (projects, graph, affected, run, cancel, doctor, release
// preview, explain, install, ...), and the run event stream as `event`
// notifications. It runs in the workspace it is started in and opens no port.
// The protocol is described in docs/reference/serve.md.
package engineserver
