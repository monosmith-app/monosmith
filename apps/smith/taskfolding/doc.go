// Package taskfolding finds task overrides that several projects repeat and
// moves them into [[groups]], so smith.toml states them once. It is pure:
// smith convert uses it on a plan, and smith doctor on an existing smith.toml.
// The result resolves to exactly the same tasks as the input.
package taskfolding
