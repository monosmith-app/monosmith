package pluginhost

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CatalogStack is one stack the workspace can use, and where it was declared.
type CatalogStack struct {
	protocol.Stack
	// Source is "smith.toml" or "plugin <name>".
	Source string `json:"source"`
	// Problem says why the stack cannot be used with the loaded plugins;
	// empty when it can.
	Problem string `json:"problem,omitempty"`
}

// Catalog is every stack the workspace can use: those its plugins publish
// and those smith.toml declares, which win over a published one of the same
// name.
type Catalog struct {
	stacks map[string]CatalogStack
}

// NewCatalog collects and checks the stacks. A stack smith.toml declares that
// cannot be used is an error; a published one is kept with its problem, which
// is reported when a project uses it.
func NewCatalog(set *PluginSet, cfg workspace.Config) (*Catalog, error) {
	c := &Catalog{stacks: map[string]CatalogStack{}}
	for _, client := range set.Clients {
		for _, st := range client.Info.Stacks {
			entry := CatalogStack{Stack: st, Source: "plugin " + client.Info.Plugin.Name}
			if expanded, err := choiceselect.ExpandStack(st, set.Offers()); err != nil {
				entry.Problem = err.Error()
			} else if err := ValidateStack(expanded, set); err != nil {
				entry.Problem = err.Error()
			} else {
				entry.Stack = expanded
			}
			c.stacks[st.Name] = entry
		}
	}
	for _, name := range sortedNames(cfg.Stacks) {
		sc := cfg.Stacks[name]
		st := protocol.Stack{
			Name: name, Description: sc.Description, Language: sc.Language,
			Toolchain: sc.Toolchain, Structure: sc.Structure, Scaffold: sc.Scaffold, Deploy: sc.Deploy,
			Choices: sc.Choices,
		}
		st, err := choiceselect.ExpandStack(st, set.Offers())
		if err != nil {
			return nil, fmt.Errorf("smith.toml [stacks.%s]: %w", name, err)
		}
		if err := ValidateStack(st, set); err != nil {
			return nil, fmt.Errorf("smith.toml [stacks.%s]: %w", name, err)
		}
		c.stacks[name] = CatalogStack{Stack: st, Source: workspace.FileName}
	}
	for lang, name := range cfg.DefaultStacks {
		if _, ok := c.stacks[name]; !ok {
			return nil, fmt.Errorf("smith.toml [default_stacks] %s = %q: no such stack; known stacks: %s", lang, name, strings.Join(c.Names(), ", "))
		}
	}
	return c, nil
}

// Get returns a usable stack by name.
func (c *Catalog) Get(name string) (protocol.Stack, error) {
	st, ok := c.stacks[name]
	if !ok {
		return protocol.Stack{}, fmt.Errorf("no stack named %q; known stacks: %s (smith stacks lists them)", name, strings.Join(c.Names(), ", "))
	}
	if st.Problem != "" {
		return protocol.Stack{}, fmt.Errorf("stack %s (%s) cannot be used: %s", name, st.Source, st.Problem)
	}
	return st.Stack, nil
}

// Names lists the stacks, sorted.
func (c *Catalog) Names() []string { return sortedNames(c.stacks) }

// All lists the stacks, sorted by name.
func (c *Catalog) All() []CatalogStack {
	out := make([]CatalogStack, 0, len(c.stacks))
	for _, name := range c.Names() {
		out = append(out, c.stacks[name])
	}
	return out
}
