// Package npmpropagation waits until the packages of one npm release can be
// fetched the way a person fetches them, before a dist-tag moves. npm needs
// minutes between a publish and the moment every edge answers for it, and a
// launcher whose platform packages are missing installs without an error and
// then fails.
package npmpropagation
