package release

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func sampleReport() Report {
	return Report{Results: []ReportResult{smithSample()}}
}

func TestTheSummaryOfTheRealSampleIsPinnedLineByLine(t *testing.T) {
	var got []string
	for _, line := range sampleReport().SummaryLines(false) {
		got = append(got, line.Text)
	}
	want := []string{
		"══ Released ════════════════════════════════════════════════",
		"smith 1.128.0 › 1.129.0 (minor)",
		"  npm        @monosmith/cli + monosmith + 6 platform packages    registry.npmjs.org",
		"               darwin x64 arm64 · linux x64 arm64 · win32 x64 arm64",
		"  PyPI       6 wheels (macOS, Linux, Windows x x64, arm64)    pypi.org",
		"  NuGet      monosmith + 6 platform packages    nuget.org",
		"               darwin x64 arm64 · linux x64 arm64 · windows x64 arm64",
		"  RubyGems   8 gems (macOS, Linux, Windows x x64, arm64)    rubygems.org",
		"  GitHub     release smith@1.129.0 with 17 files: 6 zips, 4 deb/rpm, checksums, SBOM, winget, brew, manifest",
		"  pub.dev    publish queued (pub.yml), not finished yet",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if ascii := sampleReport().SummaryLines(true)[0].Text; ascii != "== Released ==" {
		t.Errorf("ASCII banner %q", ascii)
	}
}

func TestTheSummaryIsOmittedWhenNothingWasReleasedToATarget(t *testing.T) {
	dry := sampleReport()
	dry.DryRun = true
	failed := sampleReport()
	failed.Results[0].Status = "failed"
	for name, r := range map[string]Report{"dry run": dry, "failed": failed, "empty": {}} {
		if len(r.SummaryLines(false)) != 0 || r.SummaryMarkdown() != "" {
			t.Errorf("%s printed a summary", name)
		}
	}
}

func TestTheSummaryIsInTheTextAndTheMarkdownOfTheReport(t *testing.T) {
	r := sampleReport()
	if text := r.Text(); !strings.Contains(text, "══ Released ═") || !strings.Contains(text, "  PyPI       6 wheels") {
		t.Errorf("text:\n%s", text)
	}
	md := r.Markdown()
	for _, want := range []string{
		"#### Released", "**smith** 1.128.0 → **1.129.0** (minor)", "- **npm** @monosmith/cli + monosmith + 6 platform packages (registry.npmjs.org)",
		"  darwin x64 arm64 · linux x64 arm64 · win32 x64 arm64", "- **pub.dev** publish queued (pub.yml), not finished yet",
		"([open](https://github.com/monosmith-app/monosmith/releases/tag/smith%401.129.0))",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown lacks %q:\n%s", want, md)
		}
	}
}

func TestThePubDispatchSaysItIsQueuedInTheTextAndTheMarkdown(t *testing.T) {
	r := Report{Results: []ReportResult{{Item: Item{Project: "app", Next: "1.0.0"}, PubDispatch: "app 1.0.0 from app-v1.0.0"}}}
	want := "Queued the pub.dev publish for app 1.0.0 from app-v1.0.0 (pub.yml). It is not finished yet."
	if !strings.Contains(r.Text(), want) || !strings.Contains(r.Markdown(), want) {
		t.Errorf("text:\n%s\nmarkdown:\n%s", r.Text(), r.Markdown())
	}
}

func TestAProjectWithManyTargetsKeepsItsBlockShort(t *testing.T) {
	res := ReportResult{Item: Item{Project: "web", Next: "1.0.0"}, Status: "released", Mirror: "acme/web@v1.0.0"}
	for i := 0; i < 20; i++ {
		res.Images = append(res.Images, protocol.PublishedImage{Reference: fmt.Sprintf("ghcr.io/acme/web%d", i), Tags: []string{"1.0.0"}})
	}
	lines := Report{Results: []ReportResult{res}}.SummaryLines(false)
	if len(lines) > 2+maxTargetLines+1 || !strings.Contains(lines[len(lines)-1].Text, "more targets") {
		t.Fatalf("%d lines, last %q", len(lines), lines[len(lines)-1].Text)
	}
}
