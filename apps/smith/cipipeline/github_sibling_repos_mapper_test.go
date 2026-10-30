package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func siblingModel() Model {
	m := model()
	m.SiblingRepos = []workspace.SiblingRepo{{
		Repo: "acme/VideoPlugins", Ref: "main", Env: "VIDEOPLUGINS", Dir: "VideoPlugins",
		PreBuild: "cargo build --release -p videoplugins-music-sync", TokenSecret: "PLUGINS_TOKEN",
	}}
	return m
}

// The sibling repo is checked out into the workspace, its path exported, and
// its pre-build runs after the toolchains and before smith's own setup.
func TestSiblingRepoIsCheckedOutExportedAndPreBuilt(t *testing.T) {
	out := RenderGitHub(siblingModel())
	for _, want := range []string{
		"repository: acme/VideoPlugins\n          path: VideoPlugins\n          ref: 'main'\n          token: ${{ secrets.PLUGINS_TOKEN }}\n",
		`echo "VIDEOPLUGINS=$GITHUB_WORKSPACE/VideoPlugins" >> "$GITHUB_ENV"`,
		"working-directory: VideoPlugins\n        run: cargo build --release -p videoplugins-music-sync\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	checkout, build, setup := strings.Index(out, "Check out acme/VideoPlugins"), strings.Index(out, "Build VideoPlugins"), strings.Index(out, "name: Setup")
	if checkout >= build || build >= setup {
		t.Errorf("order: checkout %d, pre-build %d, setup %d", checkout, build, setup)
	}
}

func TestNoSiblingReposChangeNothing(t *testing.T) {
	if out := RenderGitHub(model()); strings.Contains(out, "VideoPlugins") || strings.Contains(out, "GITHUB_ENV") {
		t.Errorf("unexpected sibling steps:\n%s", out)
	}
}
