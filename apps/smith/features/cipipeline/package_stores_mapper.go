package cipipeline

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Where Azure Pipelines and GitLab CI keep the package stores they cache:
// each caches only directories under its own workspace, so the stores sit
// under the pipeline workspace, or under smith's own ignored .smith/ in the
// GitLab project.
const (
	azureStoresDir  = "$(Pipeline.Workspace)/.smith-stores"
	gitlabStoresDir = ".smith/stores"
	// gitlabMaxCaches is how many caches a GitLab CI job may declare.
	gitlabMaxCaches = 4
)

// githubStores renders one actions/cache step per store, keyed on the
// runner's OS and the lockfiles. The cache falls back to the latest one of
// the same store when a lockfile changed, and a pull request from a fork
// restores it without writing the base repository's cache.
func githubStores(stores []protocol.PackageStore) string {
	var b strings.Builder
	for _, s := range stores {
		quoted := make([]string, len(s.Files))
		for i, f := range s.Files {
			quoted[i] = "'" + f + "'"
		}
		fmt.Fprintf(&b, `      - name: Cache the %[1]s download store
        uses: actions/cache@v4
        with:
          path: %[2]s
          key: smith-%[1]s-${{ runner.os }}-${{ hashFiles(%[3]s) }}
          restore-keys: |
            smith-%[1]s-${{ runner.os }}-
`, s.Name, s.Path, strings.Join(quoted, ", "))
	}
	return b.String()
}

// azureStoreVariables renders the variables that move each store under the
// pipeline workspace, or nothing.
func azureStoreVariables(stores []protocol.PackageStore) string {
	if len(stores) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("variables:\n")
	for _, s := range stores {
		fmt.Fprintf(&b, "  %s: %s/%s\n", s.Env, azureStoresDir, s.Name)
	}
	b.WriteString("\n")
	return b.String()
}

// azureStores renders one Cache@2 task per store.
func azureStores(stores []protocol.PackageStore) string {
	var b strings.Builder
	for _, s := range stores {
		fmt.Fprintf(&b, `  - task: Cache@2
    displayName: Cache the %[1]s download store
    inputs:
      key: 'smith-%[1]s | "$(Agent.OS)" | %[2]s'
      restoreKeys: |
        smith-%[1]s | "$(Agent.OS)"
      path: %[3]s/%[1]s
`, s.Name, strings.Join(s.Files, " | "), azureStoresDir)
	}
	return b.String()
}

// gitlabStoreVariables renders the variables that move each store into the
// project directory, one line each.
func gitlabStoreVariables(stores []protocol.PackageStore) string {
	var b strings.Builder
	for _, s := range stores {
		fmt.Fprintf(&b, "    %s: $CI_PROJECT_DIR/%s/%s\n", s.Env, gitlabStoresDir, s.Name)
	}
	return b.String()
}

// gitlabStores renders the job's cache: a list, one entry per store, keyed on
// the first two lockfiles (GitLab's limit) with the store's name as prefix.
// GitLab allows four caches per job; any more are left as a comment.
func gitlabStores(stores []protocol.PackageStore) string {
	if len(stores) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("  cache:\n")
	for i, s := range stores {
		if i == gitlabMaxCaches {
			fmt.Fprintf(&b, "    # GitLab allows %d caches per job; not cached: %s", gitlabMaxCaches, s.Name)
			for _, rest := range stores[i+1:] {
				b.WriteString(", " + rest.Name)
			}
			b.WriteString("\n")
			break
		}
		files := s.Files
		if len(files) > 2 {
			files = files[:2]
		}
		fmt.Fprintf(&b, "    - key:\n        prefix: %s\n        files: [%s]\n      paths:\n        - %s/%s\n", s.Name, strings.Join(files, ", "), gitlabStoresDir, s.Name)
	}
	return b.String()
}
