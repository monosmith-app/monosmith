package sbom

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ComponentsFromLock lists the packages a packages.lock.json resolved, for
// every target framework. Project references are the solution's own.
func ComponentsFromLock(lock []byte) ([]protocol.Component, error) {
	var doc struct {
		Dependencies map[string]map[string]struct {
			Type     string `json:"type"`
			Resolved string `json:"resolved"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal(lock, &doc); err != nil {
		return nil, err
	}
	var out []protocol.Component
	for _, framework := range doc.Dependencies {
		for name, d := range framework {
			if d.Type != "Project" && d.Resolved != "" {
				out = append(out, components.New("nuget", "", name, d.Resolved))
			}
		}
	}
	return components.Dedupe(out), nil
}

// ComponentsFromProject lists the PackageReference items of a .csproj,
// .fsproj or .vbproj, by the version it names (empty with central package
// management).
func ComponentsFromProject(project []byte) ([]protocol.Component, error) {
	dec := xml.NewDecoder(bytes.NewReader(project))
	var out []protocol.Component
	var open *struct{ name, version string }
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "PackageReference":
				open = &struct{ name, version string }{}
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "Include", "Update":
						open.name = a.Value
					case "Version":
						open.version = a.Value
					}
				}
			case "Version":
				var v string
				if open != nil && dec.DecodeElement(&v, &t) == nil {
					open.version = v
				}
			}
		case xml.EndElement:
			if t.Name.Local == "PackageReference" && open != nil {
				if open.name != "" {
					out = append(out, components.New("nuget", "", open.name, open.version))
				}
				open = nil
			}
		}
	}
	return components.Dedupe(out), nil
}
