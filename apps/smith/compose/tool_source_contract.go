package compose

// Tool is a program a selection needs on the computer, with whether it is
// there. The composer never installs anything itself: the line in Command is
// what the person runs (ADR 0077).
type Tool struct {
	// Choice is the pick that needs it: a choice, or the language.
	Choice string `json:"choice"`
	// Name is the tool's name in `smith tools install <name>`.
	Name string `json:"name"`
	// Title is what the tool is called in text.
	Title string `json:"title"`
	// Binary is the program on PATH.
	Binary string `json:"binary,omitempty"`
	// State is "installed", "outdated" or "missing".
	State string `json:"state"`
	// Command is the line that offers the install: `smith tools install <name>`.
	Command string `json:"command"`
}
