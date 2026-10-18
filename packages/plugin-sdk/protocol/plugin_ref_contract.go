package protocol

import "strings"

// PluginRef names the plugins a requirement or a conflict is about. Every
// field set must match; at least one must be set.
type PluginRef struct {
	// Level is the role the plugin is chosen for in the project.
	Level Role `json:"level,omitempty"`
	// Name is the plugin's name, or for a language plugin one of its
	// languages ("node", "javascript").
	Name string `json:"name,omitempty"`
	// Slot is a task the plugin provides ("test").
	Slot string `json:"slot,omitempty"`
}

// RequiresLanguage is a requirement on the project's language plugin.
func RequiresLanguage(name string) PluginRef {
	return PluginRef{Level: RoleLanguage, Name: name}
}

// ConflictsWith names a plugin that must not be in the same project.
func ConflictsWith(name string) PluginRef {
	return PluginRef{Name: name}
}

// IsZero reports whether the reference names nothing.
func (r PluginRef) IsZero() bool { return r == PluginRef{} }

// String is how errors name it: "language plugin node", "the test slot".
func (r PluginRef) String() string {
	var parts []string
	if r.Level != "" {
		parts = append(parts, string(r.Level)+" plugin")
	} else {
		parts = append(parts, "plugin")
	}
	if r.Name != "" {
		parts = append(parts, r.Name)
	}
	if r.Slot != "" {
		parts = append(parts, "providing "+r.Slot)
	}
	return strings.Join(parts, " ")
}
