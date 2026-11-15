package why

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintShowsLabelledLines(t *testing.T) {
	var b bytes.Buffer
	Print(&b, Explanation{
		Path: "a/b.json", Known: true, Kind: KindDerived, Owner: OwnerSmith, Summary: "smith owns it.",
		Rule: "renderer r v1", Parts: []Part{{Renderer: "r", Version: 1, Mode: "keys", Owner: "smith", State: "current"}},
		Change: []string{"one", "two"},
	})
	want := []string{"a/b.json\n", "  Summary  smith owns it.", "  Owner    smith", "  Part     r v1 (keys, owner smith, current)", "  Change   one", "           two"}
	for _, w := range want {
		if !strings.Contains(b.String(), w) {
			t.Errorf("output lacks %q:\n%s", w, b.String())
		}
	}
}
