// Package npmbinaries publishes an app's binaries as npm packages: one package
// per platform holding the binary, and wrapper packages whose bin (the shim)
// runs it. The platform packages go first, and the wrappers take their dist-tag
// only after every platform package can be fetched from the registry (ADR 0082).
// The shim itself is shim_template.js: when the platform package is missing it
// says whether the package is on the registry, not there yet, or cannot be
// asked, offers the fix, and installs nothing without a yes (SMITH_AUTO_FIX=1
// accepts it for CI). Its choices are tested by shim_template.test.js, which
// `go test` runs when node is installed.
package npmbinaries
