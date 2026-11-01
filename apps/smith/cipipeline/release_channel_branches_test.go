package cipipeline

import (
	"strings"
	"testing"
)

func channelModel() Model {
	m := model()
	m.Pack = true
	m.ChannelBranches = []string{"next"}
	return m
}

func TestEveryProviderReleasesOnAChannelBranchAndPacksOnlyOnTheReleaseBranch(t *testing.T) {
	m := channelModel()
	for name, c := range map[string]struct {
		out  string
		want []string
	}{
		"github": {RenderGitHub(m), []string{
			"branches: [main, next]",
			"(github.ref_name == 'main' || github.ref_name == 'next')",
			"run: smith ci pack\n",
		}},
		"gitlab": {RenderGitLab(m), []string{
			`$CI_COMMIT_BRANCH == "next" && $CI_PIPELINE_SOURCE == "push"`,
			`{ [ "$CI_COMMIT_BRANCH" = "main" ] || [ "$CI_COMMIT_BRANCH" = "next" ]; }`,
		}},
		"azure": {RenderAzure(m), []string{
			"include: [main, next]",
			"or(eq(variables['Build.SourceBranch'], 'refs/heads/main'), eq(variables['Build.SourceBranch'], 'refs/heads/next'))",
		}},
		"bitbucket": {RenderBitbucket(m), []string{"\n    next:\n"}},
		"circleci":  {RenderCircleCI(m), []string{"or: [ { equal: [ main, << pipeline.git.branch >> ] }, { equal: [ next, << pipeline.git.branch >> ] } ]"}},
	} {
		for _, want := range c.want {
			if !strings.Contains(c.out, want) {
				t.Errorf("%s: missing %q in\n%s", name, want, c.out)
			}
		}
	}
	if gh := RenderGitHub(m); strings.Contains(gh, "github.ref_name == 'next'\n        run: smith ci pack") || strings.Count(gh, "'next'") != 1 {
		t.Errorf("only the release step runs on the channel branch:\n%s", gh)
	}
}

func TestWithoutAChannelThePipelinesAreAsBefore(t *testing.T) {
	m := model()
	m.Pack = true
	if gh := RenderGitHub(m); !strings.Contains(gh, "if: github.event_name == 'push' && github.ref_name == 'main'\n        run: smith ci release") || strings.Contains(gh, "||") {
		t.Errorf("a single release branch keeps its plain condition:\n%s", gh)
	}
}
