package linuxpackages

import (
	"bytes"
	"testing"
)

func TestRPMIsAnRPMOfTheBinary(t *testing.T) {
	p := Package{Name: "monosmith", Version: "1.0.0", Arch: "amd64", Summary: "Runs a monorepo", License: "MIT", Binary: "smith"}
	raw, err := RPM(p, []byte("ELF binary"))
	if err != nil {
		t.Fatal(err)
	}
	// The lead: rpm's magic, then the name-version-release.
	if !bytes.HasPrefix(raw, []byte{0xed, 0xab, 0xee, 0xdb}) {
		t.Fatalf("no rpm lead: % x", raw[:4])
	}
	if !bytes.Contains(raw[10:76], []byte("monosmith-1.0.0-1")) {
		t.Errorf("lead name = %q", raw[10:76])
	}
	for _, want := range []string{"/usr/bin/", "smith", "x86_64", "MIT"} {
		if !bytes.Contains(raw, []byte(want)) {
			t.Errorf("rpm lacks %q", want)
		}
	}
}
