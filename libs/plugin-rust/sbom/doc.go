// Package sbom lists the crates a Rust project ships for its release's SBOM:
// the packages in the nearest Cargo.lock, or the dependencies its Cargo.toml
// declares when there is no lockfile.
package sbom
