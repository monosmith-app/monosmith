// Package pluginusage counts the projects a built-in plugin discovers, so that
// switching a plugin off can say what disappears. It sits above pluginhost
// because counting needs the project graph, and the graph needs pluginhost.
package pluginusage
