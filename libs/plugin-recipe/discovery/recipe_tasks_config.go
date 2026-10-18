package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// Tasks are a recipe project's: test checks the recipe, with its parents, for
// every combination of answers, so a pull request that breaks one branch of a
// recipe fails. It runs in the project's folder.
func Tasks(file string) map[string]protocol.Task {
	return map[string]protocol.Task{
		"test": {
			Command:     []string{"{smith}", "recipe", "check", file, "--all"},
			Inputs:      []string{"**/*"},
			Cache:       true,
			Description: "Check the recipe for every combination of answers (smith recipe check --all)",
		},
	}
}
