package discovery

import (
	"encoding/json"
	"strings"
)

// AppManifest is what discovery reads from app.json.
type AppManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	// Stack is the stack the bundle is built with, if any.
	Stack string `json:"stack"`
	// DependsOn names what the bundle is built from: identifiers other projects
	// export (a crate or package name), which become edges of the project graph.
	DependsOn []string `json:"dependsOn"`
}

// ParseAppManifest reads app.json. It returns an error for a file that is not
// a JSON object, so a broken app.json is reported instead of silently skipped.
func ParseAppManifest(src []byte) (AppManifest, error) {
	var m AppManifest
	if err := json.Unmarshal(src, &m); err != nil {
		return AppManifest{}, err
	}
	m.Name, m.Version, m.Stack = strings.TrimSpace(m.Name), strings.TrimSpace(m.Version), strings.TrimSpace(m.Stack)
	return m, nil
}
