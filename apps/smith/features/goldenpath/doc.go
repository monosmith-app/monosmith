// Package goldenpath holds the end-to-end test of smith's promise: create a
// monorepo, add an app, an internal lib and a publishable package in every
// built-in language, and lint, type-check, test, build, run and release them
// with no configuration written by hand. It needs network access, Node, Go and
// Python, and runs only with the e2e build tag:
//
//	go test -tags e2e -timeout 40m ./goldenpath
package goldenpath
