package dependencies

import (
	"encoding/xml"
	"regexp"
	"strings"
)

// The elements of an MSBuild file that carry a NuGet package version:
// PackageReference in a project (or Directory.Build.props), and
// PackageVersion in Directory.Packages.props, where central package
// management keeps one version for the whole workspace.
const (
	ElementReference = "PackageReference"
	ElementVersion   = "PackageVersion"
)

// PackageVersion is one package and the version an MSBuild file gives it.
type PackageVersion struct {
	Element string
	// Name is the package id as written.
	Name string
	// Version is the version as written ("" when the file gives none, as a
	// reference does under central package management).
	Version string
}

type item struct {
	Include         string `xml:"Include,attr"`
	Update          string `xml:"Update,attr"`
	VersionAttr     string `xml:"Version,attr"`
	VersionChild    string `xml:"Version"`
	VersionOverride string `xml:"VersionOverride,attr"`
}

type msbuildFile struct {
	ItemGroups []struct {
		References []item `xml:"PackageReference"`
		Versions   []item `xml:"PackageVersion"`
	} `xml:"ItemGroup"`
}

// ParsePackages reads the packages and versions an MSBuild file declares. A
// reference that only overrides a central version, or has no version, is
// left out.
func ParsePackages(src []byte) ([]PackageVersion, error) {
	var f msbuildFile
	if err := xml.Unmarshal(src, &f); err != nil {
		return nil, err
	}
	var out []PackageVersion
	add := func(element string, items []item) {
		for _, it := range items {
			name := it.Include
			if name == "" {
				name = it.Update
			}
			version := strings.TrimSpace(it.VersionAttr)
			if version == "" {
				version = strings.TrimSpace(it.VersionChild)
			}
			if name == "" || version == "" {
				continue
			}
			out = append(out, PackageVersion{Element: element, Name: name, Version: version})
		}
	}
	for _, g := range f.ItemGroups {
		add(ElementReference, g.References)
		add(ElementVersion, g.Versions)
	}
	return out, nil
}

// plainVersion is a version sync and up may rewrite: a plain version, not a
// range ([1.0,2.0)), a floating one (1.*) or a property ($(Version)).
var plainVersion = regexp.MustCompile(`^\d+(\.\d+){0,3}(-[0-9A-Za-z.-]+)?$`)

// Rewritable reports whether sync and up may change a version.
func Rewritable(version string) bool { return plainVersion.MatchString(version) }
