// Package tooldefaults gives a language plugin the two things it needs to run
// a linter with smith's own defaults: finding out whether the repository
// brought its own configuration (which then wins), and materialising the
// plugin's embedded default under .smith/defaults/ so it can be passed to the
// tool on the command line.
package tooldefaults
