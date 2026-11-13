package linting

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// configNames are the files golangci-lint reads itself.
var configNames = []string{".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json"}

// Lint is the lint task's command and description for one module.
type Lint struct {
	Command []string
	// Fix is Command with golangci-lint's --fix, which rewrites what it can.
	Fix         []string
	Description string
}

// withFix returns the lint with Fix set: the command's own arguments, then
// --fix right after "run".
func withFix(l Lint) Lint {
	l.Fix = append([]string{}, l.Command[:2]...)
	l.Fix = append(l.Fix, "--fix")
	l.Fix = append(l.Fix, l.Command[2:]...)
	return l
}

// Command returns how to lint the module at projectRoot. A golangci-lint
// configuration in the module or above it (up to the workspace root) is the
// repository's override: golangci-lint finds it itself and nothing is passed.
// Without one, smith's default is written to .smith/defaults/ and passed with
// --config.
func Command(workspaceRoot, projectRoot string) (Lint, error) {
	base := []string{"golangci-lint", "run", "--allow-serial-runners"}
	if tooldefaults.FindConfig(workspaceRoot, projectRoot, tooldefaults.HasFile(configNames...)) {
		return withFix(Lint{Command: base, Description: "golangci-lint run (the repository's .golangci config)"}), nil
	}
	rel, err := tooldefaults.WriteDefault(workspaceRoot, tooldefaults.GolangciLint.File, defaultConfig)
	if err != nil {
		return Lint{}, err
	}
	return withFix(Lint{
		Command:     append(base, "--config", "{workspaceRoot}/"+rel),
		Description: "golangci-lint run (smith's default config)",
	}), nil
}
