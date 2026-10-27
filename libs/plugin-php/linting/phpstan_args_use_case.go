package linting

import (
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// configNames are the files PHPStan reads from its working directory.
var configNames = []string{"phpstan.neon", "phpstan.neon.dist", "phpstan.dist.neon"}

// Args returns the arguments `phpstan analyse` takes for its rules. PHPStan
// looks for its configuration in the working directory only, which is the
// package's: when one is there nothing is passed and PHPStan reads it
// (including its level); otherwise smith's default is written to
// .smith/defaults/ and passed with -c. The second result says which of the two
// applies, for the task's description.
func Args(workspaceRoot, projectRoot string) (args []string, own bool, err error) {
	if tooldefaults.HasFile(configNames...)(filepath.Join(workspaceRoot, filepath.FromSlash(projectRoot))) {
		return nil, true, nil
	}
	rel, err := tooldefaults.WriteDefault(workspaceRoot, tooldefaults.PHPStan.File, defaultConfig)
	if err != nil {
		return nil, false, err
	}
	return []string{"-c", "{workspaceRoot}/" + rel}, false, nil
}
