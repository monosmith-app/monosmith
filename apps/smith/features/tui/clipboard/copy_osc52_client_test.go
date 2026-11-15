package clipboard

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestCopyWritesOneOSC52Sequence(t *testing.T) {
	var out bytes.Buffer
	if err := Copy(&out, "smith doctor --fix"); err != nil {
		t.Fatal(err)
	}
	want := "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte("smith doctor --fix")) + "\x07"
	if out.String() != want {
		t.Fatalf("got %q, want %q", out.String(), want)
	}
}

func TestCopyRefusesEmptyTextAndNoWriter(t *testing.T) {
	if err := Copy(&bytes.Buffer{}, ""); err != ErrNothing {
		t.Fatalf("empty: %v", err)
	}
	if err := Copy(nil, "x"); err != ErrNothing {
		t.Fatalf("nil writer: %v", err)
	}
}

func TestCopyCutsTextTooLongForTheTerminal(t *testing.T) {
	var out bytes.Buffer
	if err := Copy(&out, strings.Repeat("a", maxBytes*2)); err != nil {
		t.Fatal(err)
	}
	if out.Len() > maxBytes*2 {
		t.Fatalf("sequence of %d bytes", out.Len())
	}
}
