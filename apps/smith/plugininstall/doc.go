// Package plugininstall puts a packaged plugin on the plugin search path: it
// fetches the archive (a local file or directory, an https URL, or a GitHub
// release asset), checks its digests, extracts it into a staging directory
// beside the destination and only then swaps it in. It never runs the plugin;
// whether smith may start it later is plugintrust's decision, and recording a
// pin is the caller's step.
package plugininstall
