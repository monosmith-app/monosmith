package sbom

import (
	"encoding/json"
	"time"
)

type bomComponent struct {
	Type    string `json:"type"`
	BOMRef  string `json:"bom-ref,omitempty"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	PURL    string `json:"purl,omitempty"`
}

// RenderCycloneDX renders a CycloneDX 1.5 JSON document for the subject and
// its components.
func RenderCycloneDX(sub Subject, comps []Component, at time.Time, smithVersion string) ([]byte, error) {
	list := make([]bomComponent, 0, len(comps))
	for _, c := range comps {
		list = append(list, bomComponent{Type: "library", BOMRef: c.PURL, Name: c.Name, Version: c.Version, PURL: c.PURL})
	}
	meta := map[string]any{
		"timestamp": at.UTC().Format(time.RFC3339),
		"tools":     map[string]any{"components": []map[string]string{{"type": "application", "name": "smith", "version": smithVersion}}},
		"component": bomComponent{Type: "application", BOMRef: sub.Name, Name: sub.Name, Version: sub.Version, PURL: sub.PURL},
	}
	if len(list) == 0 {
		// An empty list is a statement, not a gap: say so.
		meta["properties"] = []map[string]string{{"name": "smith:components", "value": "none declared: the project lists no third-party dependency in a lockfile or manifest"}}
	}
	doc := map[string]any{
		"bomFormat":   "CycloneDX",
		"specVersion": "1.5",
		"version":     1,
		"metadata":    meta,
		"components":  list,
	}
	return json.MarshalIndent(doc, "", "  ")
}
