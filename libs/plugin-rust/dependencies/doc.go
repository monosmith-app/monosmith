// Package dependencies lists and changes the crates a Cargo workspace
// declares: each crate's Cargo.toml and the root [workspace.dependencies],
// so `smith sync`, `smith up` and the one-version check cover Rust. Adding
// and removing crates goes through cargo itself.
package dependencies
