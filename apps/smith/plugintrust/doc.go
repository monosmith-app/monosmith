// Package plugintrust remembers which external plugins the user has decided to
// trust, and decides whether a plugin about to start still is trusted. A
// decision belongs to one workspace, one plugin name, the command line that
// starts it and the SHA-256 of the files that command runs; changing any of
// them makes the plugin untrusted again. The record lives in the user's smith
// home, never in the repository.
package plugintrust
