package pluginarchive

import (
	"strings"
	"testing"
)

func TestChecksumsRoundTripSorted(t *testing.T) {
	sums := map[string]string{"b": sum("b"), "a/x": sum("x")}
	raw := FormatChecksums(sums)
	if !strings.HasPrefix(string(raw), sum("x")+"  a/x\n") {
		t.Fatalf("not sorted or not in sha256sum format: %s", raw)
	}
	back, err := ParseChecksums(raw)
	if err != nil || len(back) != 2 || back["b"] != sum("b") {
		t.Fatalf("got %v, %v", back, err)
	}
}

func TestParseChecksumsRefusesBadLines(t *testing.T) {
	for name, raw := range map[string]string{
		"short digest": "abc  a\n",
		"no name":      sum("a") + "\n",
		"traversal":    sum("a") + "  ../a\n",
		"duplicate":    sum("a") + "  a\n" + sum("b") + "  a\n",
	} {
		if _, err := ParseChecksums([]byte(raw)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
