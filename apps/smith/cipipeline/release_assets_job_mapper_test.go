package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestGitHubBuildsReleaseFilesOnARunnerPerPlatformAndUploadsThemWithoutArtifacts(t *testing.T) {
	m := NewModel(Contributions{AssetPlatforms: []string{"windows-x64", "macos-arm64", "linux-arm64"}}, "dev", "main")
	out := RenderGitHub(m)
	for _, want := range []string{
		"  assets:\n    needs: ci\n    if: github.event_name == 'push' && github.ref_name == 'main'",
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

func TestLimitsSayWhichProvidersDoNotBuildReleaseFiles(t *testing.T) {
	m := NewModel(Contributions{AssetPlatforms: []string{"linux-x64"}}, "dev", "main")
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github", "gitlab"}}, Release: workspace.ReleaseSection{Branch: "main"}}
	limits := strings.Join(Limits(cfg, m), "\n")
	if !strings.Contains(limits, "gitlab: the release files a [release] build command makes") || strings.Contains(limits, "github:") {
		t.Errorf("limits:\n%s", limits)
	}
}
