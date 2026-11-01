package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGitHubBuildsReleaseFilesOnARunnerPerPlatformAndUploadsThemWithoutArtifacts(t *testing.T) {
	m := NewModel(Contributions{AssetPlatforms: []string{"windows-x64", "macos-arm64", "linux-arm64"}}, "dev", "main")
	out := RenderGitHub(m)
	for _, want := range []string{
		"  assets:\n    needs: ci\n    if: ${{ always() && needs.ci.outputs.released == 'true' && github.event_name == 'push' && github.ref_name == 'main' }}",
		"- platform: windows-x64\n            os: windows-latest",
		"- platform: macos-arm64\n            os: macos-latest",
		"- platform: linux-arm64\n            os: ubuntu-24.04-arm",
		"run: smith ci assets\n        env:\n          SMITH_PLATFORM: ${{ matrix.platform }}",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("workflow lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "upload-artifact") {
		t.Errorf("the release files go to the release, not to an artifact:\n%s", out)
	}
	if plain := RenderGitHub(NewModel(Contributions{}, "dev", "main")); strings.Contains(plain, "smith ci assets") {
		t.Errorf("a workspace without a build command gets no assets job")
	}
}

func TestAssetsJobSetsUpWithTheAssetsFlagAndTheCiJobDoesNot(t *testing.T) {
	out := RenderGitHub(NewModel(Contributions{AssetPlatforms: []string{"linux-x64"}}, "dev", "main"))
	if strings.Count(out, "run: smith ci setup --assets\n") != 1 || strings.Count(out, "run: smith ci setup\n") < 1 {
		t.Errorf("only the assets job passes --assets:\n%s", out)
	}
	if i := strings.Index(out, "  assets:"); !strings.Contains(out[i:], "smith ci setup --assets") || strings.Contains(out[:i], "--assets") {
		t.Errorf("--assets belongs to the assets job:\n%s", out)
	}
}

func TestTheAssetsSetupLeavesOutTheProjectsThatOptedOutOfInstall(t *testing.T) {
	no, yes := false, true
	project := func(name string, install *bool, build string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name}, Plugin: "python", Release: &workspace.ProjectRelease{Build: build, Install: install}}
	}
	all := []*projectgraph.Project{project("zip", &no, "make zip"), project("keep", &yes, "make"), project("unset", nil, "make"), project("noBuild", &no, "")}
	names := func(assets bool) string {
		var got []string
		for _, p := range servedProjects(all, "python", assets) {
			got = append(got, p.Name)
		}
		return strings.Join(got, ",")
	}
	if got := names(false); got != "zip,keep,unset,noBuild" {
		t.Errorf("the ci job installs every project: %s", got)
	}
	if got := names(true); got != "keep,unset,noBuild" {
		t.Errorf("the assets job leaves out install = false: %s", got)
	}
}

func TestLimitsSayWhichProvidersDoNotBuildReleaseFiles(t *testing.T) {
	m := NewModel(Contributions{AssetPlatforms: []string{"linux-x64"}}, "dev", "main")
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github", "gitlab"}}, Release: workspace.ReleaseSection{Branch: "main"}}
	limits := strings.Join(Limits(cfg, m), "\n")
	if !strings.Contains(limits, "gitlab: the release files a [release] build command makes") || strings.Contains(limits, "github:") {
		t.Errorf("limits:\n%s", limits)
	}
}

// The assets job checks the sibling repository out and pre-builds it too,
// before it runs smith ci assets.
func TestAssetsJobChecksOutAndPreBuildsSiblingRepos(t *testing.T) {
	m := siblingModel()
	m.AssetPlatforms = []string{"linux-x64"}
	job := githubAssetsJob(m)
	for _, want := range []string{"Check out acme/VideoPlugins", "token: ${{ secrets.PLUGINS_TOKEN }}", "VIDEOPLUGINS=", "Build VideoPlugins"} {
		if !strings.Contains(job, want) {
			t.Errorf("missing %q in:\n%s", want, job)
		}
	}
	if strings.Index(job, "Build VideoPlugins") >= strings.Index(job, "smith ci assets") {
		t.Errorf("pre-build must come before smith ci assets:\n%s", job)
	}
}
