// Package releasesign signs the checksums a release attaches, with minisign,
// so the launchers on npm, PyPI and NuGet run an archive only when its
// checksum is in a list the publisher signed (ADR 0101).
package releasesign
