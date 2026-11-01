package cipipeline

import (
	"strings"
	"testing"
)

func TestTheNativeAndAssetsJobsRunOnAReleaseThatExistsEvenWhenALaterStepFails(t *testing.T) {
	m := model()
	m.Native = true
	m.AssetPlatforms = []string{"linux-x64"}
	m.Attest = true
	out := RenderGitHub(m)
	release := strings.Index(out, "- name: Release\n        id: release\n")
	attest := strings.Index(out, "- name: Attest build provenance")
	if release < 0 || attest < release {
		t.Fatalf("the release step has an id and precedes the attestation (%d, %d):\n%s", release, attest, out)
	}
	// A push that releases nothing must not start the jobs that attach files:
	// the output comes from smith ci release alone, never from a marker step.
	if strings.Contains(out, "Mark the release done") || strings.Contains(out, "steps.released") {
		t.Errorf("no step may set released on a push that released nothing:\n%s", out)
	}
	for _, want := range []string{
		"      released: ${{ steps.release.outputs.released }}\n    steps:",
		"  native:\n    needs: ci\n    if: ${{ always() && (needs.ci.result == 'success' || (github.event_name == 'push' && needs.ci.outputs.released == 'true')) }}",
		"  assets:\n    needs: ci\n    if: ${{ always() && ((needs.ci.outputs.released == 'true'",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestWithNoJobAfterTheCiJobThereIsNoMarkerAndNoOutputs(t *testing.T) {
	out := RenderGitHub(model())
	if strings.Contains(out, "Mark the release done") || strings.Contains(out, "outputs:") {
		t.Errorf("a workspace with no native or asset jobs keeps its pipeline as it was:\n%s", out)
	}
}

func TestTheSignInsRunWhereTheReleaseRunsAChannelBranchIncluded(t *testing.T) {
	m := model()
	m.VSCEEntra, m.RubyGemsTrusted = true, true
	m.ChannelBranches = []string{"next"}
	out := RenderGitHub(m)
	for _, step := range []string{"Sign in to Azure", "Sign in to RubyGems"} {
		want := "- name: " + step + "\n        if: github.event_name == 'push' && (github.ref_name == 'main' || github.ref_name == 'next')\n"
		if !strings.Contains(out, want) {
			t.Errorf("%s must run on the channel branch too:\n%s", step, out)
		}
	}
	plain := model()
	plain.VSCEEntra = true
	if !strings.Contains(RenderGitHub(plain), "- name: Sign in to Azure\n        if: github.event_name == 'push' && github.ref_name == 'main'\n") {
		t.Error("without a channel the condition is as it was")
	}
}
