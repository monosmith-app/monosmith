package protocol

// Stack is a named set of plugins a project is built with: one language
// plugin, and the toolchain, structure, scaffold and deploy plugins on top.
// Plugins are named, never assumed to be built in.
type Stack struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Language is the language plugin that discovers the stack's projects.
	Language string `json:"language"`
	// Toolchain plugins decide how projects build, test and lint, in order:
	// a later one may replace a task an earlier one added.
	Toolchain []string `json:"toolchain,omitempty"`
	Structure []string `json:"structure,omitempty"`
	// Scaffold generates new projects; empty means the language plugin does.
	Scaffold string   `json:"scaffold,omitempty"`
	Deploy   []string `json:"deploy,omitempty"`
	// Choices make the stack a saved selection (see Choice). When they are
	// set, the host derives the language, the scaffold and the toolchain from
	// them, and Toolchain lists only the plugins to add after those. The
	// plugins of the stack get the choices in GenerateParams.Choices.
	Choices []string `json:"choices,omitempty"`
}

// NoStack, recorded as a project's stack, means the language plugin alone,
// even where the workspace sets a default stack.
const NoStack = "none"

// Plugins lists the stack's plugins other than its language and scaffold,
// in the order they apply.
func (s Stack) Plugins() []string {
	out := append([]string(nil), s.Toolchain...)
	out = append(out, s.Structure...)
	return append(out, s.Deploy...)
}
