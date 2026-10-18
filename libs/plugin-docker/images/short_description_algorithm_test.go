package images

import (
	"strings"
	"testing"
)

func TestAShortDescriptionWithinTheLimitIsKept(t *testing.T) {
	got, cut := FitShortDescription("  Crawl the web\n with recipes ")
	if got != "Crawl the web with recipes" || cut {
		t.Errorf("%q %v", got, cut)
	}
}

func TestALongShortDescriptionIsCutAtAWordWithAnEllipsis(t *testing.T) {
	// OpenCraw's azure-host description is 127 characters.
	text := "Crawl the web with recipes: the OpenCraw engine over HTTP, with its MCP endpoint, packaged as an Azure Functions host app."
	got, cut := FitShortDescription(text)
	if !cut || got != "Crawl the web with recipes: the OpenCraw engine over HTTP, with its MCP endpoint, packaged as an…" {
		t.Errorf("%q %v", got, cut)
	}
	if n := len([]rune(got)); n > ShortDescriptionLimit {
		t.Errorf("%d characters", n)
	}
}

func TestAWordLongerThanTheLimitIsCutWhereItStands(t *testing.T) {
	got, cut := FitShortDescription(strings.Repeat("é", 150))
	if !cut || len([]rune(got)) != ShortDescriptionLimit || !strings.HasSuffix(got, "…") {
		t.Errorf("%q %v", got, cut)
	}
}

func TestADescriptionOfExactlyTheLimitIsNotCut(t *testing.T) {
	text := strings.Repeat("a", ShortDescriptionLimit)
	if got, cut := FitShortDescription(text); cut || got != text {
		t.Errorf("%q %v", got, cut)
	}
}
