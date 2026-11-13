package functionhost

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"

// Tools are the programs the app commands run (npm). A test replaces them.
type Tools struct {
	// Run executes argv in dir and streams its output.
	Run func(dir string, argv []string) error
	// Output executes argv in dir and returns what it printed on stdout.
	Output func(dir string, argv []string) (string, error)
}

// DefaultTools runs the real programs.
//
// Returns the tools that call toolrun.
func DefaultTools() Tools {
	return Tools{
		Run:    func(dir string, argv []string) error { return toolrun.Run(dir, argv) },
		Output: func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) },
	}
}
