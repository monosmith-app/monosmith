// A minimal smith language plugin in Go, built on the plugin SDK.
//
// It teaches smith one made-up project type: every directory that holds a
// widget.json is a project, and each gets a "hello" task.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// widgetManifest is the content of a widget.json.
type widgetManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Kind    string `json:"kind"` // app, lib or package
}

type widget struct{ root string }

// Describe answers "initialize". serve fills in the protocol version and the
// capabilities from the interfaces this type implements.
func (w *widget) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	w.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "widget", Version: "0.1.0"},
		Kind:      protocol.KindLanguage,
		Roles:     []protocol.Role{protocol.RoleLanguage}, // not also scaffold: this plugin does not generate
		Languages: []string{"widget"},
		Manifests: []string{"widget.json"},
	}, nil
}

// Discover answers "discover". Having this method is what advertises the
// discover capability.
func (w *widget) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range p.Manifests { // workspace-relative, forward slashes
		raw, err := os.ReadFile(filepath.Join(w.root, filepath.FromSlash(manifest)))
		if err != nil {
			return result, err
		}
		var m widgetManifest
		if err := json.Unmarshal(raw, &m); err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		if m.Name == "" {
			continue // not ours
		}
		kind := protocol.ProjectKind(m.Kind)
		if kind == "" {
			kind = protocol.ProjectLib
		}
		result.Projects = append(result.Projects, protocol.Project{
			Name:     m.Name,
			Root:     path.Dir(manifest),
			Language: "widget",
			Kind:     kind,
			Version:  m.Version,
			Exports:  []string{m.Name},
			Tasks: map[string]protocol.Task{
				"hello": {
					Command:     []string{"git", "--version"},
					Inputs:      []string{"widget.json"},
					Cache:       true,
					Description: "say hello, as a widget does",
				},
			},
		})
	}
	return result, nil
}

func main() { serve.Main(&widget{}) }
