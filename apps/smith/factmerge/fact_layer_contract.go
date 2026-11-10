package factmerge

// Scope says how narrow a layer of facts is. A narrower scope wins over a
// wider one.
type Scope int

// The scopes, from the widest to the narrowest. ScopeRecipe is reserved: no
// layer uses it until a recipe can carry facts.
const (
	// ScopePluginDefault is a default a plugin publishes through its
	// Contribution.
	ScopePluginDefault Scope = iota
	// ScopeRecipe is a fact a recipe sets.
	ScopeRecipe
	// ScopeWorkspace is smith.toml [facts].
	ScopeWorkspace
	// ScopeProject is smith.toml [projects.<name>.facts].
	ScopeProject
)

// String names the scope the way smith explain prints it.
func (s Scope) String() string {
	switch s {
	case ScopePluginDefault:
		return "plugin default"
	case ScopeRecipe:
		return "recipe"
	case ScopeWorkspace:
		return "workspace"
	case ScopeProject:
		return "project"
	}
	return "unknown"
}

// Layer is the facts one source publishes at one scope.
type Layer struct {
	Scope Scope
	// Source names who published the layer ("plugin typescript (stack
	// node-jest)", "smith.toml [facts]"); it appears in conflicts and in
	// smith explain.
	Source string
	Facts  map[string]any
}

// Fact is one merged fact: its value and where the value came from.
type Fact struct {
	Value any
	// Scope is the scope of the layer that supplied the value.
	Scope Scope
	// Source names the layer that supplied the value; when several layers of
	// the same scope agree, it lists them all, separated by " and ".
	Source string
}
