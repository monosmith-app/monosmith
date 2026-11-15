// Package pluginpack turns a plugin directory into the archive `smith plugin
// install` reads: every file of the directory, the program its manifest names
// for each platform marked executable, and a checksums file, packed in a
// fixed order with fixed timestamps so the same directory always gives the
// same bytes. It never runs the plugin.
package pluginpack
