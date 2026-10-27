package discovery

import (
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Crate is what smith reads from a Cargo.toml.
type Crate struct {
	// Name is the package name; empty for a virtual (workspace-only)
	// manifest.
	Name    string
	Version string
	// Description is the package's description, when it is a string.
	Description string
	// Kind and Stack are [package.metadata.smith] kind and stack.
	Kind, Stack string
	// Native is [package.metadata.smith] native = true: the crate is built
	// on each operating system, so its release archives come from a native
	// job per runner instead of cross-compiling on one.
	Native bool
	// PathDependencies are the crates it depends on by path (the workspace's
	// own), by package name.
	PathDependencies []string
	// Members are [workspace] members, for a workspace root.
	Members []string
}

type manifest struct {
	Package *struct {
		Name    string `toml:"name"`
		Version any    `toml:"version"`
		// Description is a string, or a table when inherited from the workspace.
		Description any `toml:"description"`
		Metadata    struct {
			Smith struct {
				Kind   string `toml:"kind"`
				Stack  string `toml:"stack"`
				Native bool   `toml:"native"`
			} `toml:"smith"`
		} `toml:"metadata"`
	} `toml:"package"`
	Workspace *struct {
		Members []string `toml:"members"`
	} `toml:"workspace"`
	Dependencies      map[string]any `toml:"dependencies"`
	DevDependencies   map[string]any `toml:"dev-dependencies"`
	BuildDependencies map[string]any `toml:"build-dependencies"`
}

// ParseCargo reads a Cargo.toml.
func ParseCargo(src []byte) (Crate, error) {
	var m manifest
	if err := toml.Unmarshal(src, &m); err != nil {
		return Crate{}, err
	}
	var c Crate
	if m.Package != nil {
		c.Name = m.Package.Name
		if v, ok := m.Package.Version.(string); ok {
			c.Version = v
		}
		if d, ok := m.Package.Description.(string); ok {
			c.Description = strings.TrimSpace(d)
		}
		c.Kind, c.Stack = m.Package.Metadata.Smith.Kind, m.Package.Metadata.Smith.Stack
		c.Native = m.Package.Metadata.Smith.Native
	}
	if m.Workspace != nil {
		c.Members = m.Workspace.Members
	}
	seen := map[string]bool{}
	for _, deps := range []map[string]any{m.Dependencies, m.DevDependencies, m.BuildDependencies} {
		for key, raw := range deps {
			table, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if _, local := table["path"]; !local {
				continue
			}
			name := key
			if pkg, ok := table["package"].(string); ok && pkg != "" {
				name = pkg
			}
			if !seen[name] {
				seen[name] = true
				c.PathDependencies = append(c.PathDependencies, name)
			}
		}
	}
	sort.Strings(c.PathDependencies)
	return c, nil
}
