// Package sbom writes a CycloneDX 1.5 software bill of materials for a released
// app. The project's plugin lists its components from the ecosystem's lockfile
// or manifest (the sbom/components method of the plugin protocol); this package
// renders and writes the document. A project whose plugin cannot list
// components gets no SBOM; one that declares no dependency (a native app with
// no vcpkg or Conan manifest) gets an explicitly empty one.
package sbom
