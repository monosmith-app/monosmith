package clicontract

// CommandDescription is one command in `smith commands --json`.
type CommandDescription struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	// Group is workspace, projects, dependencies, pipeline or inspect; an
	// editor offers every command outside inspect.
	Group       string                `json:"group"`
	Description string                `json:"description"`
	Arguments   []ArgumentDescription `json:"arguments"`
	Options     []OptionDescription   `json:"options"`
	// Subcommands are the command's own subcommands (`deps up`), described the
	// same way and in the same group; absent when it has none.
	Subcommands []CommandDescription `json:"subcommands,omitempty"`
}

// ArgumentDescription is one positional argument.
type ArgumentDescription struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Variadic    bool     `json:"variadic"`
	Choices     []string `json:"choices,omitempty"`
}

// OptionDescription is one flag.
type OptionDescription struct {
	// Name is the long name without dashes.
	Name        string `json:"name"`
	Flags       string `json:"flags"`
	Short       string `json:"short,omitempty"`
	Description string `json:"description"`
	// TakesValue is none (a switch), required or optional.
	TakesValue   string   `json:"takesValue"`
	Variadic     bool     `json:"variadic"`
	Repeatable   bool     `json:"repeatable"`
	Negated      bool     `json:"negated"`
	Choices      []string `json:"choices,omitempty"`
	DefaultValue any      `json:"defaultValue,omitempty"`
}

// KindDescription is one entry of `smith kinds --json`: what `smith add <kind>`
// accepts.
type KindDescription struct {
	Kind          string   `json:"kind"`
	Language      string   `json:"language"`
	Label         string   `json:"label"`
	Description   string   `json:"description"`
	Flags         []string `json:"flags"`
	RequiredFlags []string `json:"requiredFlags,omitempty"`
}

// Info is `smith info --json`.
type Info struct {
	CLI       CLIInfo        `json:"cli"`
	Workspace *WorkspaceInfo `json:"workspace"`
}

// CLIInfo is the installed and newest smith.
type CLIInfo struct {
	Version string `json:"version"`
	// Latest is null when it could not be found out.
	Latest          *string `json:"latest"`
	UpdateAvailable bool    `json:"updateAvailable"`
}

// WorkspaceInfo is the workspace smith runs in, and its smith.toml as JSON.
type WorkspaceInfo struct {
	Root   string         `json:"root"`
	Config map[string]any `json:"config"`
}
