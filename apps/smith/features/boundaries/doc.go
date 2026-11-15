// Package boundaries checks which projects depend on which against the rules
// in smith.toml [boundaries] (and the built-in one: nothing depends on an
// app). It is `smith boundaries`; smith doctor reports the same violations
// and smith ci verify fails on them.
package boundaries
