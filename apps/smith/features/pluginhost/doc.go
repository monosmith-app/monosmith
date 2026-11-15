// Package pluginhost starts the workspace's plugins and routes requests to the
// right one. Built-in plugins ship inside the smith binary but still run
// out-of-process — smith starts itself as `smith plugin serve <name>` — so they
// speak exactly the protocol an external plugin does.
package pluginhost
