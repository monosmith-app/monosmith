package tooldefaults

// Tool is one configuration smith owns for a tool: the file a plugin writes
// under Dir, and the file in the repository that replaces it. `smith eject`
// copies the first to the second when the repository wants to own it.
type Tool struct {
	// Name is what a person types: `smith eject ruff`.
	Name string
	// Language is the plugin that runs the tool, for listings.
	Language string
	// File is the default's name under Dir.
	File string
	// Target is the workspace-relative file that overrides the default.
	Target string
}

// The tools smith ships a default for. A plugin names its default file from
// here, so the plugin and `smith eject` cannot drift apart.
var (
	GolangciLint = Tool{Name: "golangci-lint", Language: "go", File: "golangci.yml", Target: ".golangci.yml"}
	Ruff         = Tool{Name: "ruff", Language: "python", File: "ruff.toml", Target: "ruff.toml"}
	Mypy         = Tool{Name: "mypy", Language: "python", File: "mypy.ini", Target: "mypy.ini"}
	Pytest       = Tool{Name: "pytest", Language: "python", File: "pytest.ini", Target: "pytest.ini"}
)

// Tools lists every Tool.
var Tools = []Tool{GolangciLint, Ruff, Mypy, Pytest}

// Lookup finds a tool by name.
func Lookup(name string) (Tool, bool) {
	for _, t := range Tools {
		if t.Name == name {
			return t, true
		}
	}
	return Tool{}, false
}
