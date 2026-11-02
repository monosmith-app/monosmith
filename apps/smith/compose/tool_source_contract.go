package compose

// Tool is a program a selection needs on the computer.
type Tool struct {
	// Choice is the pick that needs it.
	Choice string `json:"choice"`
	// Name is what the tool is called in text.
	Name string `json:"name"`
	// Binary is the program on PATH; empty when the project installs it itself.
	Binary string `json:"binary,omitempty"`
	// Install says how to get it, in one sentence.
	Install string `json:"install,omitempty"`
}

// ToolSource says which tools a pick needs. It is the hook where the data of
// the tool installer joins the composer: a source returns nothing for a pick
// that needs no tool. The composer never installs anything itself.
//
// Parameters: choice is the choice name; language is its language.
// Returns: the tools, or nil.
type ToolSource func(choice, language string) []Tool
