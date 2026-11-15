// Package pluginarchive owns the on-disk format of a packaged plugin: a zip
// whose root holds plugin.json, the plugin's files and checksums.sha256, a
// sha256sum-style list of every other file. It writes that zip the same way
// every time, and extracts one (or copies a plugin directory) into a fresh
// directory without ever running anything and without letting an entry write
// outside it. Packing and installing both depend on it, never on each other.
package pluginarchive
