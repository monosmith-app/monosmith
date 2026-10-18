package discovery

import (
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Pubspec is what smith reads from a pubspec.yaml.
type Pubspec struct {
	Name, Version string
	// Description is the pubspec's description.
	Description string
	// Flutter is true for a package that depends on the Flutter SDK.
	Flutter bool
	// Kind and Stack are the smith section's.
	Kind, Stack string
	// Dependencies are the names of every dependency, dev ones too.
	Dependencies []string
	// Workspace lists a workspace root's member directories.
	Workspace []string
	// PublishTo is where the package publishes: none, a repository URL, or
	// "" for pub.dev.
	PublishTo string
}

type pubspecYAML struct {
	Name            string         `yaml:"name"`
	Version         string         `yaml:"version"`
	Description     string         `yaml:"description"`
	PublishTo       string         `yaml:"publish_to"`
	Dependencies    map[string]any `yaml:"dependencies"`
	DevDependencies map[string]any `yaml:"dev_dependencies"`
	Workspace       []string       `yaml:"workspace"`
	Smith           struct {
		Kind  string `yaml:"kind"`
		Stack string `yaml:"stack"`
	} `yaml:"smith"`
}

// ParsePubspec reads a pubspec.yaml.
func ParsePubspec(src []byte) (Pubspec, error) {
	var y pubspecYAML
	if err := yaml.Unmarshal(src, &y); err != nil {
		return Pubspec{}, err
	}
	p := Pubspec{Name: y.Name, Version: y.Version, Description: strings.TrimSpace(y.Description), PublishTo: y.PublishTo, Kind: y.Smith.Kind, Stack: y.Smith.Stack, Workspace: y.Workspace}
	for _, deps := range []map[string]any{y.Dependencies, y.DevDependencies} {
		for name, spec := range deps {
			if sdk, ok := spec.(map[string]any); ok && sdk["sdk"] == "flutter" && name == "flutter" {
				p.Flutter = true
				continue
			}
			p.Dependencies = append(p.Dependencies, name)
		}
	}
	sort.Strings(p.Dependencies)
	return p, nil
}
