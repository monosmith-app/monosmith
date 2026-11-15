// Package nativebuild is `smith ci native`: projects that must be built on
// each OS (a Go app using cgo) are linted, tested, built and packaged on the
// runner, and on a release push each runner attaches its own platform's build
// to the app's GitHub Release.
package nativebuild
