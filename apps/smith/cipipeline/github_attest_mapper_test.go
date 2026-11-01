package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func attestModel() Model {
	m := model()
	m.Attest = true
	return m
}

func TestTheAttestationStepFailsTheJobByDefaultAndContinuesWhenOptional(t *testing.T) {
	plain := RenderGitHub(attestModel())
	if !strings.Contains(plain, "- name: Attest build provenance") || strings.Contains(plain, "continue-on-error") {
		t.Fatalf("the default step fails the job when GitHub refuses:\n%s", plain)
	}
	m := attestModel()
	m.AttestOptional = true
	optional := RenderGitHub(m)
	if !strings.Contains(optional, "hashFiles('dist/release/**') != ''\n        continue-on-error: true\n        uses: actions/attest-build-provenance@v3") {
		t.Fatalf("an optional step continues on error:\n%s", optional)
	}
}

func TestAttestOffLeavesTheStepAndItsPermissionsOut(t *testing.T) {
	for attest, wantStep := range map[string]bool{"": true, "on": true, "optional": true, "off": false} {
		cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}, Attest: attest}}
		out := Rendered(cfg, attestModel())[GitHubPath]
		if got := strings.Contains(out, "Attest build provenance"); got != wantStep {
			t.Errorf("attest %q: step present %v, want %v", attest, got, wantStep)
		}
		if got := strings.Contains(out, "attestations: write"); got != wantStep {
			t.Errorf("attest %q: permission present %v, want %v", attest, got, wantStep)
		}
		if got := strings.Contains(out, "continue-on-error: true"); got != (attest == "optional") {
			t.Errorf("attest %q: continue-on-error %v", attest, got)
		}
	}
}
