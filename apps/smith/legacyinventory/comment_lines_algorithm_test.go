package legacyinventory

import (
	"strings"
	"testing"
)

func TestWithoutCommentLinesDropsProseButKeepsCode(t *testing.T) {
	src := "import a from 'kept'\n// do not edit @mnci/eslint-config\n  /* a block\n  naming typescript-eslint */\nconst b = require('after') // trailing mention of rollup\n/** doc */ export const c = 1\n"
	got := WithoutCommentLines(src, ".mjs")
	for _, want := range []string{"kept", "after", "rollup", "export const c"} {
		if !strings.Contains(got, want) {
			t.Errorf("lost %q:\n%s", want, got)
		}
	}
	for _, gone := range []string{"@mnci/eslint-config", "typescript-eslint", "a block"} {
		if strings.Contains(got, gone) {
			t.Errorf("kept %q:\n%s", gone, got)
		}
	}
}

func TestWithoutCommentLinesKeepsAGlobThatLooksLikeAComment(t *testing.T) {
	// A string that starts a line with /* is not at the start of a comment
	// when the line begins with a quote.
	src := "const files = [\n  'packages/*',\n  'apps/*',\n]\nimport x from 'jest'\n"
	if got := WithoutCommentLines(src, ".ts"); got != src {
		t.Fatalf("changed:\n%s", got)
	}
}

func TestWithoutCommentLinesReadsHashCommentsInShellAndYaml(t *testing.T) {
	got := WithoutCommentLines("# run rollup\nnpx jest\n", ".yml")
	if strings.Contains(got, "rollup") || !strings.Contains(got, "jest") {
		t.Fatal(got)
	}
}
