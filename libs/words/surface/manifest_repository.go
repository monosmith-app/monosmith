package surface

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed *.json
var manifests embed.FS

// Names are the surfaces smith has an interface for.
var Names = []string{"tui", "vscode", "jetbrains"}

// Parse reads a manifest from its JSON text.
//
// Errors: JSON that is not a manifest, or has a field the manifest lacks.
func Parse(data []byte) (Manifest, error) {
	var m Manifest
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return Manifest{}, fmt.Errorf("surface manifest: %w", err)
	}
	return m, nil
}

// Load reads the embedded manifest of the surface called name.
//
// Errors: a name with no manifest, or a manifest that does not parse.
func Load(name string) (Manifest, error) {
	data, err := manifests.ReadFile(name + ".json")
	if err != nil {
		return Manifest{}, fmt.Errorf("surface %q: %w", name, err)
	}
	return Parse(data)
}
