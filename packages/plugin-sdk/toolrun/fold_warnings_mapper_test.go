package toolrun

import (
	"strings"
	"testing"
)

const engineBlocks = `npm warn EBADENGINE Unsupported engine {
npm warn EBADENGINE   package: 'a@1.0.0',
npm warn EBADENGINE   required: { node: '>=22' },
npm warn EBADENGINE   current: { node: 'v20.1.0', npm: '10.0.0' }
npm warn EBADENGINE }
npm warn EBADENGINE Unsupported engine {
npm warn EBADENGINE   package: 'b@2.0.0',
npm warn EBADENGINE   required: { node: '>=22' },
npm warn EBADENGINE }
npm warn deprecated left-pad@1.0.0: gone
`

func TestFoldWarningsFoldsEngineBlocksIntoOneLine(t *testing.T) {
	got := FoldWarnings(engineBlocks, "rerun with --verbose")
	if len(got) != 2 {
		t.Fatalf("%q", got)
	}
	if got[0] != "2 packages want a newer Node than this one (rerun with --verbose)" {
		t.Fatalf("%q", got[0])
	}
	if !strings.HasPrefix(got[1], "1 package is deprecated") {
		t.Fatalf("%q", got[1])
	}
}

func TestFoldWarningsSaysNothingWhenThereAreNone(t *testing.T) {
	if got := FoldWarnings("added 3 packages\n", "x"); len(got) != 0 {
		t.Fatalf("%q", got)
	}
}
