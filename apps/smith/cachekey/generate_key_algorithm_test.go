package cachekey

import (
	"bytes"
	"crypto/rand"
	"regexp"
	"strings"
	"testing"
)

func TestGenerateMakes64HexCharactersAndNeverTheSameTwice(t *testing.T) {
	a, err := Generate(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := Generate(rand.Reader)
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(a) || a == b {
		t.Fatalf("keys %q and %q", a, b)
	}
}

func TestGenerateReportsAShortRandomSource(t *testing.T) {
	if _, err := Generate(bytes.NewReader([]byte("short"))); err == nil || !strings.Contains(err.Error(), "random source") {
		t.Fatalf("got %v", err)
	}
}
