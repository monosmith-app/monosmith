package scaffold

import (
	"cmp"
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/architecture"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectnaming"
)

// renderWorkspaceConfig is the smith.toml smith new writes: the template with
// the answers in place of the commented defaults. A setting left at its
// default stays the commented line, so the file shows it without repeating it.
func renderWorkspaceConfig(opts NewOptions, quotedProviders string) (string, error) {
	text := fmt.Sprintf(configTemplate, opts.Name, namespaceLines(opts), opts.ModulePrefix, quotedProviders)

	if len(opts.DefaultStacks) > 0 {
		var lines strings.Builder
		for _, lang := range sortedKeys(opts.DefaultStacks) {
			stack := opts.DefaultStacks[lang]
			if lang != "node" || !hasOption(NodeStacks(), stack) {
				return "", fmt.Errorf("--default-stack %s=%s: smith new offers %s for node; set other stacks later with smith config set default_stacks.<language> <stack>", lang, stack, optionValues(NodeStacks()))
			}
			fmt.Fprintf(&lines, "%s = %q\n", lang, stack)
		}
		text = strings.Replace(text, defaultStacksText, lines.String(), 1)
	}

	if opts.Architecture != "" && opts.Architecture != DefaultArchitecture {
		chosen, err := architecture.Find(opts.Architecture)
		if err != nil {
			return "", fmt.Errorf("--architecture: %w", err)
		}
		replacement := ""
		if chosen.Table != "" {
			replacement = "[" + chosen.Table + "]\n"
		}
		text = strings.Replace(text, conventionText, replacement, 1)
	}

	var release []string
	if opts.ReleaseBranch != "" && opts.ReleaseBranch != DefaultReleaseBranch {
		release = append(release, fmt.Sprintf("branch = %q", opts.ReleaseBranch))
	}
	if opts.ReleaseRemote != "" && opts.ReleaseRemote != DefaultReleaseRemote {
		release = append(release, fmt.Sprintf("remote = %q", opts.ReleaseRemote))
	}
	if opts.GitHubReleases {
		release = append(release, "github_releases = true")
	}
	if len(release) > 0 {
		text = strings.Replace(text, releaseText, "[release]\n"+strings.Join(release, "\n")+"\n", 1)
	}

	var cache []string
	if opts.CacheDir != "" && opts.CacheDir != DefaultCacheDir {
		cache = append(cache, fmt.Sprintf("dir = %q", opts.CacheDir))
	}
	switch {
	case opts.CacheMaxMB < 0:
		cache = append(cache, "max_size_mb = 0")
	case opts.CacheMaxMB > 0 && opts.CacheMaxMB != DefaultCacheMaxMB:
		cache = append(cache, fmt.Sprintf("max_size_mb = %d", opts.CacheMaxMB))
	}
	if len(cache) > 0 {
		text = strings.Replace(text, cacheText, "[cache]\n"+strings.Join(cache, "\n")+"\n", 1)
	}
	if opts.NoDevcontainer {
		text = strings.Replace(text, "devcontainer = true", "devcontainer = false", 1)
	}
	if opts.NoAgent {
		text += "\n# No MonoSmith agent files in this workspace; smith agent install adds them.\n[agent]\nenabled = false\n"
	}
	return text, nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// namespaceLines are the namespace and scope lines of the new smith.toml. A
// setting the person gave is written; one left at its default stays a
// commented line, so the file shows what names new projects without
// repeating it.
func namespaceLines(opts NewOptions) string {
	namespace := fmt.Sprintf("# namespace = %q", opts.Name)
	if opts.Namespace != "" {
		namespace = fmt.Sprintf("namespace = %q", opts.Namespace)
	}
	scope := fmt.Sprintf("# scope = %q", "@"+projectnaming.Kebab(cmp.Or(opts.Namespace, opts.Name)))
	if opts.Scope != "" {
		scope = fmt.Sprintf("scope = %q", opts.Scope)
	}
	return "# The word every language puts before a new project's name (acme-core, @acme/core,\n# Acme.Core). Default: the workspace name. The scope and module_prefix derive from it.\n" +
		namespace + "\n# npm scope for Node packages. Default: @ and the namespace.\n" + scope + "\n"
}
