package discovery

import (
	"encoding/xml"
	"strings"
)

// Pom is what smith reads from a pom.xml.
type Pom struct {
	GroupID, ArtifactID, Version, Packaging string
	// Description is the pom's <description>.
	Description string
	// Kind and Stack are the smith.kind and smith.stack properties.
	Kind, Stack string
	// Dependencies are groupId:artifactId of every dependency.
	Dependencies []string
	// Modules are an aggregator's module directories.
	Modules []string
}

// Coordinates is groupId:artifactId, what other modules depend on.
func (p Pom) Coordinates() string { return p.GroupID + ":" + p.ArtifactID }

type pomXML struct {
	GroupID     string `xml:"groupId"`
	ArtifactID  string `xml:"artifactId"`
	Version     string `xml:"version"`
	Packaging   string `xml:"packaging"`
	Description string `xml:"description"`
	Parent      struct {
		GroupID string `xml:"groupId"`
		Version string `xml:"version"`
	} `xml:"parent"`
	Properties struct {
		Entries []struct {
			XMLName xml.Name
			Value   string `xml:",chardata"`
		} `xml:",any"`
	} `xml:"properties"`
	Dependencies []struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
	} `xml:"dependencies>dependency"`
	Modules []string `xml:"modules>module"`
}

// ParsePom reads a pom.xml; groupId and version fall back to the parent's.
func ParsePom(src []byte) (Pom, error) {
	var x pomXML
	if err := xml.Unmarshal(src, &x); err != nil {
		return Pom{}, err
	}
	p := Pom{GroupID: strings.TrimSpace(x.GroupID), ArtifactID: strings.TrimSpace(x.ArtifactID), Version: strings.TrimSpace(x.Version), Packaging: strings.TrimSpace(x.Packaging), Description: strings.TrimSpace(x.Description), Modules: x.Modules}
	if p.GroupID == "" {
		p.GroupID = strings.TrimSpace(x.Parent.GroupID)
	}
	if p.Version == "" {
		p.Version = strings.TrimSpace(x.Parent.Version)
	}
	if p.Packaging == "" {
		p.Packaging = "jar"
	}
	for _, e := range x.Properties.Entries {
		switch e.XMLName.Local {
		case "smith.kind":
			p.Kind = strings.TrimSpace(e.Value)
		case "smith.stack":
			p.Stack = strings.TrimSpace(e.Value)
		}
	}
	for _, d := range x.Dependencies {
		group := strings.TrimSpace(d.GroupID)
		if group == "${project.groupId}" {
			group = p.GroupID
		}
		p.Dependencies = append(p.Dependencies, group+":"+strings.TrimSpace(d.ArtifactID))
	}
	return p, nil
}
