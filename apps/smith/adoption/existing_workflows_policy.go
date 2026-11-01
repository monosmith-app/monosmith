package adoption

import (
	"path"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
)

// ExistingWorkflows warns about each workflow of the user's under
// .github/workflows that smith does not own: --init adds its own ci.yml beside
// them, and both run on every push, so CI runs twice.
//
// Parameters: files are the repository's tracked files (slash paths); read
// returns a file's text, "" when it is missing; providers are the CI providers
// the workspace will use.
// Returns: one warning per workflow, none when the workspace does not use GitHub.
func ExistingWorkflows(files []string, read func(string) string, providers []string) []Item {
	github := false
	for _, p := range providers {
		github = github || p == "github"
	}
	if !github {
		return nil
	}
	var items []Item
	for _, f := range files {
		if path.Dir(f) != ".github/workflows" || (!strings.HasSuffix(f, ".yml") && !strings.HasSuffix(f, ".yaml")) {
			continue
		}
		// The file smith owns has its own warning (it is kept as .pre-smith).
		if slices.Contains(cipipeline.OwnedPaths, f) || strings.Contains(read(f), GeneratedMarker) {
			continue
		}
		items = append(items, Item{
			Message: f + " is a workflow of yours: --init adds " + cipipeline.GitHubPath + " beside it, and both run on a push, so CI would run twice",
			Steps: []string{
				"keep it if it does something smith's pipeline does not (a deploy, a release)",
				"or delete it, or limit its \"on:\" to workflow_dispatch, once smith's pipeline covers it",
			},
			Step: "keep it if it does something smith's pipeline does not (a deploy, a release)",
		})
	}
	return items
}
