package releasing

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

const publishNotices = `npm notice 📦  @monosmith/eslint-config@1.1.0
npm notice Tarball Contents  (25 files listed)
npm notice 1.1kB package.json
npm notice Tarball Details
npm notice name: @monosmith/eslint-config
npm notice Publishing to https://registry.npmjs.org/ with tag latest and public access
`

func TestAnE404ProbeIsAnAnswerAndNotAnError(t *testing.T) {
	for stderr, want := range map[string]bool{
		"npm error code E404\nnpm error 404 No match found for version 1.1.0": true,
		"npm error 404 Not Found - GET https://registry.npmjs.org/x":          true,
		"npm error code ECONNRESET":                                           false,
		"":                                                                    false,
	} {
		if got := notFound(stderr); got != want {
			t.Errorf("%q: %v", stderr, got)
		}
	}
}

func TestAPublishKeepsOnlyWhereItPublishedUnlessVerbose(t *testing.T) {
	if got := keepPublishLines(publishNotices); got != "npm notice Publishing to https://registry.npmjs.org/ with tag latest and public access\n" {
		t.Errorf("%q", got)
	}
	if keepPublishLines("npm notice 📦  x\n") != "" {
		t.Error("a listing alone leaves nothing")
	}
	t.Setenv(toolrun.VerboseEnv, "")
	if verbose() {
		t.Error("not verbose by default")
	}
	t.Setenv(toolrun.VerboseEnv, "1")
	if !verbose() {
		t.Error("SMITH_VERBOSE shows the listing")
	}
}

func TestTheFailureCarriesTheLastLinesOfNpm(t *testing.T) {
	text := strings.Repeat("npm notice x\n", 20) + "npm error code E403\nnpm error you cannot publish over the previously published versions\n"
	got := lastLines(text, 2)
	if got != "npm error code E403\nnpm error you cannot publish over the previously published versions" {
		t.Errorf("%q", got)
	}
}
