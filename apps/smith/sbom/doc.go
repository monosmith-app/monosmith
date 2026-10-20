// Package sbom writes a CycloneDX software bill of materials for a released
// app or package, built from its ecosystem's lockfile or manifest. Go (go.mod)
// and npm (package-lock.json, falling back to package.json) are supported; a
// project of another language gets no SBOM yet.
package sbom
