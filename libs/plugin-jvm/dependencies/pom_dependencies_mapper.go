package dependencies

import (
	"encoding/xml"
	"regexp"
	"strings"
)

// Where a Maven dependency is declared.
const (
	SectionDependencies = "dependencies"
	SectionManagement   = "dependencyManagement"
	// SectionPropertyPrefix starts the section of a dependency whose version
	// is a ${property} the same pom defines: "property:junit.version". The
	// property is what changes.
	SectionPropertyPrefix = "property:"
)

// Coordinate is a dependency's groupId:artifactId and the version a build
// file gives it.
type Coordinate struct {
	Section string
	// Name is groupId:artifactId.
	Name string
	// Version is the version as written, or, for a ${property} the pom
	// defines, the property's value.
	Version string
}

type pomXML struct {
	Dependencies []pomDependency `xml:"dependencies>dependency"`
	Management   []pomDependency `xml:"dependencyManagement>dependencies>dependency"`
}

type pomDependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

var (
	propertiesBlock = regexp.MustCompile(`(?s)<properties>(.*?)</properties>`)
	propertyEntry   = regexp.MustCompile(`<([\w.\-]+)>([^<]*)</[\w.\-]+>`)
	propertyRef     = regexp.MustCompile(`^\$\{([\w.\-]+)\}$`)
)

// ParsePom reads the dependencies a pom.xml declares directly and in
// <dependencyManagement>, with the versions they give. A version that is a
// ${property} this pom defines becomes that property's value, in a section
// that names it; one it does not define (the parent's) stays as written.
// A dependency with no version takes the managed one and declares nothing.
func ParsePom(src []byte) ([]Coordinate, error) {
	var x pomXML
	if err := xml.Unmarshal(src, &x); err != nil {
		return nil, err
	}
	properties := map[string]string{}
	if m := propertiesBlock.FindSubmatch(src); m != nil {
		for _, e := range propertyEntry.FindAllSubmatch(m[1], -1) {
			properties[string(e[1])] = strings.TrimSpace(string(e[2]))
		}
	}
	var out []Coordinate
	add := func(section string, deps []pomDependency) {
		for _, d := range deps {
			version := strings.TrimSpace(d.Version)
			if version == "" || d.GroupID == "" || d.ArtifactID == "" {
				continue
			}
			c := Coordinate{Section: section, Name: strings.TrimSpace(d.GroupID) + ":" + strings.TrimSpace(d.ArtifactID), Version: version}
			if m := propertyRef.FindStringSubmatch(version); m != nil {
				if value, ok := properties[m[1]]; ok {
					c.Section, c.Version = SectionPropertyPrefix+m[1], value
				}
			}
			out = append(out, c)
		}
	}
	add(SectionDependencies, x.Dependencies)
	add(SectionManagement, x.Management)
	return out, nil
}

// plainVersion is a version sync and up may rewrite: dotted numbers with an
// optional qualifier (5.10.2, 2.0.0.RELEASE, 1.0-rc1); not a range, a
// ${property} or LATEST.
var plainVersion = regexp.MustCompile(`^\d+(\.\d+)*([-.][A-Za-z][\w.]*)?$`)

// Rewritable reports whether sync and up may change a version.
func Rewritable(version string) bool { return plainVersion.MatchString(version) }
