// Package releasepack makes a release's package files attachable to GitHub
// Releases (registry = "github-release"): it checksums them, finds out which
// files a release holds already, and writes the one bundle (release-bundle.json
// and INSTALL.md) that lists every package of a release run with its files,
// checksums and the commands that install the set together.
package releasepack
