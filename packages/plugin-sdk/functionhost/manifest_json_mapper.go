package functionhost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// readManifest reads a package.json into a map.
func readManifest(file string) (map[string]any, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var manifest map[string]any
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	return manifest, nil
}

// marshalManifest writes a manifest with two spaces of indent and a final
// newline, and leaves < and & as they are.
func marshalManifest(manifest map[string]any) ([]byte, error) {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(manifest); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// writeManifest writes a manifest to a file.
func writeManifest(file string, manifest map[string]any) error {
	text, err := marshalManifest(manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(file, text, 0o644)
}

// stringMap reads a JSON object of strings (the dependencies of a manifest);
// a value of another kind gives an empty map.
func stringMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}
