// Package toolinstall finds the programs a workspace needs on the machine,
// says which are missing, and offers the ways to get them (ADR 0076). Plugins
// declare the tools as data (protocol.ToolNeed). This package detects them in
// parallel, reads the version a workspace pins, detects the package managers
// of the system and builds a plan of presented choices. It runs an install
// command only after the person chose it, prints the command first, and never
// runs one that needs administrator rights in silence.
package toolinstall
