package discovery

import (
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// inputs are what a package's results depend on: its sources and the
// workspace's resolution.
var inputs = []string{"**/*", "!vendor/**", "{workspaceRoot}/composer.json", "{workspaceRoot}/composer.lock"}

// Shape is what a package's tasks depend on.
type Shape struct {
	Kind protocol.ProjectKind
	// Tests is true when the package has a tests/ directory.
	Tests bool
	// Bin is the app's executable, project-relative.
	Bin string
	// PHPStan are the arguments PHPStan takes for its rules: -c with smith's default, or none when the package has a phpstan.neon.
	PHPStan []string
}

// Tasks are PHPUnit's and PHPStan's for one package, run in its directory
// with the workspace's vendor/: one resolution, one autoloader, for every
// package. PHPStan runs over the sources and the tests, at level 8 with smith's
// default configuration unless the package has a phpstan.neon (see the linting
// slice).
func Tasks(s Shape) map[string]protocol.Task {
	vendor := "{workspaceRoot}/vendor/"
	paths := []string{"src"}
	if s.Tests {
		paths = append(paths, "tests")
	}
	lint := slices.Concat([]string{"php", vendor + "bin/phpstan", "analyse", "--no-progress", "--memory-limit=1G"}, s.PHPStan, []string{"--autoload-file", vendor + "autoload.php"}, paths)
	lintDescription := "PHPStan (the package's phpstan.neon)"
	if len(s.PHPStan) > 0 {
		lintDescription = "PHPStan, level 8 (smith's default config)"
	}
	tasks := map[string]protocol.Task{
		"lint": {Command: lint, Inputs: tooldefaults.WithDefaultInputs(inputs, s.PHPStan), Cache: true, Description: lintDescription},
	}
	if s.Tests {
		tasks["test"] = protocol.Task{Command: []string{"php", vendor + "bin/phpunit", "--bootstrap", vendor + "autoload.php", "tests"}, Inputs: inputs, Cache: true, Description: "PHPUnit"}
	}
	if s.Kind == protocol.ProjectApp && s.Bin != "" {
		tasks["start"] = protocol.Task{Command: []string{"php", s.Bin}, Description: "Run the app"}
	}
	return tasks
}
