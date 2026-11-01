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
	release := strings.Index(out, "- name: Release\n")
	marker := strings.Index(out, "- name: Mark the release done\n        id: released\n        if: github.event_name == 'push' && github.ref_name == 'main'\n        run: echo \"done=true\" >> \"$GITHUB_OUTPUT\"")
	attest := strings.Index(out, "- name: Attest build provenance")
	if release < 0 || marker < release || attest < marker {
		t.Fatalf("the marker follows the release step and precedes the attestation (%d, %d, %d):\n%s", release, marker, attest, out)
	}
	for _, want := range []string{
		"  ci:\n    runs-on: ubuntu-latest\n    outputs:\n      released: ${{ steps.released.outputs.done }}\n    steps:",
		"  native:\n    needs: ci\n    if: ${{ always() && (needs.ci.result == 'success' || (github.event_name == 'push' && needs.ci.outputs.released == 'true')) }}",
		"  assets:\n    needs: ci\n    if: ${{ always() && needs.ci.outputs.released == 'true'",
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
