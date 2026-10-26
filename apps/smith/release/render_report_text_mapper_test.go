package release

import (
	"strings"
	"testing"
)

func TestTextOfADryRunHasNoTableAndSaysWhatNext(t *testing.T) {
	r := Report{DryRun: true, Unchanged: []string{"a", "b"}, Results: []ReportResult{{Item: Item{Project: "kit", Current: "1.0.0", Next: "1.1.0"}, Status: "planned"}}}
	text := r.Text()
	if strings.Contains(text, "|") || strings.Contains(text, "###") || strings.Contains(text, "kit") {
		t.Fatalf("a dry run prints the plan once, in plain text:\n%s", text)
	}
	if !strings.Contains(text, "→ Push to the release branch") {
		t.Fatalf("%s", text)
	}
	if md := r.Markdown(); !strings.Contains(md, "| Package |") {
		t.Fatalf("the Markdown stays for CI summaries:\n%s", md)
	}
}

func TestTextOfADryRunNamesEachProjectThatWouldFail(t *testing.T) {
	r := Report{DryRun: true, Error: "web: cannot", Results: []ReportResult{
		{Item: Item{Project: "web", Next: "1.0.0"}, Status: "failed", Error: "cannot build linux/arm64"},
		{Item: Item{Project: "tool", Next: "1.0.0"}, Status: "planned"},
	}}
	for name, text := range map[string]string{"text": r.Text(), "markdown": r.Markdown()} {
		if !strings.Contains(text, "would fail") || !strings.Contains(text, "web: cannot build linux/arm64") || strings.Contains(text, "Stopped") || strings.Contains(text, "Re-run") {
			t.Fatalf("%s:\n%s", name, text)
		}
	}
}

func TestTextOfARealReleaseListsEachPackage(t *testing.T) {
	r := Report{Results: []ReportResult{{Item: Item{Project: "kit", Next: "1.1.0"}, Status: "released"}}, Error: "boom"}
	text := r.Text()
	if !strings.Contains(text, "kit") || !strings.Contains(text, "new → 1.1.0  released") || !strings.Contains(text, "Stopped: boom") {
		t.Fatalf("%s", text)
	}
}
