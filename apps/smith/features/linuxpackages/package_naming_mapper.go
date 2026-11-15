package linuxpackages

import (
	"fmt"
	"strings"
)

// Package is one platform's package: what it is called and what it holds.
type Package struct {
	Name, Version string
	// Arch is Go's (amd64, arm64).
	Arch                       string
	Summary, Homepage, License string
	Maintainer                 string
	// Binary is the command's name in /usr/bin.
	Binary string
}

var debArches = map[string]string{"amd64": "amd64", "arm64": "arm64"}
var rpmArches = map[string]string{"amd64": "x86_64", "arm64": "aarch64"}

// Supported reports whether a Go architecture has packages.
func Supported(arch string) bool {
	_, ok := debArches[arch]
	return ok
}

// FileName is a package's file name, in each format's convention:
// monosmith_1.0.0_amd64.deb and monosmith-1.0.0-1.x86_64.rpm.
func FileName(kind string, p Package) (string, error) {
	switch kind {
	case "deb":
		return fmt.Sprintf("%s_%s_%s.deb", p.Name, packageVersion(p.Version), debArches[p.Arch]), nil
	case "rpm":
		return fmt.Sprintf("%s-%s-%s.%s.rpm", p.Name, packageVersion(p.Version), rpmRelease, rpmArches[p.Arch]), nil
	}
	return "", fmt.Errorf("unknown Linux package %q (deb or rpm)", kind)
}

// rpmRelease is the package's own revision; a new version restarts it.
const rpmRelease = "1"

// packageVersion turns a semver prerelease (1.0.0-rc.1) into one dpkg and
// rpm order before its release: 1.0.0~rc.1. Build metadata is dropped.
func packageVersion(v string) string {
	v, _, _ = strings.Cut(v, "+")
	return strings.Replace(v, "-", "~", 1)
}
