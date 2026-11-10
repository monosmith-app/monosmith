package doctorview

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFileOfTakesTheFileTheFixWritesFirst(t *testing.T) {
	f := doctor.Finding{Finding: protocol.Finding{Problem: "smith.toml is old", Fix: &protocol.Fix{Files: []protocol.FileEdit{{Path: "apps/web/Dockerfile"}}}}}
	if got := FileOf(f); got != "apps/web/Dockerfile" {
		t.Fatalf("got %q", got)
	}
}

func TestFileOfReadsAPathInTheProblemOrTheDetail(t *testing.T) {
	cases := map[string]doctor.Finding{
		"packages/web/tsconfig.json": {Finding: protocol.Finding{Problem: "`packages/web/tsconfig.json` does not extend the base"}},
		"smith.toml":                 {Finding: protocol.Finding{Problem: "the cache settings", Detail: "set in smith.toml, line 4."}},
		".gitignore":                 {Finding: protocol.Finding{Problem: "dist is not in .gitignore"}},
	}
	for want, f := range cases {
		if got := FileOf(f); got != want {
			t.Errorf("%q: got %q", want, got)
		}
	}
}

func TestFileOfIsEmptyForAFindingAboutNoFile(t *testing.T) {
	for _, text := range []string{
		"golangci-lint is 2.14.0, CI uses 2.15.0",
		"install Node.js, e.g. with nvm",
		"libs/x is missing",
		"see https://example.com/docs.html",
	} {
		if got := FileOf(doctor.Finding{Finding: protocol.Finding{Problem: text}}); got != "" {
			t.Errorf("%q: got %q", text, got)
		}
	}
}
