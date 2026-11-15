// Package smithreleases reads smith's own releases: the newest version on a
// channel, and one version's archive, checked against the minisign-signed
// checksums.txt and unpacked into <smith home>/versions/<version> (ADR 0101).
// It follows the same rules as the launchers on npm, PyPI and NuGet, so a
// smith they started updates itself where they look for it.
package smithreleases
