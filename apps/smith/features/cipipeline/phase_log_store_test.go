package cipipeline

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestAPhaseFoldsItsToolOutputBetweenABannerAndItsResult(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("SMITH_ASCII", "")
	t.Setenv("NO_COLOR", "1")
	var out bytes.Buffer
	log := newPhaseLog(&out, "setup")
	log.banner("Preparing 2 projects.")
	fmt.Fprintln(&out, "npm output")
	log.say("%s", "Step  Result")
	log.conclusion = "1 step ran"
	log.finish(nil, 3200*time.Millisecond)

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if !strings.HasPrefix(lines[0], "══ Setup ═") || len(lines[0]) > 3*bannerWidth {
		t.Errorf("banner: %q", lines[0])
	}
	want := []string{"Preparing 2 projects.", "::group::smith ci setup: output", "npm output", "::endgroup::", "Step  Result", "✓ Setup passed in 3.2s: 1 step ran"}
	if got := lines[1:]; strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("lines\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if strings.Contains(out.String(), "\x1b") {
		t.Error("a CI log must stay plain text")
	}
}

func TestAFailedPhaseNamesTheWayForwardAndLeavesNoGroupOpen(t *testing.T) {
	t.Setenv("TF_BUILD", "True")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("SMITH_ASCII", "1")
	var out bytes.Buffer
	log := newPhaseLog(&out, "audit")
	log.banner("Auditing 3 projects.")
	log.finish(errors.New("boom"), 75*time.Second)
	got := out.String()
	for _, want := range []string{"== Audit ==", "##[group]smith ci audit: output", "##[endgroup]", "[x] Audit failed after 1m 15s", "→ the output above shows the cause; repeat the phase with `smith ci audit`"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Count(got, "##[group]") != strings.Count(got, "##[endgroup]") {
		t.Errorf("unbalanced groups:\n%s", got)
	}
}

func TestDurationsReadLikeSpeech(t *testing.T) {
	for d, want := range map[time.Duration]string{400 * time.Millisecond: "0.4s", 3200 * time.Millisecond: "3.2s", 12 * time.Second: "12s", 192 * time.Second: "3m 12s"} {
		if got := formatDuration(d); got != want {
			t.Errorf("%v: %q, want %q", d, got, want)
		}
	}
}
