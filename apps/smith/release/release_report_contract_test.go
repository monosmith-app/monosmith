package release

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestMarkdownNamesTheMirror(t *testing.T) {
	result := ReportResult{Item: Item{Project: "client", Next: "1.2.0"}, Status: "released", Mirror: "acme/client-php@v1.2.0"}
	if got := (Report{Results: []ReportResult{result}}).Markdown(); !strings.Contains(got, "Mirrored client to acme/client-php@v1.2.0") {
		t.Errorf("release:\n%s", got)
	}
	if got := (Report{DryRun: true, Results: []ReportResult{result}}).Markdown(); !strings.Contains(got, "Would mirror client to acme/client-php@v1.2.0") {
		t.Errorf("preview:\n%s", got)
	}
	if got := (Report{Results: []ReportResult{{Item: Item{Project: "web"}}}}).Markdown(); strings.Contains(got, "irror") {
		t.Errorf("no mirror, no line:\n%s", got)
	}
}

func TestMarkdownAndJSONNameTheImagesPushedWithTheirDigests(t *testing.T) {
	image := protocol.PublishedImage{Reference: "ghcr.io/acme/api", Tags: []string{"sha-0123456", "latest", "1.2.0"}, Digest: "sha256:abc"}
	result := ReportResult{Item: Item{Project: "api", Next: "1.2.0"}, Status: "released", Images: []protocol.PublishedImage{image}}
	got := (Report{Results: []ReportResult{result}}).Markdown()
	if !strings.Contains(got, "Images pushed for api:\n\n- `ghcr.io/acme/api@sha256:abc [tags: sha-0123456, latest, 1.2.0]`") {
		t.Errorf("markdown:\n%s", got)
	}
	raw, err := json.Marshal(result)
	if err != nil || !strings.Contains(string(raw), `"images":[{"reference":"ghcr.io/acme/api","tags":["sha-0123456","latest","1.2.0"],"digest":"sha256:abc"}]`) {
		t.Errorf("json %s %v", raw, err)
	}
	if got := (Report{Results: []ReportResult{{Item: Item{Project: "web"}}}}).Markdown(); strings.Contains(got, "Images") {
		t.Errorf("no images, no section:\n%s", got)
	}
	if line := ImageLine(protocol.PublishedImage{Reference: "r/i", Tags: []string{"1.0.0"}}); line != "r/i [tags: 1.0.0]" {
		t.Errorf("without a digest: %q", line)
	}
}
