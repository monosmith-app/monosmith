// Package declaredimports checks that a Go module declares what it imports.
// go.work merges the build lists of every module in the workspace, so a
// module can import a package that only another module requires and still
// build here; whoever fetches the module on its own gets "no required module
// provides package". The check reads the module's imports against its go.mod
// requires. The standard library, the module's own packages, the modules of
// go.work and the modules a replace directive names are never reported;
// versions are not compared, and a require nothing imports is not reported.
package declaredimports
