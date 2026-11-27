package dependencies

import (
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestADryRunListsEachDeclarationItWouldRewriteAndWritesNothing(t *testing.T) {
	chosen := []Update{{Plugin: "node", Name: "axios", Current: "1.6.0", Latest: "1.6.8", Level: "patch", Sites: []protocol.Declaration{
		{Project: "web", File: "apps/web/package.json", Name: "axios", Spec: "^1.6.0"},
		{Project: "api", File: "apps/api/package.json", Name: "axios", Spec: "~1.6.0"},
	}}}
	var b bytes.Buffer
	PrintDryRun(&b, chosen, true)
	got := b.String()
	for _, want := range []string{
		"Dry run: 2 declarations would change.",
		"  apps/web/package.json  axios  ^1.6.0 → ^1.6.8",
		"  apps/api/package.json  axios  ~1.6.0 → ~1.6.8",
		"Then smith would install the new versions",
		"Nothing was changed: run the same command without --dry-run",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	b.Reset()
	PrintDryRun(&b, chosen[:1], false)
	if strings.Contains(b.String(), "install") {
		t.Fatalf("--no-install says nothing of an install:\n%s", b.String())
	}
}
