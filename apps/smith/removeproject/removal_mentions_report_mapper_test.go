package removeproject

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteMentionsBoundsTheListAndSaysSmithKeepsOut(t *testing.T) {
	r := Removal{Project: "mvd-tui", Root: "apps/mvd-tui"}
	for i := 0; i < 25; i++ {
		r.Mentions = append(r.Mentions, Mention{File: "docs/f" + string(rune('a'+i)) + ".md", Line: 1, Kind: MentionOther})
	}
	var b bytes.Buffer
	WriteMentions(&b, r)
	out := b.String()
	if !strings.Contains(out, "Still mentions mvd-tui") || !strings.Contains(out, "and 5 more") || !strings.Contains(out, "does not change") {
		t.Fatal(out)
	}
	if strings.Count(out, "docs/f") != 20 {
		t.Fatal(out)
	}
	b.Reset()
	WriteMentions(&b, Removal{Project: "x"})
	if b.Len() != 0 {
		t.Fatal(b.String())
	}
}
