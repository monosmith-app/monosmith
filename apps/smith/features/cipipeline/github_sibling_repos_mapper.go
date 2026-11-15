package cipipeline

import (
	"fmt"
	"strings"
)

// githubSiblingCheckouts are the steps, right after the workspace's own
// checkout, that check each [[ci.sibling_repos]] repository out into the
// workspace and export its path: $GITHUB_WORKSPACE/<dir> in the variable the
// entry names, for every later step.
func githubSiblingCheckouts(m Model) string {
	var b strings.Builder
	for _, r := range m.SiblingRepos {
		fmt.Fprintf(&b, "      - name: Check out %s\n        uses: actions/checkout@v7\n        with:\n          repository: %s\n          path: %s\n", r.Repo, r.Repo, r.Dir)
		if r.Ref != "" {
			fmt.Fprintf(&b, "          ref: %s\n", yamlValue(r.Ref))
		}
		if r.TokenSecret != "" {
			fmt.Fprintf(&b, "          token: ${{ secrets.%s }}\n", r.TokenSecret)
		}
		fmt.Fprintf(&b, "      - name: Export %s\n        shell: bash\n        run: echo \"%s=$GITHUB_WORKSPACE/%s\" >> \"$GITHUB_ENV\"\n", r.Env, r.Env, r.Dir)
	}
	return b.String()
}

// githubSiblingPreBuilds run each entry's pre-build command in its checkout.
// They come after the toolchains are installed (the command may need them)
// and before smith sets the workspace up.
func githubSiblingPreBuilds(m Model) string {
	var b strings.Builder
	for _, r := range m.SiblingRepos {
		if r.PreBuild == "" {
			continue
		}
		fmt.Fprintf(&b, "      - name: Build %s\n        shell: bash\n        working-directory: %s\n        run: %s", r.Dir, r.Dir, runValue("        ", r.PreBuild))
	}
	return b.String()
}
