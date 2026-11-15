package taskrun

import (
	"strings"
	"sync"
	"testing"
)

func TestPrefixWriterHoldsLinesUntilReleased(t *testing.T) {
	var out strings.Builder
	var mu sync.Mutex
	w := newPrefixWriter("web:build │ ", &out, &mu)
	w.hold = true
	_, _ = w.Write([]byte("one\ntwo"))
	w.Flush()
	if out.Len() != 0 {
		t.Fatalf("held output reached the screen: %q", out.String())
	}
	if w.log.String() != "one\ntwo" {
		t.Fatalf("the cache log must keep everything: %q", w.log.String())
	}
	w.Release()
	if out.String() != "web:build │ one\nweb:build │ two\n" {
		t.Fatalf("released output: %q", out.String())
	}
}
